package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/version"
)

const maxTransitionRequestBytes = 1 << 20

func exactMethod(method string, handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			writeAPIError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
			return
		}
		handler(w, r)
	})
}

func routeValue(r *http.Request, name string) string {
	if value := r.PathValue(name); value != "" {
		return value
	}
	// Legacy path routing is performed by gorilla/mux outside the VM process,
	// so Go's Request.PathValue is not populated. The dynamic value is always
	// the final segment for LOCALITY's legacy routes.
	return path.Base(strings.TrimRight(r.URL.Path, "/"))
}

func (vm *VM) apiHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /status", vm.handleStatus)
	mux.HandleFunc("GET /genesis", vm.handleGenesis)
	mux.HandleFunc("GET /state", vm.handleState)
	mux.HandleFunc("GET /capacity", vm.handleCapacity)
	mux.HandleFunc("GET /currentness", vm.handleCurrentness)
	mux.HandleFunc("POST /drafts", vm.handleDraftTransition)
	mux.HandleFunc("POST /transitions", vm.handleIssueTransition)
	mux.HandleFunc("GET /transitions/{transitionID}", vm.handleGetTransition)
	mux.HandleFunc("GET /receipts/{transitionID}", vm.handleGetReceipt)
	mux.HandleFunc("GET /blocks/{blockID}", vm.handleGetBlock)
	return mux
}

func (vm *VM) handleDraftTransition(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxTransitionRequestBytes))
	if err != nil {
		writeAPIError(w, http.StatusRequestEntityTooLarge, err)
		return
	}
	request := new(TransitionDraftRequest)
	if err := decodeStrict(body, request); err != nil {
		writeAPIError(w, http.StatusBadRequest, err)
		return
	}
	unsigned, err := vm.draftTransition(request)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"unsigned":  unsigned,
		"next_step": "sign this exact unsigned object with presence-avalanche-vm --sign-unsigned, then POST the resulting canonical transition to /transitions",
		"effect":    "DRAFT_ONLY_NO_STATE_CHANGE",
	})
}

func (vm *VM) handleStatus(w http.ResponseWriter, _ *http.Request) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		writeAPIError(w, http.StatusServiceUnavailable, errNotInitialized)
		return
	}
	block, err := vm.getBlockLocked(vm.lastAcceptedID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	capacity, err := block.postState.capacitySnapshot()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"healthy":                   true,
		"vm_version":                vmVersion,
		"implementation":            implementationID,
		"protocol":                  vmIDDomain,
		"protocol_contract_sha256": protocolContractSHA256,
		"rpcchainvm_protocol":       version.RPCChainVMProtocol,
		"operation_contract_sha256": fmt.Sprintf("%x", sha256.Sum256(protocolContractJSON)),
		"avalanchego_target":        "v1.15.0",
		"avalanchego_profile":       avalancheGoProfile,
		"genesis_id":                vm.genesis.Domain.GenesisID,
		"locality_id":               vm.genesis.Locality.ID,
		"last_accepted_block":       vm.lastAcceptedID.String(),
		"height":                    block.height,
		"revision":                  block.postState.Revision,
		"state_commitment":          block.postState.StateCommitment,
		"active_locus_id":           block.postState.ActiveLocusID,
		"body":                      block.postState.Body,
		"capacity":                  capacity,
		"last_pulse_id":             block.postState.LastPulseID,
		"successor_boundary":        block.postState.SuccessorBoundary,
		"pending_transitions":       len(vm.pendingOrder),
	})
}

func (vm *VM) handleCapacity(w http.ResponseWriter, _ *http.Request) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		writeAPIError(w, http.StatusServiceUnavailable, errNotInitialized)
		return
	}
	state, err := vm.acceptedStateLocked()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	snapshot, err := state.capacitySnapshot()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"capacity": snapshot, "posture": state.Body.Posture,
		"state_commitment": state.StateCommitment,
		"effect":           "SIGNED_LOCAL_CAPACITY_ACCOUNT_NOT_EXTERNAL_RESOURCE_PROOF",
	})
}

func (vm *VM) handleCurrentness(w http.ResponseWriter, r *http.Request) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		writeAPIError(w, http.StatusServiceUnavailable, errNotInitialized)
		return
	}
	state, err := vm.acceptedStateLocked()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	snapshot, err := state.capacitySnapshot()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	response := map[string]any{
		"body": state.Body, "capacity": snapshot, "last_pulse_id": state.LastPulseID,
		"state_commitment": state.StateCommitment, "pending_transitions": len(vm.pendingOrder),
		"effect": "LOCAL_CURRENTNESS_ONLY_NOT_EXTERNAL_LIVENESS_PROOF",
	}
	if pulse, exists := state.Pulses[state.LastPulseID]; exists {
		response["last_pulse"] = pulse
	}
	observedAtText := r.URL.Query().Get("observed_at")
	carrierSetSHA256 := r.URL.Query().Get("carrier_set_sha256")
	if observedAtText != "" || carrierSetSHA256 != "" {
		observedAt, err := strconv.ParseInt(observedAtText, 10, 64)
		if err != nil || observedAt < state.LastObservedAt || observedAt > time.Now().UTC().Add(maxClockSkew).Unix() {
			writeAPIError(w, http.StatusBadRequest, errors.New("observed_at must be an integer within the accepted local timeline and future-skew ceiling"))
			return
		}
		if err := requireDigest("carrier_set_sha256", carrierSetSHA256); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		commitment, err := state.currentnessCommitment(observedAt, carrierSetSHA256)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, err)
			return
		}
		response["next_pulse"] = map[string]any{
			"observed_at": observedAt, "carrier_set_sha256": carrierSetSHA256,
			"currentness_commitment_sha256": commitment,
			"warning":                       "valid only while this exact state commitment remains current and no pending transition precedes it",
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (vm *VM) handleGenesis(w http.ResponseWriter, _ *http.Request) {
	vm.lock.RLock()
	defer vm.lock.RUnlock()
	if !vm.initialized {
		writeAPIError(w, http.StatusServiceUnavailable, errNotInitialized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"genesis":          vm.genesis,
		"genesis_sha256":   hex.EncodeToString(vm.genesisBlock.id[:]),
		"genesis_block_id": vm.genesisBlock.id.String(),
	})
}

func (vm *VM) handleState(w http.ResponseWriter, _ *http.Request) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		writeAPIError(w, http.StatusServiceUnavailable, errNotInitialized)
		return
	}
	state, err := vm.acceptedStateLocked()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (vm *VM) handleIssueTransition(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxTransitionRequestBytes))
	if err != nil {
		writeAPIError(w, http.StatusRequestEntityTooLarge, err)
		return
	}
	transitionID, err := vm.issueTransition(body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"transition_id": transitionID,
		"status":        "PENDING_CONSENSUS",
		"effect":        "NO_ACCEPTANCE_UNTIL_BLOCK_ACCEPTED",
	})
}

func (vm *VM) handleGetTransition(w http.ResponseWriter, r *http.Request) {
	transitionID := routeValue(r, "transitionID")
	if err := requireDigest("transition_id", transitionID); err != nil {
		writeAPIError(w, http.StatusBadRequest, err)
		return
	}
	vm.lock.RLock()
	defer vm.lock.RUnlock()
	if pending, exists := vm.pending[transitionID]; exists {
		writeJSON(w, http.StatusOK, map[string]any{"status": "PENDING_CONSENSUS", "transition": pending})
		return
	}
	transitionBytes, err := vm.db.Get(transitionKey(transitionID))
	if errors.Is(err, database.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	transition := new(Transition)
	if err := decodeCanonical(transitionBytes, transition); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ACCEPTED", "transition": transition})
}

func (vm *VM) handleGetReceipt(w http.ResponseWriter, r *http.Request) {
	transitionID := routeValue(r, "transitionID")
	if err := requireDigest("transition_id", transitionID); err != nil {
		writeAPIError(w, http.StatusBadRequest, err)
		return
	}
	vm.lock.RLock()
	defer vm.lock.RUnlock()
	receiptBytes, err := vm.db.Get(receiptKey(transitionID))
	if errors.Is(err, database.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	receipt := new(Receipt)
	if err := decodeStrict(receiptBytes, receipt); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}

func (vm *VM) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	blockID, err := ids.FromString(strings.TrimSpace(routeValue(r, "blockID")))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, fmt.Errorf("block_id: %w", err))
		return
	}
	vm.lock.Lock()
	defer vm.lock.Unlock()
	block, err := vm.getBlockLocked(blockID)
	if errors.Is(err, database.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err)
		return
	}
	response := map[string]any{
		"block_id":  block.id.String(),
		"height":    block.height,
		"timestamp": block.timestamp,
		"status":    block.status.String(),
	}
	if block.postState != nil {
		response["state_commitment"] = block.postState.StateCommitment
	}
	if !block.genesis {
		response["block"] = block.wire
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAPIError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

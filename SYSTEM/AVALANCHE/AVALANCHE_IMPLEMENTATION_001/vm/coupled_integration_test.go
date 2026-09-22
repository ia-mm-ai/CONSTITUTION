//go:build integration

package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ava-labs/avalanchego/snow"
)

// Consensus is driven explicitly here, not simulated by a FIELD client stub.
// This qualifies the real VM/HTTP/FIELD boundary, not a validator network.
func TestCoupledFIELDLifecycle(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Fatal("Node.js is required for the coupled FIELD integration test")
	}
	_, genesis, _ := testGenesis(t)
	hostPublic, hostPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	participantPublic, participantPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	genesis.Locality.Authority.PublicKey = hex.EncodeToString(hostPublic)
	genesisBytes, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	db := memdb.New()
	initialize := func() (*VM, error) {
		vm := new(VM)
		if err := vm.Initialize(context.Background(), nil, db, genesisBytes, nil, nil, nil, nil); err != nil {
			return nil, err
		}
		return vm, vm.SetState(context.Background(), snow.NormalOp)
	}
	vm, err := initialize()
	if err != nil {
		t.Fatal(err)
	}
	controlBytes := make([]byte, 32)
	if _, err := rand.Read(controlBytes); err != nil {
		t.Fatal(err)
	}
	controlToken := hex.EncodeToString(controlBytes)
	journalBytes := make([]byte, 16)
	if _, err := rand.Read(journalBytes); err != nil {
		t.Fatal(err)
	}
	journalName := ".coupled-journal-" + hex.EncodeToString(journalBytes)
	journalDirectory := filepath.Join("..", "integration", journalName)
	if err := os.Mkdir(journalDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(journalDirectory); err != nil {
			t.Errorf("clean integration journals: %v", err)
		}
	})
	var lock sync.Mutex
	var accepted, restarts int
	operations := make(map[string]int)
	const prefix = "/ext/bc/LOCAL-COUPLED"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		if !strings.HasPrefix(r.URL.Path, "/__test/") {
			http.StripPrefix(prefix, vm.apiHandler()).ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodPost || r.Header.Get("X-Test-Control") != controlToken {
			http.Error(w, "test controller authorization required", http.StatusForbidden)
			return
		}
		switch r.URL.Path {
		case "/__test/accept":
			block, err := vm.BuildBlock(r.Context())
			if err == nil {
				err = block.Verify(r.Context())
			}
			if err == nil {
				err = block.Accept(r.Context())
			}
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, err)
				return
			}
			accepted++
			state, err := vm.acceptedStateLocked()
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, err)
				return
			}
			operations[state.Events[state.LastTransitionID].Operation]++
			writeJSON(w, http.StatusOK, map[string]any{
				"accepted_block_id": block.ID().String(), "height": block.Height(),
			})
		case "/__test/restart":
			before, err := vm.LastAccepted(r.Context())
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, err)
				return
			}
			reloaded, err := initialize()
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, err)
				return
			}
			vm = reloaded
			restarts++
			after, err := vm.LastAccepted(r.Context())
			if err != nil {
				writeAPIError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"previous_last_accepted": before.String(), "last_accepted": after.String(),
				"boundary": "VM_REINITIALIZATION_FROM_SAME_DATABASE_NOT_PROCESS_RESTART",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Cleanup(func() {
		lock.Lock()
		defer lock.Unlock()
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown VM: %v", err)
		}
	})
	input, err := json.Marshal(map[string]any{
		"endpoint": server.URL + prefix, "control_endpoint": server.URL,
		"control_token": controlToken, "host_id": genesis.Locality.ID,
		"go_version":      runtime.Version(),
		"journal_name":    journalName,
		"host_public_key": hex.EncodeToString(hostPublic), "host_seed": hex.EncodeToString(hostPrivate.Seed()),
		"participant_id":         operationalActorID(participantPublic),
		"participant_public_key": hex.EncodeToString(participantPublic),
		"participant_seed":       hex.EncodeToString(participantPrivate.Seed()),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", filepath.Join("..", "integration", "driver.mjs"))
	command.Stdin = bytes.NewReader(input)
	output, runErr := command.CombinedOutput()
	clear(input)
	clear(hostPrivate)
	clear(participantPrivate)
	if runErr != nil {
		t.Fatalf("coupled FIELD driver: %v\n%s", runErr, output)
	}
	var result struct {
		AcceptedTransitions int    `json:"accepted_transitions"`
		FinalRevision       uint64 `json:"final_revision"`
		FinalPosture        string `json:"final_posture"`
		Result              string `json:"result"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("driver did not return qualification JSON: %v\n%s", err, output)
	}
	lock.Lock()
	defer lock.Unlock()
	state := acceptedState(t, vm)
	if result.Result != "PASS" || result.AcceptedTransitions != accepted || result.FinalRevision != state.Revision {
		t.Fatalf("FIELD/VM result mismatch: result=%+v accepted=%d revision=%d", result, accepted, state.Revision)
	}
	if accepted < 29 || restarts != 2 || state.ActiveLocusID != "" || state.Body.PresenceCount != 0 {
		t.Fatalf("incomplete lifecycle: accepted=%d restarts=%d active=%q presence=%d", accepted, restarts, state.ActiveLocusID, state.Body.PresenceCount)
	}
	for _, operation := range []string{
		opBound, opDeclareCapacity, opPresentForm, opGateDisposition, opEnter,
		opObserveCrossing, opMatterDisposition, opRecordEmergence, opCorrect,
		opCheckpointDeparture, opExit, opClose, opReenter, opPulse,
	} {
		if operations[operation] == 0 {
			t.Errorf("operation %s never accepted by the real VM", operation)
		}
	}
	if err := state.validateLoaded(); err != nil {
		t.Fatalf("final state integrity: %v", err)
	}
	t.Logf("real VM + FIELD: %d accepted transitions, %d database reloads, revision %d, posture %s; validator network NOT RUN",
		accepted, restarts, result.FinalRevision, result.FinalPosture)
	if evidencePath := os.Getenv("COUPLED_EVIDENCE_PATH"); evidencePath != "" {
		var formatted bytes.Buffer
		if err := json.Indent(&formatted, output, "", "  "); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(evidencePath, formatted.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

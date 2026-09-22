package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/version"

	smblock "github.com/ava-labs/avalanchego/snow/engine/snowman/block"
)

var (
	_ smblock.ChainVM = (*VM)(nil)

	errNotInitialized       = errors.New("PRESENCE AVALANCHE VM is not initialized")
	errNoPendingTransitions = errors.New("no valid pending PRESENCE transitions")
	errInvalidPreference    = errors.New("preferred block is unavailable or unverified")
)

var (
	genesisBytesKey = []byte("locality/meta/genesis-bytes")
	genesisIDKey    = []byte("locality/meta/genesis-id")
	lastAcceptedKey = []byte("locality/meta/last-accepted")
	currentStateKey = []byte("locality/state/current")
)

type VM struct {
	lock sync.RWMutex

	db             database.Database
	chainContext   *snow.Context
	genesis        *Genesis
	genesisBytes   []byte
	genesisBlock   *localityBlock
	blocks         map[ids.ID]*localityBlock
	lastAcceptedID ids.ID
	preferredID    ids.ID
	chainState     snow.State
	initialized    bool

	pending      map[string]*Transition
	pendingOrder []string
	notify       chan struct{}
}

func (vm *VM) Initialize(
	_ context.Context,
	chainContext *snow.Context,
	db database.Database,
	genesisBytes []byte,
	_ []byte,
	_ []byte,
	_ []*common.Fx,
	_ common.AppSender,
) error {
	genesis, err := parseGenesis(genesisBytes)
	if err != nil {
		return fmt.Errorf("parse PRESENCE runtime genesis: %w", err)
	}
	initialState, err := initialRuntimeState(genesis)
	if err != nil {
		return fmt.Errorf("construct initial state: %w", err)
	}
	genesisBlock := newGenesisBlock(vm, genesisBytes, initialState)

	storedGenesis, err := db.Get(genesisBytesKey)
	switch {
	case errors.Is(err, database.ErrNotFound):
		stateBytes, err := json.Marshal(initialState)
		if err != nil {
			return err
		}
		vdb := versiondb.New(db)
		for _, write := range []struct {
			key   []byte
			value []byte
		}{
			{genesisBytesKey, genesisBytes},
			{genesisIDKey, genesisBlock.id[:]},
			{lastAcceptedKey, genesisBlock.id[:]},
			{currentStateKey, stateBytes},
			{blockKey(genesisBlock.id), genesisBytes},
			{stateByBlockKey(genesisBlock.id), stateBytes},
			{heightKey(0), genesisBlock.id[:]},
		} {
			if err := vdb.Put(write.key, write.value); err != nil {
				return fmt.Errorf("stage genesis state: %w", err)
			}
		}
		if err := vdb.Commit(); err != nil {
			return fmt.Errorf("commit genesis state: %w", err)
		}
	case err != nil:
		return fmt.Errorf("read stored genesis: %w", err)
	case !bytes.Equal(storedGenesis, genesisBytes):
		return errors.New("stored immutable genesis differs from supplied genesis")
	default:
		storedGenesisID, err := db.Get(genesisIDKey)
		if err != nil {
			return fmt.Errorf("read stored genesis ID: %w", err)
		}
		if !bytes.Equal(storedGenesisID, genesisBlock.id[:]) {
			return errors.New("stored genesis ID does not match supplied genesis")
		}
	}

	lastAcceptedBytes, err := db.Get(lastAcceptedKey)
	if err != nil {
		return fmt.Errorf("read last accepted block: %w", err)
	}
	lastAcceptedID, err := ids.ToID(lastAcceptedBytes)
	if err != nil {
		return fmt.Errorf("decode last accepted block ID: %w", err)
	}

	vm.lock.Lock()
	defer vm.lock.Unlock()
	vm.db = db
	vm.chainContext = chainContext
	vm.genesis = genesis
	vm.genesisBytes = bytes.Clone(genesisBytes)
	vm.genesisBlock = genesisBlock
	vm.blocks = map[ids.ID]*localityBlock{genesisBlock.id: genesisBlock}
	vm.lastAcceptedID = lastAcceptedID
	vm.preferredID = lastAcceptedID
	vm.chainState = snow.Initializing
	vm.pending = make(map[string]*Transition)
	vm.pendingOrder = nil
	vm.notify = make(chan struct{}, 1)
	vm.initialized = true

	if lastAcceptedID != genesisBlock.id {
		lastAccepted, err := vm.loadAcceptedBlockLocked(lastAcceptedID)
		if err != nil {
			vm.initialized = false
			return fmt.Errorf("load last accepted block: %w", err)
		}
		vm.blocks[lastAcceptedID] = lastAccepted
	}
	return nil
}

func (vm *VM) SetState(_ context.Context, state snow.State) error {
	if state != snow.Initializing && state != snow.StateSyncing && state != snow.Bootstrapping && state != snow.NormalOp {
		return snow.ErrUnknownState
	}
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return errNotInitialized
	}
	vm.chainState = state
	return nil
}

func (vm *VM) Shutdown(context.Context) error {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return nil
	}
	vm.initialized = false
	return vm.db.Close()
}

func (*VM) Version(context.Context) (string, error) { return vmVersion, nil }

func (vm *VM) CreateHandlers(context.Context) (map[string]http.Handler, error) {
	return map[string]http.Handler{
		"/status":                     exactMethod(http.MethodGet, vm.handleStatus),
		"/genesis":                    exactMethod(http.MethodGet, vm.handleGenesis),
		"/state":                      exactMethod(http.MethodGet, vm.handleState),
		"/capacity":                   exactMethod(http.MethodGet, vm.handleCapacity),
		"/currentness":                exactMethod(http.MethodGet, vm.handleCurrentness),
		"/drafts":                     exactMethod(http.MethodPost, vm.handleDraftTransition),
		"/transitions":                exactMethod(http.MethodPost, vm.handleIssueTransition),
		"/transitions/{transitionID}": exactMethod(http.MethodGet, vm.handleGetTransition),
		"/receipts/{transitionID}":    exactMethod(http.MethodGet, vm.handleGetReceipt),
		"/blocks/{blockID}":           exactMethod(http.MethodGet, vm.handleGetBlock),
	}, nil
}

func (vm *VM) NewHTTPHandler(context.Context) (http.Handler, error) {
	return vm.apiHandler(), nil
}

func (vm *VM) HealthCheck(context.Context) (interface{}, error) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return nil, errNotInitialized
	}
	lastAccepted, err := vm.getBlockLocked(vm.lastAcceptedID)
	if err != nil {
		return nil, err
	}
	capacity, err := lastAccepted.postState.capacitySnapshot()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"healthy":             true,
		"vmVersion":           vmVersion,
		"rpcChainVMProtocol":  version.RPCChainVMProtocol,
		"lastAcceptedBlockID": vm.lastAcceptedID.String(),
		"height":              lastAccepted.height,
		"revision":            lastAccepted.postState.Revision,
		"stateCommitment":     lastAccepted.postState.StateCommitment,
		"activeLocusID":       lastAccepted.postState.ActiveLocusID,
		"body":                lastAccepted.postState.Body,
		"capacity":            capacity,
		"lastPulseID":         lastAccepted.postState.LastPulseID,
		"successorBoundary":   lastAccepted.postState.SuccessorBoundary,
	}, nil
}

func (vm *VM) WaitForEvent(ctx context.Context) (common.Message, error) {
	vm.lock.RLock()
	if !vm.initialized {
		vm.lock.RUnlock()
		return 0, errNotInitialized
	}
	if len(vm.pendingOrder) > 0 {
		vm.lock.RUnlock()
		return common.PendingTxs, nil
	}
	notify := vm.notify
	vm.lock.RUnlock()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-notify:
		return common.PendingTxs, nil
	}
}

func (vm *VM) BuildBlock(context.Context) (snowman.Block, error) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return nil, errNotInitialized
	}
	if vm.chainState != snow.NormalOp {
		return nil, fmt.Errorf("cannot build while chain state is %s", vm.chainState)
	}
	parent, err := vm.getBlockLocked(vm.preferredID)
	if err != nil || parent.postState == nil || parent.status == choices.Rejected {
		return nil, errInvalidPreference
	}
	workingState, err := parent.postState.clone()
	if err != nil {
		return nil, err
	}
	selected := make([]*Transition, 0, maxTransitionsPerBlock)
	remainingOrder := make([]string, 0, len(vm.pendingOrder))
	for _, transitionID := range vm.pendingOrder {
		transition, exists := vm.pending[transitionID]
		if !exists {
			continue
		}
		if len(selected) >= maxTransitionsPerBlock {
			remainingOrder = append(remainingOrder, transitionID)
			continue
		}
		trial, err := workingState.clone()
		if err != nil {
			return nil, err
		}
		if _, err := trial.apply(vm.genesis, transition, transitionID); err != nil {
			delete(vm.pending, transitionID)
			continue
		}
		workingState = trial
		selected = append(selected, transition)
		delete(vm.pending, transitionID)
	}
	vm.pendingOrder = remainingOrder
	if len(selected) == 0 {
		vm.drainNotifyLocked()
		return nil, errNoPendingTransitions
	}
	timestamp := time.Now().UTC().Truncate(time.Second)
	if timestamp.Before(parent.timestamp) {
		timestamp = parent.timestamp
	}
	wire := &WireBlock{
		Schema:      blockSchema,
		ParentID:    parent.id.String(),
		Timestamp:   timestamp.Unix(),
		Height:      parent.height + 1,
		Transitions: selected,
	}
	block, err := newWireBlock(vm, wire, true)
	if err != nil {
		return nil, err
	}
	vm.blocks[block.id] = block
	if len(vm.pendingOrder) > 0 {
		vm.signalPendingLocked()
	} else {
		vm.drainNotifyLocked()
	}
	return block, nil
}

func (vm *VM) ParseBlock(_ context.Context, blockBytes []byte) (snowman.Block, error) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return nil, errNotInitialized
	}
	if bytes.Equal(blockBytes, vm.genesisBytes) {
		return vm.genesisBlock, nil
	}
	block, err := blockFromBytes(vm, blockBytes, choices.Processing)
	if err != nil {
		return nil, err
	}
	if existing, exists := vm.blocks[block.id]; exists {
		return existing, nil
	}
	if _, err := vm.db.Get(blockKey(block.id)); err == nil {
		accepted, err := vm.loadAcceptedBlockLocked(block.id)
		if err != nil {
			return nil, err
		}
		vm.blocks[block.id] = accepted
		return accepted, nil
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	vm.blocks[block.id] = block
	return block, nil
}

func (vm *VM) GetBlock(_ context.Context, blockID ids.ID) (snowman.Block, error) {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return nil, errNotInitialized
	}
	return vm.getBlockLocked(blockID)
}

func (vm *VM) SetPreference(_ context.Context, blockID ids.ID) error {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return errNotInitialized
	}
	block, err := vm.getBlockLocked(blockID)
	if err != nil || block.status == choices.Rejected || block.postState == nil {
		return errInvalidPreference
	}
	vm.preferredID = blockID
	return nil
}

func (vm *VM) LastAccepted(context.Context) (ids.ID, error) {
	vm.lock.RLock()
	defer vm.lock.RUnlock()
	if !vm.initialized {
		return ids.Empty, errNotInitialized
	}
	return vm.lastAcceptedID, nil
}

func (vm *VM) GetBlockIDAtHeight(_ context.Context, height uint64) (ids.ID, error) {
	vm.lock.RLock()
	defer vm.lock.RUnlock()
	if !vm.initialized {
		return ids.Empty, errNotInitialized
	}
	blockIDBytes, err := vm.db.Get(heightKey(height))
	if err != nil {
		return ids.Empty, err
	}
	return ids.ToID(blockIDBytes)
}

func (vm *VM) issueTransition(transitionBytes []byte) (string, error) {
	transition, transitionID, err := parseTransition(transitionBytes)
	if err != nil {
		return "", err
	}
	if transition.Unsigned.ObservedAt > time.Now().UTC().Add(maxClockSkew).Unix() {
		return "", errors.New("transition observed_at exceeds the local future-skew ceiling")
	}
	transitionIDHex := hex.EncodeToString(transitionID[:])
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return "", errNotInitialized
	}
	if _, exists := vm.pending[transitionIDHex]; exists {
		return transitionIDHex, nil
	}
	preferred, err := vm.getBlockLocked(vm.preferredID)
	if err != nil || preferred.postState == nil {
		return "", errInvalidPreference
	}
	if _, exists := preferred.postState.Events[transitionIDHex]; exists {
		return "", errors.New("transition is already accepted")
	}
	preview, err := preferred.postState.clone()
	if err != nil {
		return "", err
	}
	for _, pendingID := range vm.pendingOrder {
		pending, exists := vm.pending[pendingID]
		if !exists {
			continue
		}
		if _, err := preview.apply(vm.genesis, pending, pendingID); err != nil {
			return "", fmt.Errorf("pending transition %s became invalid: %w", pendingID, err)
		}
	}
	if _, err := preview.apply(vm.genesis, transition, transitionIDHex); err != nil {
		return "", err
	}
	vm.pending[transitionIDHex] = transition
	vm.pendingOrder = append(vm.pendingOrder, transitionIDHex)
	vm.signalPendingLocked()
	return transitionIDHex, nil
}

func (vm *VM) verifyBlock(block *localityBlock) error {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return errNotInitialized
	}
	if block.postState != nil {
		return nil
	}
	if time.Until(block.timestamp) > maxClockSkew {
		return errors.New("block timestamp exceeds maximum clock skew")
	}
	parent, err := vm.getBlockLocked(block.parentID)
	if err != nil {
		return fmt.Errorf("load parent: %w", err)
	}
	if parent.status == choices.Rejected || parent.postState == nil {
		return errors.New("parent is rejected or unverified")
	}
	if block.height != parent.height+1 {
		return errors.New("block height does not follow parent")
	}
	if block.timestamp.Before(parent.timestamp) {
		return errors.New("block timestamp precedes parent")
	}
	state, err := parent.postState.clone()
	if err != nil {
		return err
	}
	receipts := make([]*Receipt, 0, len(block.wire.Transitions))
	for i, transition := range block.wire.Transitions {
		if transition.Unsigned.ObservedAt > block.timestamp.Add(maxClockSkew).Unix() {
			return fmt.Errorf("transition %d observed_at exceeds block future-skew ceiling", i)
		}
		transitionID, err := transition.ID()
		if err != nil {
			return err
		}
		transitionIDHex := hex.EncodeToString(transitionID[:])
		receipt, err := state.apply(vm.genesis, transition, transitionIDHex)
		if err != nil {
			return fmt.Errorf("transition %d (%s): %w", i, transitionIDHex, err)
		}
		receipts = append(receipts, receipt)
	}
	block.postState = state
	block.receipts = receipts
	block.status = choices.Processing
	vm.blocks[block.id] = block
	return nil
}

func (vm *VM) acceptBlock(block *localityBlock) error {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return errNotInitialized
	}
	if block.postState == nil {
		return errors.New("cannot accept an unverified block")
	}
	if block.parentID != vm.lastAcceptedID {
		return errors.New("accepted block does not extend last accepted block")
	}
	stateBytes, err := json.Marshal(block.postState)
	if err != nil {
		return err
	}
	vdb := versiondb.New(vm.db)
	for _, write := range []struct {
		key   []byte
		value []byte
	}{
		{blockKey(block.id), block.bytes},
		{stateByBlockKey(block.id), stateBytes},
		{heightKey(block.height), block.id[:]},
		{lastAcceptedKey, block.id[:]},
		{currentStateKey, stateBytes},
	} {
		if err := vdb.Put(write.key, write.value); err != nil {
			return err
		}
	}
	for i, transition := range block.wire.Transitions {
		transitionBytes, err := transition.Bytes()
		if err != nil {
			return err
		}
		transitionID, err := transition.ID()
		if err != nil {
			return err
		}
		transitionIDHex := hex.EncodeToString(transitionID[:])
		receiptBytes, err := json.Marshal(block.receipts[i])
		if err != nil {
			return err
		}
		if err := vdb.Put(transitionKey(transitionIDHex), transitionBytes); err != nil {
			return err
		}
		if err := vdb.Put(receiptKey(transitionIDHex), receiptBytes); err != nil {
			return err
		}
		delete(vm.pending, transitionIDHex)
	}
	if err := vdb.Commit(); err != nil {
		return fmt.Errorf("commit accepted LOCALITY block: %w", err)
	}
	block.status = choices.Accepted
	vm.lastAcceptedID = block.id
	vm.preferredID = block.id
	vm.blocks[block.id] = block
	filteredPending := vm.pendingOrder[:0]
	for _, transitionID := range vm.pendingOrder {
		if _, exists := vm.pending[transitionID]; exists {
			filteredPending = append(filteredPending, transitionID)
		}
	}
	vm.pendingOrder = filteredPending
	if len(vm.pendingOrder) == 0 {
		vm.drainNotifyLocked()
	} else {
		vm.signalPendingLocked()
	}
	return nil
}

func (vm *VM) rejectBlock(block *localityBlock) error {
	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return errNotInitialized
	}
	block.status = choices.Rejected
	block.postState = nil
	if block.builtLocally {
		for _, transition := range block.wire.Transitions {
			transitionID, err := transition.ID()
			if err != nil {
				continue
			}
			transitionIDHex := hex.EncodeToString(transitionID[:])
			if _, exists := vm.pending[transitionIDHex]; !exists {
				vm.pending[transitionIDHex] = transition
				vm.pendingOrder = append(vm.pendingOrder, transitionIDHex)
			}
		}
	}
	vm.signalPendingLocked()
	return nil
}

func (vm *VM) getBlockLocked(blockID ids.ID) (*localityBlock, error) {
	if block, exists := vm.blocks[blockID]; exists {
		return block, nil
	}
	block, err := vm.loadAcceptedBlockLocked(blockID)
	if err != nil {
		return nil, err
	}
	vm.blocks[blockID] = block
	return block, nil
}

func (vm *VM) loadAcceptedBlockLocked(blockID ids.ID) (*localityBlock, error) {
	blockBytes, err := vm.db.Get(blockKey(blockID))
	if err != nil {
		return nil, err
	}
	stateBytes, err := vm.db.Get(stateByBlockKey(blockID))
	if err != nil {
		return nil, err
	}
	state := new(RuntimeState)
	if err := decodeStrict(stateBytes, state); err != nil {
		return nil, err
	}
	if err := state.validateLoaded(); err != nil {
		return nil, err
	}
	if blockID == vm.genesisBlock.id {
		block := newGenesisBlock(vm, vm.genesisBytes, state)
		return block, nil
	}
	block, err := blockFromBytes(vm, blockBytes, choices.Accepted)
	if err != nil {
		return nil, err
	}
	if block.id != blockID {
		return nil, errors.New("stored block bytes do not match requested block ID")
	}
	block.postState = state
	return block, nil
}

func (vm *VM) acceptedStateLocked() (*RuntimeState, error) {
	block, err := vm.getBlockLocked(vm.lastAcceptedID)
	if err != nil {
		return nil, err
	}
	return block.postState.clone()
}

func (vm *VM) signalPendingLocked() {
	if len(vm.pendingOrder) == 0 {
		return
	}
	select {
	case vm.notify <- struct{}{}:
	default:
	}
}

func (vm *VM) drainNotifyLocked() {
	select {
	case <-vm.notify:
	default:
	}
}

func blockKey(blockID ids.ID) []byte {
	return append([]byte("locality/block/"), blockID[:]...)
}

func stateByBlockKey(blockID ids.ID) []byte {
	return append([]byte("locality/state/block/"), blockID[:]...)
}

func heightKey(height uint64) []byte {
	key := make([]byte, len("locality/height/")+8)
	copy(key, "locality/height/")
	binary.BigEndian.PutUint64(key[len("locality/height/"):], height)
	return key
}

func transitionKey(transitionID string) []byte { return []byte("locality/transition/" + transitionID) }
func receiptKey(transitionID string) []byte    { return []byte("locality/receipt/" + transitionID) }

func (*VM) Connected(context.Context, ids.NodeID, *version.Application) error { return nil }
func (*VM) Disconnected(context.Context, ids.NodeID) error                    { return nil }
func (*VM) AppRequest(context.Context, ids.NodeID, uint32, time.Time, []byte) error {
	return nil
}
func (*VM) AppRequestFailed(context.Context, ids.NodeID, uint32, *common.AppError) error {
	return nil
}
func (*VM) AppResponse(context.Context, ids.NodeID, uint32, []byte) error { return nil }
func (*VM) AppGossip(context.Context, ids.NodeID, []byte) error           { return nil }

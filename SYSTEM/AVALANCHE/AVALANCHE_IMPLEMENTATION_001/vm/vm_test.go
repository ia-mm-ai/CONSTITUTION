package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
)

func initializeTestVM(t *testing.T, db *memdb.Database) (*VM, []byte, *Genesis, testAuthorities) {
	t.Helper()
	genesisBytes, genesis, keys := testGenesis(t)
	vm := new(VM)
	if err := vm.Initialize(context.Background(), nil, db, genesisBytes, nil, nil, nil, nil); err != nil {
		t.Fatalf("initialize VM: %v", err)
	}
	if err := vm.SetState(context.Background(), snow.NormalOp); err != nil {
		t.Fatalf("set normal operation: %v", err)
	}
	return vm, genesisBytes, genesis, keys
}

func acceptedState(t *testing.T, vm *VM) *RuntimeState {
	t.Helper()
	vm.lock.Lock()
	defer vm.lock.Unlock()
	state, err := vm.acceptedStateLocked()
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func TestVMLifecyclePersistsAndRestarts(t *testing.T) {
	db := memdb.New()
	vm, genesisBytes, genesis, keys := initializeTestVM(t, db)
	state := acceptedState(t, vm)
	transition := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-001", BoundPayload{
		LocusID: "LOCUS-001", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	transitionBytes, err := transition.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	transitionID, err := vm.issueTransition(transitionBytes)
	if err != nil {
		t.Fatalf("issue transition: %v", err)
	}
	waitCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	message, err := vm.WaitForEvent(waitCtx)
	if err != nil || message == 0 {
		t.Fatalf("wait for pending transition: message=%v err=%v", message, err)
	}

	built, err := vm.BuildBlock(context.Background())
	if err != nil {
		t.Fatalf("build block: %v", err)
	}
	block := built.(*localityBlock)
	if err := block.Verify(context.Background()); err != nil {
		t.Fatalf("verify block: %v", err)
	}
	if block.Status() != choices.Processing {
		t.Fatalf("verified status = %s", block.Status())
	}
	if err := block.Accept(context.Background()); err != nil {
		t.Fatalf("accept block: %v", err)
	}
	if block.Status() != choices.Accepted {
		t.Fatalf("accepted status = %s", block.Status())
	}
	lastAccepted, err := vm.LastAccepted(context.Background())
	if err != nil || lastAccepted != block.ID() {
		t.Fatalf("last accepted = %s err=%v", lastAccepted, err)
	}
	atHeight, err := vm.GetBlockIDAtHeight(context.Background(), 1)
	if err != nil || atHeight != block.ID() {
		t.Fatalf("height index = %s err=%v", atHeight, err)
	}
	if _, err := db.Get(receiptKey(transitionID)); err != nil {
		t.Fatalf("receipt not persisted: %v", err)
	}
	if got := acceptedState(t, vm); got.Revision != 1 || got.ActiveLocusID != "LOCUS-001" {
		t.Fatalf("accepted state revision=%d active=%q", got.Revision, got.ActiveLocusID)
	}

	restarted := new(VM)
	if err := restarted.Initialize(context.Background(), nil, db, genesisBytes, nil, nil, nil, nil); err != nil {
		t.Fatalf("restart VM: %v", err)
	}
	if got := acceptedState(t, restarted); got.Revision != 1 || got.StateCommitment != block.postState.StateCommitment {
		t.Fatalf("restarted state revision=%d commitment=%s", got.Revision, got.StateCommitment)
	}
	parsed, err := restarted.ParseBlock(context.Background(), block.Bytes())
	if err != nil || parsed.ID() != block.ID() || parsed.(*localityBlock).Status() != choices.Accepted {
		t.Fatalf("parse accepted block after restart: id=%v err=%v", parsed, err)
	}
}

func TestRejectedLocallyBuiltBlockRequeuesTransition(t *testing.T) {
	db := memdb.New()
	vm, _, genesis, keys := initializeTestVM(t, db)
	state := acceptedState(t, vm)
	transition := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-001", BoundPayload{
		LocusID: "LOCUS-001", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	bytes, _ := transition.Bytes()
	if _, err := vm.issueTransition(bytes); err != nil {
		t.Fatal(err)
	}
	built, err := vm.BuildBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	block := built.(*localityBlock)
	if err := block.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := block.Reject(context.Background()); err != nil {
		t.Fatal(err)
	}
	if block.Status() != choices.Rejected {
		t.Fatalf("status = %s", block.Status())
	}
	if _, err := vm.BuildBlock(context.Background()); err != nil {
		t.Fatalf("requeued transition was not buildable: %v", err)
	}
}

func TestDraftFillsConsensusContextWithoutChangingState(t *testing.T) {
	db := memdb.New()
	vm, _, genesis, keys := initializeTestVM(t, db)
	stateBefore := acceptedState(t, vm)
	payload, _ := json.Marshal(BoundPayload{
		LocusID: "LOCUS-001", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	unsigned, err := vm.draftTransition(&TransitionDraftRequest{
		Operation: opBound, ActorID: genesis.Locality.ID,
		ActorPublicKey: genesis.Locality.Authority.PublicKey, LocusID: "LOCUS-001", ObservedAt: 1, Payload: payload,
	})
	if err != nil {
		t.Fatalf("draft transition: %v", err)
	}
	if unsigned.Revision != 1 || unsigned.Nonce != 0 || unsigned.PreviousStateCommitment != stateBefore.StateCommitment {
		t.Fatalf("draft context = revision %d nonce %d predecessor %s", unsigned.Revision, unsigned.Nonce, unsigned.PreviousStateCommitment)
	}
	if got := acceptedState(t, vm); got.Revision != 0 || got.StateCommitment != stateBefore.StateCommitment {
		t.Fatal("draft mutated accepted state")
	}
	transition, err := signTransition(*unsigned, keys.host)
	if err != nil {
		t.Fatal(err)
	}
	transitionBytes, _ := transition.Bytes()
	if _, err := vm.issueTransition(transitionBytes); err != nil {
		t.Fatalf("issue drafted transition: %v", err)
	}
}

func TestRestartDetectsPersistedStateTampering(t *testing.T) {
	db := memdb.New()
	vm, genesisBytes, genesis, keys := initializeTestVM(t, db)
	state := acceptedState(t, vm)
	transition := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-001", BoundPayload{
		LocusID: "LOCUS-001", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	transitionBytes, _ := transition.Bytes()
	if _, err := vm.issueTransition(transitionBytes); err != nil {
		t.Fatal(err)
	}
	built, _ := vm.BuildBlock(context.Background())
	block := built.(*localityBlock)
	if err := block.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := block.Accept(context.Background()); err != nil {
		t.Fatal(err)
	}
	tampered := *block.postState
	tampered.StateCommitment = digestText("tampered commitment")
	tamperedBytes, _ := json.Marshal(&tampered)
	if err := db.Put(stateByBlockKey(block.ID()), tamperedBytes); err != nil {
		t.Fatal(err)
	}
	restarted := new(VM)
	err := restarted.Initialize(context.Background(), nil, db, genesisBytes, nil, nil, nil, nil)
	if err == nil || !errors.Is(err, errNotInitialized) && !containsText(err.Error(), "commitment mismatch") {
		t.Fatalf("tampered state should fail restart, got %v", err)
	}
}

func TestImmutableGenesisMismatchIsRejected(t *testing.T) {
	db := memdb.New()
	_, genesisBytes, _, _ := initializeTestVM(t, db)
	altered := append([]byte(nil), genesisBytes...)
	altered = append(altered, '\n')
	vm := new(VM)
	if err := vm.Initialize(context.Background(), nil, db, altered, nil, nil, nil, nil); err == nil {
		t.Fatal("different genesis bytes were accepted on existing database")
	}
}

func TestFutureObservedAtIsRejectedBySubmissionAndBlockVerification(t *testing.T) {
	vm, _, genesis, keys := initializeTestVM(t, memdb.New())
	state := acceptedState(t, vm)
	payload, _ := json.Marshal(BoundPayload{
		LocusID: "LOCUS-FUTURE", PurposeSHA256: digestText("future purpose"),
		ClosureConditionSHA256: digestText("future closure"), CapacityCeilingUnits: 100,
	})
	future := time.Now().UTC().Add(time.Hour).Unix()
	transition, err := signTransition(UnsignedTransition{
		Schema: transitionSchema, Operation: opBound, Revision: 1,
		PreviousStateCommitment: state.StateCommitment, ActorID: genesis.Locality.ID,
		Nonce: 0, LocusID: "LOCUS-FUTURE", ObservedAt: future,
		Effect: operationEffects[opBound], Payload: payload,
	}, keys.host)
	if err != nil {
		t.Fatal(err)
	}
	transitionBytes, _ := transition.Bytes()
	if _, err := vm.issueTransition(transitionBytes); err == nil || !containsText(err.Error(), "future-skew") {
		t.Fatalf("future observation was accepted for local submission: %v", err)
	}
	parent, err := vm.GetBlock(context.Background(), vm.genesisBlock.id)
	if err != nil {
		t.Fatal(err)
	}
	wire := &WireBlock{
		Schema: blockSchema, ParentID: parent.ID().String(), Timestamp: time.Now().UTC().Unix(),
		Height: 1, Transitions: []*Transition{transition},
	}
	block, err := newWireBlock(vm, wire, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := block.Verify(context.Background()); err == nil || !containsText(err.Error(), "future-skew") {
		t.Fatalf("future observation was accepted from a proposed block: %v", err)
	}
}

func containsText(value, fragment string) bool {
	return len(value) >= len(fragment) && (value == fragment || len(fragment) == 0 || findSubstring(value, fragment))
}

func findSubstring(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}

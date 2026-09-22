package main

// Regression coverage for the shared operation-scope contract. These tests
// prove the structural repair of the predecessor coupling defect: the FIELD
// required every operation except BOUND to name the active locus while the VM
// required DECLARE_CAPACITY to use an empty locus_id. One shared normative
// contract now classifies every operation, both runtimes load the same file,
// and unknown or unclassified operations fail closed.

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/ava-labs/avalanchego/database/memdb"
	"strings"
	"testing"
)

func lifecycleWithActiveLocus(t *testing.T) (*RuntimeState, *Genesis, testAuthorities, string, string) {
	t.Helper()
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	const hostID = "PRESENCE-RUNTIME-TEST"
	const locusID = "LOCUS-SCOPE-001"
	bound := makeTransition(t, state, hostID, keys.host, opBound, locusID, BoundPayload{
		LocusID: locusID, PurposeSHA256: digestText("scope purpose"),
		ClosureConditionSHA256: digestText("scope closure"), CapacityCeilingUnits: 100,
	})
	state, _, _ = mustApply(t, state, genesis, bound)
	return state, genesis, keys, hostID, locusID
}

// unsignedFor builds an UnsignedTransition without signing it, so tests can
// probe validation of shapes that signTransition itself refuses to sign.
func unsignedFor(state *RuntimeState, actorID string, key ed25519.PrivateKey, operation, locusID string, payload json.RawMessage) UnsignedTransition {
	return UnsignedTransition{
		Schema:                  transitionSchema,
		Operation:               operation,
		Revision:                state.Revision + 1,
		PreviousStateCommitment: state.StateCommitment,
		ActorID:                 actorID,
		ActorPublicKey:          hex.EncodeToString(key.Public().(ed25519.PublicKey)),
		Nonce:                   state.NextNonces[actorID],
		LocusID:                 locusID,
		ObservedAt:              int64(state.Revision + 1000),
		Effect:                  operationEffects[operation],
		Payload:                 payload,
	}
}

// Requirement 1: BOUND succeeds with a new locus ID.
func TestBoundSucceedsWithNewLocusID(t *testing.T) {
	state, _, _, _, locusID := lifecycleWithActiveLocus(t)
	if state.ActiveLocusID != locusID {
		t.Fatalf("BOUND did not open the proposed new locus: active=%q", state.ActiveLocusID)
	}
}

// Requirement 2: DECLARE_CAPACITY succeeds with empty locus_id while a locus
// is active. This is the exact predecessor coupling failure, repaired.
func TestDeclareCapacityEmptyLocusSucceedsWhileLocusActive(t *testing.T) {
	state, genesis, keys, hostID, locusID := lifecycleWithActiveLocus(t)
	if state.ActiveLocusID != locusID {
		t.Fatalf("precondition failed: no active locus")
	}
	declare := makeTransition(t, state, hostID, keys.host, opDeclareCapacity, "", DeclareCapacityPayload{
		ActualUnits: 90, ResourceCommitmentSHA256: digestText("resources"), BasisSHA256: digestText("basis"),
	})
	state, receipt, _ := mustApply(t, state, genesis, declare)
	if receipt.Effect != "SETS_SIGNED_LOCAL_CAPACITY_ACCOUNT_ONLY" {
		t.Fatalf("unexpected declare-capacity effect %q", receipt.Effect)
	}
	if state.ActiveLocusID != locusID {
		t.Fatalf("body-local capacity declaration disturbed the active locus: %q", state.ActiveLocusID)
	}
}

// Requirement 3: DECLARE_CAPACITY with the active locus ID is rejected before
// signing or submission.
func TestDeclareCapacityWithActiveLocusIDRejectedBeforeSigning(t *testing.T) {
	state, genesis, keys, hostID, locusID := lifecycleWithActiveLocus(t)

	if err := validateScopeShape(opDeclareCapacity, locusID); err == nil {
		t.Fatal("pre-signing shape validation accepted DECLARE_CAPACITY with a locus_id")
	}

	vm, _, vmGenesis, vmKeys := initializeTestVM(t, memdb.New())
	boundPayload, _ := json.Marshal(BoundPayload{
		LocusID: locusID, PurposeSHA256: digestText("scope purpose"),
		ClosureConditionSHA256: digestText("scope closure"), CapacityCeilingUnits: 100,
	})
	boundDraft, err := vm.draftTransition(&TransitionDraftRequest{
		Operation: opBound, ActorID: vmGenesis.Locality.ID, ActorPublicKey: vmGenesis.Locality.Authority.PublicKey,
		LocusID: locusID, ObservedAt: 1, Payload: boundPayload,
	})
	if err != nil {
		t.Fatalf("draft BOUND: %v", err)
	}
	bound, err := signTransition(*boundDraft, vmKeys.host)
	if err != nil {
		t.Fatal(err)
	}
	boundBytes, err := bound.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := vm.issueTransition(boundBytes); err != nil {
		t.Fatalf("submit BOUND: %v", err)
	}
	declarePayload, _ := json.Marshal(DeclareCapacityPayload{
		ActualUnits: 90, ResourceCommitmentSHA256: digestText("resources"), BasisSHA256: digestText("basis"),
	})
	if _, err := vm.draftTransition(&TransitionDraftRequest{
		Operation: opDeclareCapacity, ActorID: vmGenesis.Locality.ID, ActorPublicKey: vmGenesis.Locality.Authority.PublicKey,
		LocusID: locusID, ObservedAt: 2, Payload: declarePayload,
	}); err == nil || !strings.Contains(err.Error(), "empty transition locus_id") {
		t.Fatalf("draft with active locus should be rejected before signing, got %v", err)
	}
	if _, err := vm.draftTransition(&TransitionDraftRequest{
		Operation: opDeclareCapacity, ActorID: vmGenesis.Locality.ID, ActorPublicKey: vmGenesis.Locality.Authority.PublicKey,
		LocusID: "", ObservedAt: 2, Payload: declarePayload,
	}); err != nil {
		t.Fatalf("draft with empty locus should be accepted while a locus is pending/active, got %v", err)
	}

	unsigned := unsignedFor(state, hostID, keys.host, opDeclareCapacity, locusID, mustMarshal(t, DeclareCapacityPayload{
		ActualUnits: 90, ResourceCommitmentSHA256: digestText("resources"), BasisSHA256: digestText("basis"),
	}))
	if _, err := signTransition(unsigned, keys.host); err == nil || !strings.Contains(err.Error(), "empty transition locus_id") {
		t.Fatalf("signing should refuse DECLARE_CAPACITY with a locus_id, got %v", err)
	}
	transition := &Transition{Unsigned: unsigned}
	if err := transition.ValidateSyntax(); err == nil || !strings.Contains(err.Error(), "empty transition locus_id") {
		t.Fatalf("syntax validation should reject DECLARE_CAPACITY with a locus_id, got %v", err)
	}
	if _, _, err := applyToClone(state, genesis, transition); err == nil {
		t.Fatal("apply accepted DECLARE_CAPACITY bound to the active locus")
	}
}

// Requirement 4: every ACTIVE_LOCUS operation rejects an empty or different
// locus.
func TestActiveLocusOperationsRejectEmptyOrForeignLocus(t *testing.T) {
	state, _, keys, hostID, _ := lifecycleWithActiveLocus(t)
	for operation, entry := range scopeContract.Operations {
		if entry.Scope != scopeActiveLocus {
			continue
		}
		empty := unsignedFor(state, hostID, keys.host, operation, "", json.RawMessage(`{}`))
		emptyTransition := &Transition{Unsigned: empty}
		if err := emptyTransition.ValidateSyntax(); err == nil || !strings.Contains(err.Error(), "nonempty transition locus_id") {
			t.Fatalf("%s accepted an empty locus at syntax validation: %v", operation, err)
		}
		if err := state.enforceOperationScope(&empty); err == nil {
			t.Fatalf("%s accepted an empty locus against state", operation)
		}
		foreign := unsignedFor(state, hostID, keys.host, operation, "LOCUS-FOREIGN-001", json.RawMessage(`{}`))
		if err := state.enforceOperationScope(&foreign); err == nil || !strings.Contains(err.Error(), "exact accepted active locus") {
			t.Fatalf("%s accepted a foreign locus: %v", operation, err)
		}
	}
}

// Requirement 5: every BODY_LOCAL operation rejects a nonempty locus.
func TestBodyLocalOperationsRejectNonemptyLocus(t *testing.T) {
	state, _, keys, hostID, locusID := lifecycleWithActiveLocus(t)
	for operation, entry := range scopeContract.Operations {
		if entry.Scope != scopeBodyLocal {
			continue
		}
		unsigned := unsignedFor(state, hostID, keys.host, operation, locusID, json.RawMessage(`{}`))
		transition := &Transition{Unsigned: unsigned}
		if err := transition.ValidateSyntax(); err == nil || !strings.Contains(err.Error(), "empty transition locus_id") {
			t.Fatalf("%s accepted a nonempty locus at syntax validation: %v", operation, err)
		}
		if err := state.enforceOperationScope(&unsigned); err == nil {
			t.Fatalf("%s accepted a nonempty locus against state", operation)
		}
	}
}

// Requirement 6: dormant-body operations reject an active locus.
func TestDormantBodyOperationsRejectActiveLocus(t *testing.T) {
	state, _, keys, hostID, locusID := lifecycleWithActiveLocus(t)
	if state.ActiveLocusID != locusID {
		t.Fatal("precondition failed: no active locus")
	}
	for operation, entry := range scopeContract.Operations {
		if entry.Scope != scopeDormantBody {
			continue
		}
		unsigned := unsignedFor(state, hostID, keys.host, operation, "", json.RawMessage(`{}`))
		if err := state.enforceOperationScope(&unsigned); err == nil || !strings.Contains(err.Error(), "requires no active locus") {
			t.Fatalf("%s accepted an active locus: %v", operation, err)
		}
		withLocus := unsignedFor(state, hostID, keys.host, operation, locusID, json.RawMessage(`{}`))
		withLocusTransition := &Transition{Unsigned: withLocus}
		if err := withLocusTransition.ValidateSyntax(); err == nil || !strings.Contains(err.Error(), "empty transition locus_id") {
			t.Fatalf("%s accepted a nonempty locus at syntax validation: %v", operation, err)
		}
	}
}

// Requirement 7: CORRECT binds its target transition's scope rather than
// incidental current activity.
func TestCorrectBindsTargetScopeNotCurrentActivity(t *testing.T) {
	state, genesis, keys, hostID, locusID := lifecycleWithActiveLocus(t)
	declare := makeTransition(t, state, hostID, keys.host, opDeclareCapacity, "", DeclareCapacityPayload{
		ActualUnits: 90, ResourceCommitmentSHA256: digestText("resources"), BasisSHA256: digestText("basis"),
	})
	state, _, declareID := mustApply(t, state, genesis, declare)

	// The target is body-local (empty locus). While LOCUS-SCOPE-001 is
	// active, a correction forced into the active locus must fail ...
	wrongScope := makeTransition(t, state, hostID, keys.host, opCorrect, locusID, CorrectPayload{
		TargetTransitionID: declareID, ReplacementCommitment: digestText("corrected"), ReasonSHA256: digestText("reason"),
	})
	if _, _, err := applyToClone(state, genesis, wrongScope); err == nil || !strings.Contains(err.Error(), "exact scope of its target") {
		t.Fatalf("correction forced into the active locus should fail, got %v", err)
	}

	// ... and the correction bound to the target's own (empty) scope must
	// succeed even though another locus is open.
	correct := makeTransition(t, state, hostID, keys.host, opCorrect, "", CorrectPayload{
		TargetTransitionID: declareID, ReplacementCommitment: digestText("corrected"), ReasonSHA256: digestText("reason"),
	})
	state, _, correctionID := mustApply(t, state, genesis, correct)
	if got := state.Corrections[declareID][0].CorrectionTransitionID; got != correctionID {
		t.Fatalf("correction link = %s", got)
	}

	// A locus-scoped target requires the correction to bind that exact locus.
	boundID := ""
	for id, event := range state.Events {
		if event.Operation == opBound {
			boundID = id
		}
	}
	if boundID == "" {
		t.Fatal("no BOUND event recorded")
	}
	wrongEmpty := makeTransition(t, state, hostID, keys.host, opCorrect, "", CorrectPayload{
		TargetTransitionID: boundID, ReplacementCommitment: digestText("corrected"), ReasonSHA256: digestText("reason"),
	})
	if _, _, err := applyToClone(state, genesis, wrongEmpty); err == nil || !strings.Contains(err.Error(), "exact scope of its target") {
		t.Fatalf("correction of a locus-scoped target with empty locus should fail, got %v", err)
	}
}

// Requirement 8: every declared operation appears exactly once in the scope
// contract, at the exact byte level of the shared file.
func TestEveryOperationAppearsExactlyOnceInScopeContract(t *testing.T) {
	decoder := json.NewDecoder(bytes.NewReader(operationScopeContractBytes))
	depth := 0
	inOperations := false
	operationsDepth := 0
	seen := map[string]int{}
	expectKey := false
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch value := token.(type) {
		case json.Delim:
			switch value {
			case '{', '[':
				depth++
				if inOperations && depth == operationsDepth+1 {
					expectKey = true
				}
			case '}', ']':
				if inOperations && depth == operationsDepth+1 {
					inOperations = false
				}
				depth--
			}
		case string:
			if depth == 1 && value == "operations" {
				inOperations = true
				operationsDepth = depth
				continue
			}
			if inOperations && depth == operationsDepth+1 && expectKey {
				seen[value]++
				expectKey = false
				continue
			}
		}
		if inOperations && depth == operationsDepth+1 && !expectKey {
			// After consuming an operation's value object the decoder returns
			// to this depth; the next string token is the next key.
			expectKey = true
		}
	}
	if len(seen) == 0 {
		t.Fatal("no operations found in the contract bytes")
	}
	for _, operation := range allOperations {
		if seen[operation] != 1 {
			t.Fatalf("operation %s appears %d times in the scope contract; want exactly once", operation, seen[operation])
		}
	}
	if len(seen) != len(allOperations) {
		t.Fatalf("scope contract declares %d operations; VM inventory declares %d", len(seen), len(allOperations))
	}
}

// Requirement 9 (VM side): the VM operation/effect/scope inventory agrees with
// the shared contract. The FIELD-side agreement test loads the identical file
// (field/test/scope-contract.test.js), so agreement is mechanically enforced
// through one single source.
func TestVMInventoryAgreesWithScopeContract(t *testing.T) {
	if len(allOperations) != len(scopeContract.Operations) {
		t.Fatalf("VM inventory has %d operations; contract has %d", len(allOperations), len(scopeContract.Operations))
	}
	for _, operation := range allOperations {
		entry, exists := scopeContract.Operations[operation]
		if !exists {
			t.Fatalf("VM operation %s is missing from the scope contract", operation)
		}
		if entry.Effect != operationEffects[operation] {
			t.Fatalf("operation %s effect mismatch: contract %q runtime %q", operation, entry.Effect, operationEffects[operation])
		}
		scope, err := operationScope(operation)
		if err != nil {
			t.Fatalf("operation %s has no scope: %v", operation, err)
		}
		if scope != entry.Scope {
			t.Fatalf("operation %s scope mismatch: contract %q runtime %q", operation, entry.Scope, scope)
		}
	}
	sum := sha256.Sum256(operationScopeContractBytes)
	t.Logf("operation-scope contract sha256=%s", hex.EncodeToString(sum[:]))
}

// Requirement 10: unknown operations fail closed everywhere.
func TestUnknownOperationsFailClosed(t *testing.T) {
	if _, err := operationScope("NOT_A_DECLARED_OPERATION"); err == nil {
		t.Fatal("operationScope accepted an unknown operation")
	}
	if err := validateScopeShape("NOT_A_DECLARED_OPERATION", ""); err == nil {
		t.Fatal("validateScopeShape accepted an unknown operation")
	}
	state, _, _, _, _ := lifecycleWithActiveLocus(t)
	if err := state.enforceOperationScope(&UnsignedTransition{Operation: "NOT_A_DECLARED_OPERATION"}); err == nil {
		t.Fatal("enforceOperationScope accepted an unknown operation")
	}
	vm, _, _, _ := initializeTestVM(t, memdb.New())
	if _, err := vm.draftTransition(&TransitionDraftRequest{Operation: "NOT_A_DECLARED_OPERATION"}); err == nil {
		t.Fatal("draftTransition accepted an unknown operation")
	}
}

func mustMarshal(t *testing.T, value any) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

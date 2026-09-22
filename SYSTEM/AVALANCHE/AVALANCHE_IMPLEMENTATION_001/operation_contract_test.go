package main

import (
	"encoding/json"
	"testing"
)

func TestOperationContractIsExhaustiveAndUnique(t *testing.T) {
	var document operationContractDocument
	if err := json.Unmarshal(operationContractBytes, &document); err != nil {
		t.Fatal(err)
	}
	if got, want := len(document.Operations), 21; got != want {
		t.Fatalf("operation contract has %d entries; want %d", got, want)
	}
	seen := make(map[string]bool, len(document.Operations))
	for _, definition := range document.Operations {
		if seen[definition.Operation] {
			t.Fatalf("operation %s appears more than once", definition.Operation)
		}
		seen[definition.Operation] = true
		if operationEffects[definition.Operation] != definition.Effect {
			t.Fatalf("effect inventory differs for %s", definition.Operation)
		}
		if operationDefinitions[definition.Operation].Scope != definition.Scope {
			t.Fatalf("scope inventory differs for %s", definition.Operation)
		}
	}
}

func TestBoundThenBodyLocalCapacityWhileLocusActive(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	hostID := state.FormationAuthorityID
	bound := makeTransition(t, state, hostID, keys.host, opBound, "LOCUS-SCOPE-001", BoundPayload{
		LocusID: "LOCUS-SCOPE-001", PurposeSHA256: digestText("scope purpose"),
		ClosureConditionSHA256: digestText("scope closure"), CapacityCeilingUnits: 50,
	})
	state, _, _ = mustApply(t, state, genesis, bound)
	if state.ActiveLocusID != "LOCUS-SCOPE-001" {
		t.Fatal("BOUND did not open the proposed locus")
	}

	declaration := makeTransition(t, state, hostID, keys.host, opDeclareCapacity, "", DeclareCapacityPayload{
		ActualUnits: 80, ResourceCommitmentSHA256: digestText("capacity resource"),
		BasisSHA256: digestText("capacity basis"),
	})
	state, receipt, _ := mustApply(t, state, genesis, declaration)
	if receipt.LocusID != "" || state.ActiveLocusID != "LOCUS-SCOPE-001" || state.Capacity.DeclaredActualUnits != 80 {
		t.Fatal("body-local capacity declaration did not preserve the active locus")
	}

	bad := makeTransition(t, state, hostID, keys.host, opDeclareCapacity, state.ActiveLocusID, DeclareCapacityPayload{
		ActualUnits: 70, ResourceCommitmentSHA256: digestText("bad capacity resource"),
		BasisSHA256: digestText("bad capacity basis"),
	})
	if _, _, err := applyToClone(state, genesis, bad); err == nil {
		t.Fatal("DECLARE_CAPACITY with the active locus must fail")
	}
}

func TestEveryScopeClassFailsClosed(t *testing.T) {
	activeOperations := []string{
		opPresentForm, opGateDisposition, opEnter, opCheckpointDeparture, opReenter,
		opObserveCrossing, opMatterDisposition, opRecordEmergence, opExit, opClose,
		opReclaimOffer,
	}
	bodyOperations := []string{
		opIncorporateResidue, opDeclareCapacity, opPulse, opRegisterAuthority,
		opProposeSuccessor, opAttestSuccessor,
	}
	state := &RuntimeState{
		ActiveLocusID: "LOCUS-CURRENT",
		Events: map[string]EventRecord{
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {
				TransitionID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				LocusID:      "LOCUS-TARGET",
			},
		},
		Loci: map[string]*LocusState{},
	}
	for _, operation := range activeOperations {
		for _, locusID := range []string{"", "LOCUS-DIFFERENT"} {
			if err := validateOperationScope(state, operation, locusID, nil); err == nil {
				t.Errorf("%s accepted invalid locus %q", operation, locusID)
			}
		}
	}
	for _, operation := range bodyOperations {
		if err := validateOperationScope(state, operation, state.ActiveLocusID, nil); err == nil {
			t.Errorf("%s accepted a nonempty locus", operation)
		}
	}
	for _, operation := range []string{opExhaustFormation, opActivateSuccessor} {
		if err := validateOperationScope(state, operation, "", nil); err == nil {
			t.Errorf("%s accepted an active locus", operation)
		}
	}
	correction, err := json.Marshal(CorrectPayload{
		TargetTransitionID:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ReplacementCommitment: digestText("replacement"),
		ReasonSHA256:          digestText("reason"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := validateOperationScope(state, opCorrect, "LOCUS-TARGET", correction); err != nil {
		t.Fatalf("CORRECT rejected its target transition scope: %v", err)
	}
	if err := validateOperationScope(state, opCorrect, state.ActiveLocusID, correction); err == nil {
		t.Fatal("CORRECT accepted incidental current activity instead of target scope")
	}
	if err := validateOperationScope(state, "UNKNOWN_OPERATION", "", nil); err == nil {
		t.Fatal("unknown operation did not fail closed")
	}
}

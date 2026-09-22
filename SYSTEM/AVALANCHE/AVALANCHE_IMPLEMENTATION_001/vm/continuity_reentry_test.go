package main

import (
	"crypto/ed25519"
	"strings"
	"testing"
)

func TestReentryCarriesStateWithoutReopeningPriorLocus(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	participantID := operationalActorID(keys.participant.Public().(ed25519.PublicKey))
	const (
		firstLocus  = "LOCUS-CONTINUITY-001"
		secondLocus = "LOCUS-CONTINUITY-002"
	)

	bound := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, firstLocus, BoundPayload{
		LocusID: firstLocus, PurposeSHA256: digestText("first continuity locus"),
		ClosureConditionSHA256: digestText("first continuity closure"), CapacityCeilingUnits: 100,
	})
	state, _, _ = mustApply(t, state, genesis, bound)
	initialParticipantState := digestText("participant state at first ingress")
	present := makeTransition(t, state, participantID, keys.participant, opPresentForm, firstLocus, PresentFormPayload{
		LocalityReference: "DERIVED-LOCALITY-CONTINUITY-001", SourceReference: "CONTINUITY",
		SourceSHA256: digestText("participant continuity source"), NucleusVersion: "1",
		NucleusSHA256: digestText("participant nucleus 1"), FormID: formID, FormSHA256: formSHA256,
		StateCommitment: initialParticipantState, MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	})
	state, _, presentationID := mustApply(t, state, genesis, present)
	admit := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, firstLocus, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationID, Disposition: "ADMIT",
		ReasonSHA256: digestText("first admission"), WorkUnits: 10, ResolutionUnits: 2, OfferExpiresAt: 4000,
	})
	state, _, _ = mustApply(t, state, genesis, admit)
	enter := makeTransition(t, state, participantID, keys.participant, opEnter, firstLocus, EnterPayload{PresentationID: presentationID})
	state, _, firstEntryID := mustApply(t, state, genesis, enter)

	closeWithoutCheckpoint := makeTransition(t, state, genesis.Locality.ID, keys.host, opClose, firstLocus, ClosePayload{ClosureBasisSHA256: digestText("premature close")})
	if _, _, err := applyToClone(state, genesis, closeWithoutCheckpoint); err == nil || !strings.Contains(err.Error(), "lacks a departure checkpoint") {
		t.Fatalf("closure without participant checkpoint should fail, got %v", err)
	}
	exitWithoutCheckpoint := makeTransition(t, state, participantID, keys.participant, opExit, firstLocus, ExitPayload{
		DepartureCheckpointID: digestText("missing checkpoint"), ReasonSHA256: digestText("premature exit"),
	})
	if _, _, err := applyToClone(state, genesis, exitWithoutCheckpoint); err == nil || !strings.Contains(err.Error(), "departure checkpoint") {
		t.Fatalf("exit without participant checkpoint should fail, got %v", err)
	}

	departureDelta := digestText("bounded participant delta before departure")
	departureState, err := continuitySuccessorCommitment(participantID, initialParticipantState, departureDelta, 1)
	if err != nil {
		t.Fatal(err)
	}
	brokenCheckpoint := makeTransition(t, state, participantID, keys.participant, opCheckpointDeparture, firstLocus, CheckpointDeparturePayload{
		EntryTransitionID: firstEntryID, PresentationID: presentationID, FromStateCommitment: initialParticipantState,
		Passage:                  []ContinuityStep{{DeltaSHA256: departureDelta, SuccessorStateCommitment: digestText("forged successor")}},
		DepartureStateCommitment: digestText("forged successor"),
	})
	if _, _, err := applyToClone(state, genesis, brokenCheckpoint); err == nil || !strings.Contains(err.Error(), "does not succeed") {
		t.Fatalf("broken departure passage should fail, got %v", err)
	}
	checkpoint := makeTransition(t, state, participantID, keys.participant, opCheckpointDeparture, firstLocus, CheckpointDeparturePayload{
		EntryTransitionID: firstEntryID, PresentationID: presentationID, FromStateCommitment: initialParticipantState,
		Passage:                  []ContinuityStep{{DeltaSHA256: departureDelta, SuccessorStateCommitment: departureState}},
		DepartureStateCommitment: departureState,
	})
	state, checkpointReceipt, departureCheckpointID := mustApply(t, state, genesis, checkpoint)
	if !contains(checkpointReceipt.NonEffects, "NO_EXIT_BY_CHECKPOINT") || state.DepartureCheckpoints[departureCheckpointID].Status != "OPEN" {
		t.Fatal("checkpoint silently exited participant or omitted its evidence ceiling")
	}
	exit := makeTransition(t, state, participantID, keys.participant, opExit, firstLocus, ExitPayload{
		DepartureCheckpointID: departureCheckpointID, ReasonSHA256: digestText("bounded exit"),
	})
	state, _, _ = mustApply(t, state, genesis, exit)
	if state.DepartureCheckpoints[departureCheckpointID].Status != "SEALED" {
		t.Fatal("exit did not seal the departure checkpoint")
	}
	closeTransition := makeTransition(t, state, genesis.Locality.ID, keys.host, opClose, firstLocus, ClosePayload{ClosureBasisSHA256: digestText("first closure")})
	state, _, _ = mustApply(t, state, genesis, closeTransition)
	residue := state.Loci[firstLocus].Residues[participantID]
	if residue.DepartureCheckpointID != departureCheckpointID || residue.DepartureStateCommitment != departureState || residue.EntryTransitionID != firstEntryID {
		t.Fatalf("residue did not preserve the departure binding: %+v", residue)
	}

	boundSecond := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, secondLocus, BoundPayload{
		LocusID: secondLocus, PurposeSHA256: digestText("second continuity locus"),
		ClosureConditionSHA256: digestText("second continuity closure"), CapacityCeilingUnits: 100,
	})
	state, _, _ = mustApply(t, state, genesis, boundSecond)
	returnDelta := digestText("bounded participant delta while absent")
	returnState, err := continuitySuccessorCommitment(participantID, departureState, returnDelta, 1)
	if err != nil {
		t.Fatal(err)
	}
	presentAgain := makeTransition(t, state, participantID, keys.participant, opPresentForm, secondLocus, PresentFormPayload{
		LocalityReference: "DERIVED-LOCALITY-CONTINUITY-001", SourceReference: "CONTINUITY",
		SourceSHA256: digestText("participant continuity source"), NucleusVersion: "2",
		NucleusSHA256: digestText("participant nucleus 2"), FormID: formID, FormSHA256: formSHA256,
		StateCommitment: returnState, MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	})
	state, _, secondPresentationID := mustApply(t, state, genesis, presentAgain)
	admitAgain := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, secondLocus, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: secondPresentationID, Disposition: "ADMIT",
		ReasonSHA256: digestText("renewed admission"), WorkUnits: 12, ResolutionUnits: 3, OfferExpiresAt: 4000,
	})
	state, _, _ = mustApply(t, state, genesis, admitAgain)

	bypass := makeTransition(t, state, participantID, keys.participant, opEnter, secondLocus, EnterPayload{PresentationID: secondPresentationID})
	if _, _, err := applyToClone(state, genesis, bypass); err == nil || !strings.Contains(err.Error(), "must use REENTER") {
		t.Fatalf("returning actor bypassed carried-state re-entry, got %v", err)
	}
	wrongResidue := makeTransition(t, state, participantID, keys.participant, opReenter, secondLocus, ReenterPayload{
		PresentationID: secondPresentationID, PriorEntryTransitionID: firstEntryID,
		DepartureCheckpointID: departureCheckpointID, ResidueID: digestText("wrong residue"),
		Passage: []ContinuityStep{{DeltaSHA256: returnDelta, SuccessorStateCommitment: returnState}},
	})
	if _, _, err := applyToClone(state, genesis, wrongResidue); err == nil || !strings.Contains(err.Error(), "exact addressed residue") {
		t.Fatalf("re-entry accepted the wrong closure residue, got %v", err)
	}
	reenter := makeTransition(t, state, participantID, keys.participant, opReenter, secondLocus, ReenterPayload{
		PresentationID: secondPresentationID, PriorEntryTransitionID: firstEntryID,
		DepartureCheckpointID: departureCheckpointID, ResidueID: residue.ResidueID,
		Passage: []ContinuityStep{{DeltaSHA256: returnDelta, SuccessorStateCommitment: returnState}},
	})
	state, reentryReceipt, secondEntryID := mustApply(t, state, genesis, reenter)
	secondEntry := state.EntryHistory[secondEntryID]
	if secondEntry.IngressMode != "RENEWED_ENTRY" || secondEntry.PriorEntryTransitionID != firstEntryID || secondEntry.PriorDepartureCheckpointID != departureCheckpointID || secondEntry.PriorResidueID != residue.ResidueID {
		t.Fatalf("renewed entry omitted its lineage: %+v", secondEntry)
	}
	if state.DepartureCheckpoints[departureCheckpointID].Status != "CONSUMED" || state.DepartureCheckpoints[departureCheckpointID].ConsumedByEntryTransitionID != secondEntryID {
		t.Fatal("re-entry did not consume its checkpoint exactly once")
	}
	if reentryReceipt.Capacity.UnresolvedUnits != 12 || reentryReceipt.Capacity.CorrectionEgressUnits != 13 {
		t.Fatalf("re-entry did not reserve its whole lifecycle: %+v", reentryReceipt.Capacity)
	}
	if state.Loci[firstLocus].Phase != "CLOSED" || state.Loci[firstLocus].Entries[participantID].Status != "EXITED" {
		t.Fatal("re-entry reopened or rewrote the prior locus")
	}
	if _, exists := state.EntryHistory[firstEntryID]; !exists || len(state.PresentationHistory) != 2 {
		t.Fatal("renewed ingress erased predecessor history")
	}
	if !contains(reentryReceipt.NonEffects, "NO_RESTORATION_OR_RESUMPTION_OF_PRIOR_LOCUS") {
		t.Fatal("re-entry receipt omitted the non-restoration boundary")
	}
}

func TestDepartureCheckpointIsSingleUseAndCannotBranch(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	participantID := operationalActorID(keys.participant.Public().(ed25519.PublicKey))
	bound := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-SINGLE-USE", BoundPayload{
		LocusID: "LOCUS-SINGLE-USE", PurposeSHA256: digestText("single use"), ClosureConditionSHA256: digestText("single use closure"), CapacityCeilingUnits: 100,
	})
	state, _, _ = mustApply(t, state, genesis, bound)
	state, _, presentationID := presentParticipant(t, state, genesis, keys.participant, "PARTICIPANT-SINGLE-USE")
	admit := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, state.ActiveLocusID, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationID, Disposition: "ADMIT", ReasonSHA256: digestText("admit"),
		WorkUnits: 5, ResolutionUnits: 1, OfferExpiresAt: 4000,
	})
	state, _, _ = mustApply(t, state, genesis, admit)
	enter := makeTransition(t, state, participantID, keys.participant, opEnter, state.ActiveLocusID, EnterPayload{PresentationID: presentationID})
	state, _, entryID := mustApply(t, state, genesis, enter)
	stateCommitment := digestText("PARTICIPANT-SINGLE-USE state")
	checkpoint := makeTransition(t, state, participantID, keys.participant, opCheckpointDeparture, state.ActiveLocusID, CheckpointDeparturePayload{
		EntryTransitionID: entryID, PresentationID: presentationID, FromStateCommitment: stateCommitment,
		Passage: []ContinuityStep{}, DepartureStateCommitment: stateCommitment,
	})
	state, _, checkpointID := mustApply(t, state, genesis, checkpoint)
	exit := makeTransition(t, state, participantID, keys.participant, opExit, state.ActiveLocusID, ExitPayload{DepartureCheckpointID: checkpointID, ReasonSHA256: digestText("exit")})
	state, _, _ = mustApply(t, state, genesis, exit)
	presentAgain := makeTransition(t, state, participantID, keys.participant, opPresentForm, state.ActiveLocusID, PresentFormPayload{
		LocalityReference: "PARTICIPANT-SINGLE-USE", SourceReference: "CONTINUITY",
		SourceSHA256: digestText("PARTICIPANT-SINGLE-USE source"), NucleusVersion: "2",
		NucleusSHA256: digestText("PARTICIPANT-SINGLE-USE nucleus 2"), FormID: formID, FormSHA256: formSHA256,
		StateCommitment: stateCommitment, MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	})
	state, _, presentationAgainID := mustApply(t, state, genesis, presentAgain)
	admitAgain := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, state.ActiveLocusID, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationAgainID, Disposition: "ADMIT", ReasonSHA256: digestText("admit again"),
		WorkUnits: 5, ResolutionUnits: 1, OfferExpiresAt: 4000,
	})
	state, _, _ = mustApply(t, state, genesis, admitAgain)
	reenter := makeTransition(t, state, participantID, keys.participant, opReenter, state.ActiveLocusID, ReenterPayload{
		PresentationID: presentationAgainID, PriorEntryTransitionID: entryID, DepartureCheckpointID: checkpointID,
		ResidueID: "", Passage: []ContinuityStep{},
	})
	state, _, secondEntryID := mustApply(t, state, genesis, reenter)
	secondCheckpoint := makeTransition(t, state, participantID, keys.participant, opCheckpointDeparture, state.ActiveLocusID, CheckpointDeparturePayload{
		EntryTransitionID: secondEntryID, PresentationID: presentationAgainID, FromStateCommitment: stateCommitment,
		Passage: []ContinuityStep{}, DepartureStateCommitment: stateCommitment,
	})
	state, _, secondCheckpointID := mustApply(t, state, genesis, secondCheckpoint)
	secondExit := makeTransition(t, state, participantID, keys.participant, opExit, state.ActiveLocusID, ExitPayload{DepartureCheckpointID: secondCheckpointID, ReasonSHA256: digestText("second exit")})
	state, _, _ = mustApply(t, state, genesis, secondExit)
	thirdPresent := makeTransition(t, state, participantID, keys.participant, opPresentForm, state.ActiveLocusID, PresentFormPayload{
		LocalityReference: "PARTICIPANT-SINGLE-USE", SourceReference: "CONTINUITY",
		SourceSHA256: digestText("PARTICIPANT-SINGLE-USE source"), NucleusVersion: "3",
		NucleusSHA256: digestText("PARTICIPANT-SINGLE-USE nucleus 3"), FormID: formID, FormSHA256: formSHA256,
		StateCommitment: stateCommitment, MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	})
	state, _, thirdPresentationID := mustApply(t, state, genesis, thirdPresent)
	thirdAdmit := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, state.ActiveLocusID, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: thirdPresentationID, Disposition: "ADMIT", ReasonSHA256: digestText("third admit"),
		WorkUnits: 5, ResolutionUnits: 1, OfferExpiresAt: 4000,
	})
	state, _, _ = mustApply(t, state, genesis, thirdAdmit)
	branch := makeTransition(t, state, participantID, keys.participant, opReenter, state.ActiveLocusID, ReenterPayload{
		PresentationID: thirdPresentationID, PriorEntryTransitionID: entryID, DepartureCheckpointID: checkpointID,
		ResidueID: "", Passage: []ContinuityStep{},
	})
	if _, _, err := applyToClone(state, genesis, branch); err == nil || !strings.Contains(err.Error(), "sealed, unconsumed") {
		t.Fatalf("consumed checkpoint was reused to branch continuity, got %v", err)
	}
}

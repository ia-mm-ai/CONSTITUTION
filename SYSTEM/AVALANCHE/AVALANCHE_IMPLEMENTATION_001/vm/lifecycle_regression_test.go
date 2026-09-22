package main

// TestAcceptedEighteenTransitionLifecycleRegression replays the exact
// eighteen-transition lifecycle accepted by the first source crossing
// (main@5e09d1f) against the canonical PRESENCE_AVALANCHE_VM_001
// implementation. It is a regression scenario rerun on the current tree, not
// transferred predecessor evidence:
//
//	BOUND → DECLARE_CAPACITY → PULSE → PRESENT_FORM → GATE_DISPOSITION →
//	ENTER → OBSERVE_CROSSING → MATTER_DISPOSITION → RECORD_EMERGENCE →
//	CORRECT → CHECKPOINT_DEPARTURE → EXIT → CLOSE → BOUND → PRESENT_FORM →
//	GATE_DISPOSITION → REENTER → INCORPORATE_ADDRESSED_RESIDUE

import (
	"crypto/ed25519"
	"testing"
)

func TestAcceptedEighteenTransitionLifecycleRegression(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	hostID := genesis.Locality.ID
	participantID := operationalActorID(keys.participant.Public().(ed25519.PublicKey))
	const (
		firstLocus  = "LOCUS-REGRESSION-001"
		returnLocus = "LOCUS-REGRESSION-002"
	)
	expectedSequence := []string{
		opBound, opDeclareCapacity, opPulse, opPresentForm, opGateDisposition,
		opEnter, opObserveCrossing, opMatterDisposition, opRecordEmergence,
		opCorrect, opCheckpointDeparture, opExit, opClose, opBound,
		opPresentForm, opGateDisposition, opReenter, opIncorporateResidue,
	}
	acceptedSequence := make([]string, 0, len(expectedSequence))
	accept := func(transition *Transition) (*Receipt, string) {
		t.Helper()
		next, receipt, id := mustApply(t, state, genesis, transition)
		state = next
		acceptedSequence = append(acceptedSequence, receipt.Operation)
		return receipt, id
	}

	// 1 BOUND
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opBound, firstLocus, BoundPayload{
		LocusID: firstLocus, PurposeSHA256: digestText("regression lifecycle purpose"),
		ClosureConditionSHA256: digestText("regression lifecycle closure"), CapacityCeilingUnits: 100,
	}))
	// 2 DECLARE_CAPACITY (body-local while a locus is active)
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opDeclareCapacity, "", DeclareCapacityPayload{
		ActualUnits: 100, ResourceCommitmentSHA256: digestText("regression resources"),
		BasisSHA256: digestText("regression capacity basis"),
	}))
	// 3 PULSE
	carrierSet := digestText("regression carrier set")
	currentness, err := state.currentnessCommitment(int64(state.Revision+1000), carrierSet)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opPulse, "", PulsePayload{
		CurrentnessCommitmentSHA256: currentness, CarrierSetSHA256: carrierSet,
	}))
	// 4 PRESENT_FORM
	initialParticipantState := digestText("regression participant state")
	_, presentationID := accept(makeTransition(t, state, participantID, keys.participant, opPresentForm, firstLocus, PresentFormPayload{
		LocalityReference: "DERIVED-LOCALITY-REGRESSION-001", SourceReference: "CONTINUITY",
		SourceSHA256: digestText("regression participant source"), NucleusVersion: "1",
		NucleusSHA256: digestText("regression nucleus 1"), FormID: formID, FormSHA256: formSHA256,
		StateCommitment: initialParticipantState, MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	}))
	// 5 GATE_DISPOSITION
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opGateDisposition, firstLocus, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationID, Disposition: "ADMIT",
		ReasonSHA256: digestText("regression admission"), WorkUnits: 10, ResolutionUnits: 2, OfferExpiresAt: 4000,
	}))
	// 6 ENTER
	_, entryID := accept(makeTransition(t, state, participantID, keys.participant, opEnter, firstLocus, EnterPayload{PresentationID: presentationID}))
	// 7 OBSERVE_CROSSING
	crossingReceipt, crossingID := accept(makeTransition(t, state, participantID, keys.participant, opObserveCrossing, firstLocus, ObserveCrossingPayload{
		MatterID: "REGRESSION-MATTER-001", ContentSHA256: digestText("regression matter content"),
		MediaType: "application/octet-stream", Claim: "bounded regression claim",
	}))
	if !contains(crossingReceipt.NonEffects, "NO_ADMISSION_BY_CROSSING") {
		t.Fatal("crossing receipt omitted its admission ceiling")
	}
	// 8 MATTER_DISPOSITION
	_, _ = accept(makeTransition(t, state, participantID, keys.participant, opMatterDisposition, firstLocus, MatterDispositionPayload{
		MatterID: "REGRESSION-MATTER-001", Disposition: "ADMIT", ReasonSHA256: digestText("regression matter admission"),
	}))
	// 9 RECORD_EMERGENCE
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opRecordEmergence, firstLocus, RecordEmergencePayload{
		EmergenceID: "REGRESSION-EMERGENCE-001", ContributorIDs: []string{participantID},
		MatterIDs: []string{"REGRESSION-MATTER-001"}, Kind: "LOCAL_TEST",
		DescriptionSHA256: digestText("regression emergence"),
	}))
	// 10 CORRECT (target-scoped to the exact crossing transition)
	_, _ = accept(makeTransition(t, state, participantID, keys.participant, opCorrect, firstLocus, CorrectPayload{
		TargetTransitionID: crossingID, ReplacementCommitment: digestText("regression corrected commitment"),
		ReasonSHA256: digestText("regression append-only correction"),
	}))
	// 11 CHECKPOINT_DEPARTURE
	departureDelta := digestText("regression departure delta")
	departureState, err := continuitySuccessorCommitment(participantID, initialParticipantState, departureDelta, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, departureCheckpointID := accept(makeTransition(t, state, participantID, keys.participant, opCheckpointDeparture, firstLocus, CheckpointDeparturePayload{
		EntryTransitionID: entryID, PresentationID: presentationID, FromStateCommitment: initialParticipantState,
		Passage:                  []ContinuityStep{{DeltaSHA256: departureDelta, SuccessorStateCommitment: departureState}},
		DepartureStateCommitment: departureState,
	}))
	// 12 EXIT
	_, _ = accept(makeTransition(t, state, participantID, keys.participant, opExit, firstLocus, ExitPayload{
		DepartureCheckpointID: departureCheckpointID, ReasonSHA256: digestText("regression lawful exit"),
	}))
	// 13 CLOSE
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opClose, firstLocus, ClosePayload{
		ClosureBasisSHA256: digestText("regression closure basis"),
	}))
	residue := state.Loci[firstLocus].Residues[participantID]
	if residue.DepartureCheckpointID != departureCheckpointID {
		t.Fatalf("closure residue lost its departure binding: %+v", residue)
	}
	// 14 BOUND (return locus)
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opBound, returnLocus, BoundPayload{
		LocusID: returnLocus, PurposeSHA256: digestText("regression return purpose"),
		ClosureConditionSHA256: digestText("regression return closure"), CapacityCeilingUnits: 100,
	}))
	// 15 PRESENT_FORM (renewed)
	returnDelta := digestText("regression delta while absent")
	returnState, err := continuitySuccessorCommitment(participantID, departureState, returnDelta, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, renewedPresentationID := accept(makeTransition(t, state, participantID, keys.participant, opPresentForm, returnLocus, PresentFormPayload{
		LocalityReference: "DERIVED-LOCALITY-REGRESSION-001", SourceReference: "CONTINUITY",
		SourceSHA256: digestText("regression participant source"), NucleusVersion: "2",
		NucleusSHA256: digestText("regression nucleus 2"), FormID: formID, FormSHA256: formSHA256,
		StateCommitment: returnState, MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	}))
	// 16 GATE_DISPOSITION (renewed admission)
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opGateDisposition, returnLocus, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: renewedPresentationID, Disposition: "ADMIT",
		ReasonSHA256: digestText("regression renewed admission"), WorkUnits: 12, ResolutionUnits: 3, OfferExpiresAt: 4000,
	}))
	// 17 REENTER (carried state, consumes checkpoint, addresses exact residue)
	reentryReceipt, reentryID := accept(makeTransition(t, state, participantID, keys.participant, opReenter, returnLocus, ReenterPayload{
		PresentationID: renewedPresentationID, PriorEntryTransitionID: entryID,
		DepartureCheckpointID: departureCheckpointID, ResidueID: residue.ResidueID,
		Passage: []ContinuityStep{{DeltaSHA256: returnDelta, SuccessorStateCommitment: returnState}},
	}))
	if state.EntryHistory[reentryID].IngressMode != "RENEWED_ENTRY" ||
		state.DepartureCheckpoints[departureCheckpointID].Status != "CONSUMED" {
		t.Fatal("re-entry did not carry state through the consumed departure checkpoint")
	}
	if !contains(reentryReceipt.NonEffects, "NO_RESTORATION_OR_RESUMPTION_OF_PRIOR_LOCUS") {
		t.Fatal("re-entry receipt omitted the non-restoration boundary")
	}
	// 18 INCORPORATE_ADDRESSED_RESIDUE (explicit local incorporation of a
	// residue addressed to this locality; never implied by closure)
	envelope := residueEnvelope{
		Schema: residueSchema, AddressedToLocalityID: hostID, SourceLocalityID: "EXTERNAL-LOCALITY-REGRESSION-001",
		SourceLocusID: "EXTERNAL-LOCUS-REGRESSION-001", SourceStateCommitment: digestText("external regression source state"),
		ClosureTransitionID: digestText("external regression closure"), ParticipantID: "EXTERNAL-PARTICIPANT-REGRESSION-001",
		PresentationID: digestText("external regression presentation"), EntryTransitionID: digestText("external regression entry"),
		DepartureCheckpointID:    digestText("external regression departure checkpoint"),
		DepartureStateCommitment: digestText("external regression departure state"),
		Effect:                   residueEffect,
	}
	envelopeSHA, envelopeID, err := residueCommitments(envelope)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = accept(makeTransition(t, state, hostID, keys.host, opIncorporateResidue, "", IncorporateResiduePayload{
		ResidueID: envelopeID, ResidueSHA256: envelopeSHA, Schema: envelope.Schema,
		AddressedToLocalityID: envelope.AddressedToLocalityID, SourceLocalityID: envelope.SourceLocalityID,
		SourceLocusID: envelope.SourceLocusID, SourceStateCommitment: envelope.SourceStateCommitment,
		ClosureTransitionID: envelope.ClosureTransitionID, ParticipantID: envelope.ParticipantID,
		PresentationID: envelope.PresentationID, EntryTransitionID: envelope.EntryTransitionID,
		DepartureCheckpointID: envelope.DepartureCheckpointID, DepartureStateCommitment: envelope.DepartureStateCommitment,
		Effect: envelope.Effect,
	}))

	if state.Revision != uint64(len(expectedSequence)) {
		t.Fatalf("expected final revision %d, got %d", len(expectedSequence), state.Revision)
	}
	if len(acceptedSequence) != len(expectedSequence) {
		t.Fatalf("expected %d accepted transitions, got %d", len(expectedSequence), len(acceptedSequence))
	}
	for index, operation := range expectedSequence {
		if acceptedSequence[index] != operation {
			t.Fatalf("transition %d: expected %s, accepted %s", index+1, operation, acceptedSequence[index])
		}
	}
	if _, exists := state.IncorporatedResidues[envelopeID]; !exists {
		t.Fatal("explicit addressed residue incorporation was not recorded")
	}
	if err := state.validateLoaded(); err != nil {
		t.Fatalf("final state commitment invalid: %v", err)
	}
}

package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"strings"
	"testing"
)

func presentParticipant(t *testing.T, state *RuntimeState, genesis *Genesis, key ed25519.PrivateKey, localityReference string) (*RuntimeState, string, string) {
	t.Helper()
	actorID := operationalActorID(key.Public().(ed25519.PublicKey))
	presentation := makeTransition(t, state, actorID, key, opPresentForm, state.ActiveLocusID, PresentFormPayload{
		LocalityReference: localityReference, SourceReference: "CONTINUITY",
		SourceSHA256: digestText(localityReference + " source"), NucleusVersion: "1",
		NucleusSHA256: digestText(localityReference + " nucleus"), FormID: formID, FormSHA256: formSHA256,
		StateCommitment: digestText(localityReference + " state"), MediumCapabilities: testMediumCapabilities(),
		EffectCeiling: testEffectCeiling(),
	})
	next, receipt, _ := mustApply(t, state, genesis, presentation)
	return next, actorID, receipt.TransitionID
}

func TestLivingCapacityReservesFullLifecycleAtEnterAndReleasesAtExit(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	bound := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-CAPACITY-001", BoundPayload{
		LocusID: "LOCUS-CAPACITY-001", PurposeSHA256: digestText("capacity purpose"),
		ClosureConditionSHA256: digestText("capacity closure"), CapacityCeilingUnits: 60,
	})
	state, _, _ = mustApply(t, state, genesis, bound)
	state, participantID, presentationID := presentParticipant(t, state, genesis, keys.participant, "PARTICIPANT-CAPACITY-001")

	admit := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, state.ActiveLocusID, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationID, Disposition: "ADMIT",
		ReasonSHA256: digestText("whole lifecycle fits"), WorkUnits: 40, ResolutionUnits: 10, OfferExpiresAt: 2000,
	})
	state, admissionReceipt, _ := mustApply(t, state, genesis, admit)
	if admissionReceipt.Capacity.AvailableUnits != 50 {
		t.Fatalf("admission consumed capacity: available=%d", admissionReceipt.Capacity.AvailableUnits)
	}
	secondKey := deterministicKey(5)
	state, secondID, secondPresentationID := presentParticipant(t, state, genesis, secondKey, "PARTICIPANT-CAPACITY-002")
	secondAdmit := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, state.ActiveLocusID, GateDispositionPayload{
		ParticipantID: secondID, PresentationID: secondPresentationID, Disposition: "ADMIT",
		ReasonSHA256: digestText("offer does not guarantee entry"), WorkUnits: 1, ResolutionUnits: 1, OfferExpiresAt: 2000,
	})
	state, _, _ = mustApply(t, state, genesis, secondAdmit)
	enter := makeTransition(t, state, participantID, keys.participant, opEnter, state.ActiveLocusID, EnterPayload{PresentationID: presentationID})
	state, entryReceipt, entryID := mustApply(t, state, genesis, enter)
	if entryReceipt.Capacity.UnresolvedUnits != 40 || entryReceipt.Capacity.CorrectionEgressUnits != 20 || entryReceipt.Capacity.AvailableUnits != 0 {
		t.Fatalf("unexpected post-entry capacity account: %+v", entryReceipt.Capacity)
	}
	if state.Body.Posture != "ACTIVE_P1" || state.Body.PresenceCount != 1 {
		t.Fatalf("entry did not create current presence: %+v", state.Body)
	}
	secondEnter := makeTransition(t, state, secondID, secondKey, opEnter, state.ActiveLocusID, EnterPayload{PresentationID: secondPresentationID})
	if _, _, err := applyToClone(state, genesis, secondEnter); err == nil || !strings.Contains(err.Error(), "presently available") {
		t.Fatalf("entry over living capacity should fail, got %v", err)
	}

	degrade := makeTransition(t, state, genesis.Locality.ID, keys.host, opDeclareCapacity, "", DeclareCapacityPayload{
		ActualUnits: 20, ResourceCommitmentSHA256: digestText("degraded carriers"), BasisSHA256: digestText("observed degradation"),
	})
	state, degradedReceipt, _ := mustApply(t, state, genesis, degrade)
	if state.Body.Posture != "HOLD_CAPACITY_DEFICIT" || degradedReceipt.Capacity.DeficitUnits != 40 {
		t.Fatalf("capacity loss should force HOLD: body=%+v capacity=%+v", state.Body, degradedReceipt.Capacity)
	}
	blockedCrossing := makeTransition(t, state, participantID, keys.participant, opObserveCrossing, state.ActiveLocusID, ObserveCrossingPayload{
		MatterID: "MATTER-BLOCKED-BY-CAPACITY", ContentSHA256: digestText("must hold"),
		MediaType: "application/locality-test", Claim: "No new consequence during deficit.",
	})
	if _, _, err := applyToClone(state, genesis, blockedCrossing); err == nil || !strings.Contains(err.Error(), "places this motion in HOLD") {
		t.Fatalf("capacity deficit admitted new consequence: %v", err)
	}
	departureState := digestText("PARTICIPANT-CAPACITY-001 state")
	checkpoint := makeTransition(t, state, participantID, keys.participant, opCheckpointDeparture, state.ActiveLocusID, CheckpointDeparturePayload{
		EntryTransitionID: entryID, PresentationID: presentationID, FromStateCommitment: departureState,
		Passage: []ContinuityStep{}, DepartureStateCommitment: departureState,
	})
	state, _, checkpointID := mustApply(t, state, genesis, checkpoint)
	exit := makeTransition(t, state, participantID, keys.participant, opExit, state.ActiveLocusID, ExitPayload{DepartureCheckpointID: checkpointID, ReasonSHA256: digestText("lawful exit remains open")})
	state, exitReceipt, _ := mustApply(t, state, genesis, exit)
	if exitReceipt.Capacity.UnresolvedUnits != 0 || exitReceipt.Capacity.AvailableUnits != 10 || state.Body.Posture != "DORMANT_P0" {
		t.Fatalf("exit did not release lifecycle reservation: body=%+v capacity=%+v", state.Body, exitReceipt.Capacity)
	}
}

func TestPulseCommitsCurrentStateWithoutInventingPresence(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	carrierSet := digestText("carrier set")
	observedAt := int64(1000)
	commitment, err := state.currentnessCommitment(observedAt, carrierSet)
	if err != nil {
		t.Fatal(err)
	}
	pulse := makeTransition(t, state, genesis.Locality.ID, keys.host, opPulse, "", PulsePayload{
		CurrentnessCommitmentSHA256: commitment, CarrierSetSHA256: carrierSet,
	})
	state, receipt, pulseID := mustApply(t, state, genesis, pulse)
	if state.LastPulseID != pulseID || receipt.PresenceCount != 0 || receipt.Posture != "DORMANT_P0" {
		t.Fatalf("pulse fabricated presence or failed to bind currentness: %+v", receipt)
	}
	if state.Pulses[pulseID].Effect != "SIGNED_LOCAL_CURRENTNESS_NOT_EXTERNAL_LIVENESS_PROOF" {
		t.Fatal("pulse exceeded its evidence ceiling")
	}
	bad := makeTransition(t, state, genesis.Locality.ID, keys.host, opPulse, "", PulsePayload{
		CurrentnessCommitmentSHA256: digestText("wrong state"), CarrierSetSHA256: carrierSet,
	})
	if _, _, err := applyToClone(state, genesis, bad); err == nil || !strings.Contains(err.Error(), "exact pre-transition state") {
		t.Fatalf("unbound pulse should fail, got %v", err)
	}
}

func TestExpiredAdmissionOfferNeedsExplicitReclamation(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	bound := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-OFFER-001", BoundPayload{
		LocusID: "LOCUS-OFFER-001", PurposeSHA256: digestText("offer purpose"),
		ClosureConditionSHA256: digestText("offer closure"), CapacityCeilingUnits: 100,
	})
	state, _, _ = mustApply(t, state, genesis, bound)
	state, participantID, presentationID := presentParticipant(t, state, genesis, keys.participant, "PARTICIPANT-OFFER-001")
	admit := makeTransition(t, state, genesis.Locality.ID, keys.host, opGateDisposition, state.ActiveLocusID, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationID, Disposition: "ADMIT",
		ReasonSHA256: digestText("bounded offer"), WorkUnits: 10, ResolutionUnits: 2, OfferExpiresAt: 1200,
	})
	state, _, offerID := mustApply(t, state, genesis, admit)
	reclaimPayload := ReclaimAdmissionOfferPayload{
		ParticipantID: participantID, PresentationID: presentationID, OfferTransitionID: offerID,
		ReasonSHA256: digestText("offer expired"),
	}
	early := makeTransition(t, state, genesis.Locality.ID, keys.host, opReclaimOffer, state.ActiveLocusID, reclaimPayload)
	if _, _, err := applyToClone(state, genesis, early); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("unexpired offer was reclaimed: %v", err)
	}
	lateUnsigned := early.Unsigned
	lateUnsigned.ObservedAt = 1200
	late, err := signTransition(lateUnsigned, keys.host)
	if err != nil {
		t.Fatal(err)
	}
	state, receipt, _ := mustApply(t, state, genesis, late)
	gate := state.Loci[state.ActiveLocusID].Gates[participantID]
	if gate.OfferStatus != "EXPIRED" || receipt.Capacity.AvailableUnits != 90 {
		t.Fatalf("offer reclamation altered capacity or failed to expire offer: gate=%+v capacity=%+v", gate, receipt.Capacity)
	}
}

func registerContinuityAuthority(t *testing.T, state *RuntimeState, genesis *Genesis, hostKey, continuityKey ed25519.PrivateKey, keyID string) *RuntimeState {
	t.Helper()
	publicKey := continuityKey.Public().(ed25519.PublicKey)
	transition := makeTransition(t, state, genesis.Locality.ID, hostKey, opRegisterAuthority, "", RegisterContinuityAuthorityPayload{
		KeyID: keyID, PublicKey: hex.EncodeToString(publicKey), Capabilities: append([]string(nil), continuityAuthorityCapabilities...),
		MandateSHA256: digestText(keyID + " mandate"),
	})
	next, _, _ := mustApply(t, state, genesis, transition)
	return next
}

func TestFormationAuthorityExhaustsAndAttestedSuccessorFreezesPredecessor(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	state = registerContinuityAuthority(t, state, genesis, keys.host, keys.continuity1, "continuity-001")
	prematureExhaust := makeTransition(t, state, genesis.Locality.ID, keys.host, opExhaustFormation, "", ExhaustFormationAuthorityPayload{BasisSHA256: digestText("too early")})
	if _, _, err := applyToClone(state, genesis, prematureExhaust); err == nil || !strings.Contains(err.Error(), "requires 2") {
		t.Fatalf("formation authority exhausted without minimum continuity: %v", err)
	}
	state = registerContinuityAuthority(t, state, genesis, keys.host, keys.continuity2, "continuity-002")
	exhaust := makeTransition(t, state, genesis.Locality.ID, keys.host, opExhaustFormation, "", ExhaustFormationAuthorityPayload{BasisSHA256: digestText("formation complete")})
	state, _, _ = mustApply(t, state, genesis, exhaust)
	if state.Authorities[genesis.Locality.ID].Status != "EXHAUSTED" {
		t.Fatal("formation authority remained active")
	}
	formerHostMotion := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-FORBIDDEN", BoundPayload{
		LocusID: "LOCUS-FORBIDDEN", PurposeSHA256: digestText("forbidden"), ClosureConditionSHA256: digestText("forbidden"), CapacityCeilingUnits: 50,
	})
	if _, _, err := applyToClone(state, genesis, formerHostMotion); err == nil || !strings.Contains(err.Error(), "active authority capability") {
		t.Fatalf("exhausted formation authority still moved the body: %v", err)
	}

	continuity1ID := operationalActorID(keys.continuity1.Public().(ed25519.PublicKey))
	continuity2ID := operationalActorID(keys.continuity2.Public().(ed25519.PublicKey))
	proposalPayload := ProposeSuccessorPayload{
		ProposalID: "LOCALITY-VM-004-CANDIDATE-001", Protocol: "LOCALITY_VM_004", Version: "4.0.0",
		VMID: "BQeuw4nSSyB6mjvjgZXD4r4tQdpXfiUsKAC2Vyb3setqnHsFJ", RuntimeSHA256: digestText("successor runtime"),
		MigrationSHA256: digestText("migration"), InvariantSetSHA256: digestText("invariants"),
		SourceReference: "CONSTITUTION-SUCCESSOR-001", SourceSHA256: digestText("successor source"),
		ActivationMode: "FREEZE_FOR_EXTERNAL_BINARY_HANDOFF",
	}
	propose := makeTransition(t, state, continuity1ID, keys.continuity1, opProposeSuccessor, "", proposalPayload)
	state, _, _ = mustApply(t, state, genesis, propose)
	commitment := state.SuccessorProposals[proposalPayload.ProposalID].ProposalCommitment
	attest1 := makeTransition(t, state, continuity1ID, keys.continuity1, opAttestSuccessor, "", AttestSuccessorPayload{
		ProposalID: proposalPayload.ProposalID, ProposalCommitment: commitment, BasisSHA256: digestText("continuity one attests"),
	})
	state, _, _ = mustApply(t, state, genesis, attest1)
	prematureActivation := makeTransition(t, state, continuity1ID, keys.continuity1, opActivateSuccessor, "", ActivateSuccessorPayload{
		ProposalID: proposalPayload.ProposalID, ProposalCommitment: commitment,
	})
	if _, _, err := applyToClone(state, genesis, prematureActivation); err == nil || !strings.Contains(err.Error(), "requires 2") {
		t.Fatalf("successor activated below attestation threshold: %v", err)
	}
	attest2 := makeTransition(t, state, continuity2ID, keys.continuity2, opAttestSuccessor, "", AttestSuccessorPayload{
		ProposalID: proposalPayload.ProposalID, ProposalCommitment: commitment, BasisSHA256: digestText("continuity two attests"),
	})
	state, _, _ = mustApply(t, state, genesis, attest2)
	activate := makeTransition(t, state, continuity1ID, keys.continuity1, opActivateSuccessor, "", ActivateSuccessorPayload{
		ProposalID: proposalPayload.ProposalID, ProposalCommitment: commitment,
	})
	state, receipt, _ := mustApply(t, state, genesis, activate)
	if state.SuccessorBoundary == nil || receipt.Posture != "SUCCESSION_COMMITTED" {
		t.Fatalf("successor boundary was not activated: %+v", state.SuccessorBoundary)
	}
	frozenMotion := makeTransition(t, state, continuity1ID, keys.continuity1, opDeclareCapacity, "", DeclareCapacityPayload{
		ActualUnits: 90, ResourceCommitmentSHA256: digestText("after freeze"), BasisSHA256: digestText("must fail"),
	})
	if _, _, err := applyToClone(state, genesis, frozenMotion); err == nil || !strings.Contains(err.Error(), "frozen") {
		t.Fatalf("predecessor accepted motion after succession freeze: %v", err)
	}
}

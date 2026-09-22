package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"strings"
	"testing"
)

func TestLocalityLifecyclePreservesIrreducibleDistinctions(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	const (
		hostID  = "PRESENCE-RUNTIME-TEST"
		locusID = "LOCUS-001"
	)
	participantID := operationalActorID(keys.participant.Public().(ed25519.PublicKey))

	bound := makeTransition(t, state, hostID, keys.host, opBound, locusID, BoundPayload{
		LocusID: locusID, PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	state, _, _ = mustApply(t, state, genesis, bound)

	squattingPresentation := makeTransition(t, state, "DERIVED-LOCALITY-001", keys.participant, opPresentForm, locusID, PresentFormPayload{
		LocalityReference: "DERIVED-LOCALITY-001",
		SourceReference:   "LIVE-STATE-1822-17092026-MM22",
		SourceSHA256:      digestText("participant source"), NucleusVersion: "1", NucleusSHA256: digestText("nucleus"),
		FormID: formID, FormSHA256: formSHA256, StateCommitment: digestText("participant state"),
		MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	})
	if _, _, err := applyToClone(state, genesis, squattingPresentation); err == nil || !strings.Contains(err.Error(), "derived") {
		t.Fatalf("arbitrary operational actor ID should fail, got %v", err)
	}

	presentation := makeTransition(t, state, participantID, keys.participant, opPresentForm, locusID, PresentFormPayload{
		LocalityReference: "DERIVED-LOCALITY-001",
		SourceReference:   "LIVE-STATE-1822-17092026-MM22",
		SourceSHA256:      digestText("participant source"), NucleusVersion: "1", NucleusSHA256: digestText("nucleus"),
		FormID: formID, FormSHA256: formSHA256, StateCommitment: digestText("participant state"),
		MediumCapabilities: testMediumCapabilities(), EffectCeiling: testEffectCeiling(),
	})
	state, presentationReceipt, _ := mustApply(t, state, genesis, presentation)
	locus := state.Loci[locusID]
	if _, exists := locus.Gates[participantID]; exists {
		t.Fatal("presentation silently created a gate disposition")
	}
	if _, exists := locus.Entries[participantID]; exists {
		t.Fatal("presentation silently created entry")
	}

	entryBeforeAdmission := makeTransition(t, state, participantID, keys.participant, opEnter, locusID, EnterPayload{PresentationID: presentationReceipt.TransitionID})
	if _, _, err := applyToClone(state, genesis, entryBeforeAdmission); err == nil || !strings.Contains(err.Error(), "ADMIT") {
		t.Fatalf("entry without admission should fail, got %v", err)
	}

	refuse := makeTransition(t, state, hostID, keys.host, opGateDisposition, locusID, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationReceipt.TransitionID,
		Disposition: "REFUSE", ReasonSHA256: digestText("not yet"),
	})
	state, _, _ = mustApply(t, state, genesis, refuse)
	entryAfterRefusal := makeTransition(t, state, participantID, keys.participant, opEnter, locusID, EnterPayload{PresentationID: presentationReceipt.TransitionID})
	if _, _, err := applyToClone(state, genesis, entryAfterRefusal); err == nil {
		t.Fatal("entry after REFUSE should fail")
	}

	admit := makeTransition(t, state, hostID, keys.host, opGateDisposition, locusID, GateDispositionPayload{
		ParticipantID: participantID, PresentationID: presentationReceipt.TransitionID,
		Disposition: "ADMIT", ReasonSHA256: digestText("admit"), WorkUnits: 10, ResolutionUnits: 2, OfferExpiresAt: 4000,
	})
	state, _, _ = mustApply(t, state, genesis, admit)
	if _, exists := state.Loci[locusID].Entries[participantID]; exists {
		t.Fatal("admission silently created entry")
	}

	enter := makeTransition(t, state, participantID, keys.participant, opEnter, locusID, EnterPayload{PresentationID: presentationReceipt.TransitionID})
	state, _, entryID := mustApply(t, state, genesis, enter)

	crossing := makeTransition(t, state, participantID, keys.participant, opObserveCrossing, locusID, ObserveCrossingPayload{
		MatterID: "MATTER-001", ContentSHA256: digestText("matter"), MediaType: "application/locality-test", Claim: "A bounded claim, not truth.",
	})
	state, _, crossingID := mustApply(t, state, genesis, crossing)
	if len(state.Loci[locusID].Matter["MATTER-001"].Dispositions) != 0 {
		t.Fatal("crossing silently admitted matter")
	}

	prematureEmergence := makeTransition(t, state, hostID, keys.host, opRecordEmergence, locusID, RecordEmergencePayload{
		EmergenceID: "EMERGENCE-001", ContributorIDs: []string{participantID}, MatterIDs: []string{"MATTER-001"},
		Kind: "TEST", DescriptionSHA256: digestText("emergence"),
	})
	if _, _, err := applyToClone(state, genesis, prematureEmergence); err == nil || !strings.Contains(err.Error(), "not locally admitted") {
		t.Fatalf("emergence before local matter admission should fail, got %v", err)
	}

	matterAdmit := makeTransition(t, state, participantID, keys.participant, opMatterDisposition, locusID, MatterDispositionPayload{
		MatterID: "MATTER-001", Disposition: "ADMIT", ReasonSHA256: digestText("participant admits matter"),
	})
	state, _, _ = mustApply(t, state, genesis, matterAdmit)
	emergence := makeTransition(t, state, hostID, keys.host, opRecordEmergence, locusID, RecordEmergencePayload{
		EmergenceID: "EMERGENCE-001", ContributorIDs: []string{participantID}, MatterIDs: []string{"MATTER-001"},
		Kind: "TEST", DescriptionSHA256: digestText("emergence"),
	})
	state, emergenceReceipt, _ := mustApply(t, state, genesis, emergence)
	if emergenceReceipt.Effect != "RECORDS_FIELD_LOCAL_EMERGENCE_ONLY" {
		t.Fatalf("unexpected emergence effect %q", emergenceReceipt.Effect)
	}

	correction := makeTransition(t, state, participantID, keys.participant, opCorrect, locusID, CorrectPayload{
		TargetTransitionID: crossingID, ReplacementCommitment: digestText("corrected matter description"), ReasonSHA256: digestText("correction reason"),
	})
	state, _, correctionID := mustApply(t, state, genesis, correction)
	if _, exists := state.Events[crossingID]; !exists {
		t.Fatal("correction erased predecessor event")
	}
	if got := state.Corrections[crossingID][0].CorrectionTransitionID; got != correctionID {
		t.Fatalf("correction link = %s", got)
	}

	departureCommitment, err := continuitySuccessorCommitment(participantID, digestText("participant state"), digestText("participant departure delta"), 1)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := makeTransition(t, state, participantID, keys.participant, opCheckpointDeparture, locusID, CheckpointDeparturePayload{
		EntryTransitionID: entryID, PresentationID: presentationReceipt.TransitionID,
		FromStateCommitment:      digestText("participant state"),
		Passage:                  []ContinuityStep{{DeltaSHA256: digestText("participant departure delta"), SuccessorStateCommitment: departureCommitment}},
		DepartureStateCommitment: departureCommitment,
	})
	state, _, checkpointID := mustApply(t, state, genesis, checkpoint)
	exit := makeTransition(t, state, participantID, keys.participant, opExit, locusID, ExitPayload{DepartureCheckpointID: checkpointID, ReasonSHA256: digestText("exit")})
	state, _, _ = mustApply(t, state, genesis, exit)
	closeTransition := makeTransition(t, state, hostID, keys.host, opClose, locusID, ClosePayload{ClosureBasisSHA256: digestText("closure basis")})
	state, closeReceipt, _ := mustApply(t, state, genesis, closeTransition)
	if state.ActiveLocusID != "" || state.Loci[locusID].Phase != "CLOSED" {
		t.Fatal("locus did not close")
	}
	residue, exists := state.Loci[locusID].Residues[participantID]
	if !exists {
		t.Fatal("closure did not create addressed residue")
	}
	if residue.AddressedToLocalityID != "DERIVED-LOCALITY-001" {
		t.Fatalf("residue addressed to %q", residue.AddressedToLocalityID)
	}
	if _, exists := state.IncorporatedResidues[residue.ResidueID]; exists {
		t.Fatal("closure silently incorporated residue")
	}
	if !contains(closeReceipt.NonEffects, "NO_LOCAL_INCORPORATION_BY_CLOSURE") {
		t.Fatal("closure receipt omits incorporation non-effect")
	}

	wrongAddressee := makeTransition(t, state, hostID, keys.host, opIncorporateResidue, "", IncorporateResiduePayload{
		ResidueID: residue.ResidueID, ResidueSHA256: residue.ResidueSHA256, Schema: residue.Schema,
		AddressedToLocalityID: residue.AddressedToLocalityID, SourceLocalityID: residue.SourceLocalityID,
		SourceLocusID: residue.SourceLocusID, SourceStateCommitment: residue.SourceStateCommitment,
		ClosureTransitionID: residue.ClosureTransitionID, ParticipantID: residue.ParticipantID,
		PresentationID: residue.PresentationID, EntryTransitionID: residue.EntryTransitionID,
		DepartureCheckpointID: residue.DepartureCheckpointID, DepartureStateCommitment: residue.DepartureStateCommitment,
		Effect: residue.Effect,
	})
	if _, _, err := applyToClone(state, genesis, wrongAddressee); err == nil || !strings.Contains(err.Error(), "not addressed") {
		t.Fatalf("wrong locality imported addressed residue: %v", err)
	}

	externalEnvelope := residueEnvelope{
		Schema: residueSchema, AddressedToLocalityID: hostID, SourceLocalityID: "EXTERNAL-LOCALITY-001",
		SourceLocusID: "EXTERNAL-LOCUS-001", SourceStateCommitment: digestText("external source state"),
		ClosureTransitionID: digestText("external close transition"), ParticipantID: "EXTERNAL-PARTICIPANT-001",
		PresentationID: digestText("external presentation"), EntryTransitionID: digestText("external entry"),
		DepartureCheckpointID: digestText("external departure checkpoint"), DepartureStateCommitment: digestText("external departure state"),
		Effect: residueEffect,
	}
	externalSHA, externalID, err := residueCommitments(externalEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	incorporate := makeTransition(t, state, hostID, keys.host, opIncorporateResidue, "", IncorporateResiduePayload{
		ResidueID: externalID, ResidueSHA256: externalSHA, Schema: externalEnvelope.Schema,
		AddressedToLocalityID: externalEnvelope.AddressedToLocalityID, SourceLocalityID: externalEnvelope.SourceLocalityID,
		SourceLocusID: externalEnvelope.SourceLocusID, SourceStateCommitment: externalEnvelope.SourceStateCommitment,
		ClosureTransitionID: externalEnvelope.ClosureTransitionID, ParticipantID: externalEnvelope.ParticipantID,
		PresentationID: externalEnvelope.PresentationID, EntryTransitionID: externalEnvelope.EntryTransitionID,
		DepartureCheckpointID: externalEnvelope.DepartureCheckpointID, DepartureStateCommitment: externalEnvelope.DepartureStateCommitment,
		Effect: externalEnvelope.Effect,
	})
	state, _, _ = mustApply(t, state, genesis, incorporate)
	if _, exists := state.IncorporatedResidues[externalID]; !exists {
		t.Fatal("explicit residue incorporation was not recorded")
	}
	if err := state.validateLoaded(); err != nil {
		t.Fatalf("final state commitment invalid: %v", err)
	}
}

func TestTransitionCanonicalSignatureAndUnknownFieldChecks(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	transition := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "LOCUS-001", BoundPayload{
		LocusID: "LOCUS-001", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	canonical, err := transition.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := parseTransition(canonical); err != nil {
		t.Fatalf("canonical signed transition rejected: %v", err)
	}
	pretty := new(bytes.Buffer)
	if err := json.Indent(pretty, canonical, "", "  "); err != nil {
		t.Fatal(err)
	}
	if _, _, err := parseTransition([]byte(pretty.String())); err == nil || !strings.Contains(err.Error(), "canonical") {
		t.Fatalf("non-canonical transition should fail, got %v", err)
	}

	tampered := *transition
	payload := BoundPayload{}
	if err := json.Unmarshal(transition.Unsigned.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	payload.PurposeSHA256 = digestText("tampered")
	tampered.Unsigned = transition.Unsigned
	tampered.Unsigned.Payload, _ = json.Marshal(payload)
	if err := tampered.ValidateSyntax(); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("tampered signed transition should fail signature validation, got %v", err)
	}

	withUnknown := append([]byte(nil), canonical[:len(canonical)-1]...)
	withUnknown = append(withUnknown, []byte(`,"extra":true}`)...)
	if _, _, err := parseTransition(withUnknown); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field should fail, got %v", err)
	}
}

func TestGenesisRejectsAuthoritylessFormationFossil(t *testing.T) {
	genesisBytes, genesis, _ := testGenesis(t)
	genesis.Locality.Authority.PublicKey = ""
	invalid, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseGenesis(invalid); err == nil {
		t.Fatal("authorityless genesis was accepted as runnable")
	}
	if _, err := parseGenesis(genesisBytes); err != nil {
		t.Fatalf("valid explicit-authority genesis rejected: %v", err)
	}
}

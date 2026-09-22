package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	residueSchema           = "LOCALITY_ADDRESSED_RESIDUE_003"
	residueEffect           = "AVAILABLE_FOR_ADDRESSEE_DECISION_ONLY"
	residueIDDomain         = "LOCALITY_RESIDUE_ID_003"
	continuityStepSchema    = "LOCALITY_PARTICIPANT_STATE_SUCCESSION_003"
	continuityPassageSchema = "LOCALITY_PARTICIPANT_STATE_PASSAGE_003"
)

type residueEnvelope struct {
	Schema                   string `json:"schema"`
	AddressedToLocalityID    string `json:"addressed_to_locality_id"`
	SourceLocalityID         string `json:"source_locality_id"`
	SourceLocusID            string `json:"source_locus_id"`
	SourceStateCommitment    string `json:"source_state_commitment"`
	ClosureTransitionID      string `json:"closure_transition_id"`
	ParticipantID            string `json:"participant_id"`
	PresentationID           string `json:"presentation_id"`
	EntryTransitionID        string `json:"entry_transition_id"`
	DepartureCheckpointID    string `json:"departure_checkpoint_id"`
	DepartureStateCommitment string `json:"departure_state_commitment"`
	Effect                   string `json:"effect"`
}

func (s *RuntimeState) applyOperation(genesis *Genesis, transition *Transition, transitionID string) error {
	u := &transition.Unsigned
	switch u.Operation {
	case opBound:
		return s.applyBound(u, transitionID)
	case opPresentForm:
		return s.applyPresentForm(genesis, u, transitionID)
	case opGateDisposition:
		return s.applyGateDisposition(genesis, u, transitionID)
	case opEnter:
		return s.applyEnter(u, transitionID)
	case opCheckpointDeparture:
		return s.applyCheckpointDeparture(u, transitionID)
	case opReenter:
		return s.applyReenter(u, transitionID)
	case opObserveCrossing:
		return s.applyObserveCrossing(u, transitionID)
	case opMatterDisposition:
		return s.applyMatterDisposition(u, transitionID)
	case opRecordEmergence:
		return s.applyRecordEmergence(u, transitionID)
	case opCorrect:
		return s.applyCorrection(u, transitionID)
	case opExit:
		return s.applyExit(u, transitionID)
	case opClose:
		return s.applyClose(u, transitionID)
	case opIncorporateResidue:
		return s.applyIncorporateResidue(u, transitionID)
	case opDeclareCapacity:
		return s.applyDeclareCapacity(u, transitionID)
	case opPulse:
		return s.applyPulse(u, transitionID)
	case opReclaimOffer:
		return s.applyReclaimAdmissionOffer(u, transitionID)
	case opRegisterAuthority:
		return s.applyRegisterContinuityAuthority(u, transitionID)
	case opExhaustFormation:
		return s.applyExhaustFormationAuthority(genesis, u, transitionID)
	case opProposeSuccessor:
		return s.applyProposeSuccessor(u, transitionID)
	case opAttestSuccessor:
		return s.applyAttestSuccessor(u, transitionID)
	case opActivateSuccessor:
		return s.applyActivateSuccessor(genesis, u, transitionID)
	default:
		return fmt.Errorf("unsupported operation %q", u.Operation)
	}
}

func (s *RuntimeState) applyBound(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capOpenLocus); err != nil {
		return err
	}
	if s.ActiveLocusID != "" {
		return errors.New("a locus is already active")
	}
	p, err := decodePayload[BoundPayload](u.Payload)
	if err != nil {
		return err
	}
	if p.LocusID != u.LocusID {
		return errors.New("BOUND payload and transition locus IDs differ")
	}
	if _, exists := s.Loci[p.LocusID]; exists {
		return errors.New("locus ID has already been used")
	}
	if p.CapacityCeilingUnits > s.Capacity.StructuralCeilingUnits || p.CapacityCeilingUnits <= s.Capacity.CorrectionEgressReserveUnits {
		return errors.New("locus capacity ceiling must fit the structural envelope and preserve correction/egress reserve")
	}
	s.Loci[p.LocusID] = &LocusState{
		LocusID:                p.LocusID,
		Phase:                  "OPEN",
		BoundTransitionID:      transitionID,
		PurposeSHA256:          p.PurposeSHA256,
		ClosureConditionSHA256: p.ClosureConditionSHA256,
		CapacityCeilingUnits:   p.CapacityCeilingUnits,
		Presentations:          make(map[string]PresentationState),
		Gates:                  make(map[string]GateState),
		Entries:                make(map[string]EntryState),
		Matter:                 make(map[string]MatterState),
		Emergence:              make(map[string]EmergenceState),
		Residues:               make(map[string]ResidueState),
	}
	s.ActiveLocusID = p.LocusID
	return nil
}

func (s *RuntimeState) applyPresentForm(genesis *Genesis, u *UnsignedTransition, transitionID string) error {
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	if participantPresent(locus, u.ActorID) {
		return errors.New("present participant cannot replace its form while entered")
	}
	p, err := decodePayload[PresentFormPayload](u.Payload)
	if err != nil {
		return err
	}
	if p.FormID != genesis.Profile.FormID || p.FormSHA256 != genesis.Profile.FormSHA256 {
		return errors.New("form does not match the genesis-selected LOCALITY profile")
	}
	for participantID, presentation := range locus.Presentations {
		if participantID != u.ActorID && presentation.LocalityReference == p.LocalityReference {
			return errors.New("locality reference is already presented in this locus")
		}
	}
	for _, required := range requiredMediumCapabilities {
		if !contains(p.MediumCapabilities, required) {
			return fmt.Errorf("presented form lacks local-medium capability %s", required)
		}
	}
	for _, ceiling := range genesis.Profile.EffectCeiling {
		if !contains(p.EffectCeiling, ceiling) {
			return fmt.Errorf("presented form omits genesis effect ceiling %s", ceiling)
		}
	}
	presentation := PresentationState{
		PresentationID:     transitionID,
		ParticipantID:      u.ActorID,
		LocalityReference:  p.LocalityReference,
		SourceReference:    p.SourceReference,
		SourceSHA256:       p.SourceSHA256,
		NucleusVersion:     p.NucleusVersion,
		NucleusSHA256:      p.NucleusSHA256,
		FormID:             p.FormID,
		FormSHA256:         p.FormSHA256,
		StateCommitment:    p.StateCommitment,
		MediumCapabilities: sortedCopy(p.MediumCapabilities),
		EffectCeiling:      sortedCopy(p.EffectCeiling),
	}
	locus.Presentations[u.ActorID] = presentation
	s.PresentationHistory[transitionID] = presentation
	delete(locus.Gates, u.ActorID)
	return nil
}

func (s *RuntimeState) applyGateDisposition(genesis *Genesis, u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capRegulateGate); err != nil {
		return err
	}
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	p, err := decodePayload[GateDispositionPayload](u.Payload)
	if err != nil {
		return err
	}
	presentation, exists := locus.Presentations[p.ParticipantID]
	if !exists || presentation.PresentationID != p.PresentationID {
		return errors.New("gate disposition does not address the participant's current presentation")
	}
	if participantPresent(locus, p.ParticipantID) {
		return errors.New("gate posture cannot be rewritten while participant is entered")
	}
	if p.Disposition == "ADMIT" {
		if p.OfferExpiresAt <= u.ObservedAt {
			return errors.New("admission offer must expire after observed_at")
		}
		requested, err := safeCapacityAdd(p.WorkUnits, p.ResolutionUnits)
		if err != nil {
			return err
		}
		withReserve, err := safeCapacityAdd(requested, s.Capacity.CorrectionEgressReserveUnits)
		if err != nil {
			return err
		}
		if withReserve > locus.CapacityCeilingUnits {
			return errors.New("admission offer cannot fit its full lifecycle inside the locus capacity ceiling")
		}
		snapshot, err := s.capacitySnapshot()
		if err != nil {
			return err
		}
		if requested > snapshot.AvailableUnits {
			return fmt.Errorf("admission offer requires %d lifecycle units but only %d are presently available", requested, snapshot.AvailableUnits)
		}
		if p.OfferExpiresAt-u.ObservedAt > genesis.Capacity.MaxAdmissionOfferSeconds {
			return errors.New("admission offer exceeds genesis duration limit")
		}
	}
	offerStatus := "NONE"
	if p.Disposition == "ADMIT" {
		offerStatus = "OPEN"
	}
	locus.Gates[p.ParticipantID] = GateState{
		ParticipantID:   p.ParticipantID,
		PresentationID:  p.PresentationID,
		Disposition:     p.Disposition,
		ReasonSHA256:    p.ReasonSHA256,
		WorkUnits:       p.WorkUnits,
		ResolutionUnits: p.ResolutionUnits,
		OfferExpiresAt:  p.OfferExpiresAt,
		OfferStatus:     offerStatus,
		TransitionID:    transitionID,
	}
	return nil
}

func (s *RuntimeState) applyEnter(u *UnsignedTransition, transitionID string) error {
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	p, err := decodePayload[EnterPayload](u.Payload)
	if err != nil {
		return err
	}
	presentation, exists := locus.Presentations[u.ActorID]
	if !exists || presentation.PresentationID != p.PresentationID {
		return errors.New("entry does not address the actor's current presentation")
	}
	gate, exists := locus.Gates[u.ActorID]
	if !exists || gate.PresentationID != p.PresentationID || gate.Disposition != "ADMIT" {
		return errors.New("entry requires an explicit current ADMIT disposition")
	}
	if participantPresent(locus, u.ActorID) {
		return errors.New("actor is already entered")
	}
	for _, prior := range s.EntryHistory {
		if prior.ParticipantID == u.ActorID {
			return errors.New("returning actor must use REENTER with a sealed departure checkpoint")
		}
	}
	if gate.OfferStatus != "OPEN" || u.ObservedAt >= gate.OfferExpiresAt {
		return errors.New("entry requires a current unexpired admission offer")
	}
	requested, err := safeCapacityAdd(gate.WorkUnits, gate.ResolutionUnits)
	if err != nil {
		return err
	}
	snapshot, err := s.capacitySnapshot()
	if err != nil {
		return err
	}
	if requested > snapshot.AvailableUnits {
		return fmt.Errorf("entry requires %d lifecycle units but only %d are presently available", requested, snapshot.AvailableUnits)
	}
	entry := EntryState{
		ParticipantID:           u.ActorID,
		PresentationID:          p.PresentationID,
		EntryTransitionID:       transitionID,
		IngressMode:             "FIRST_ENTRY",
		ReservedWorkUnits:       gate.WorkUnits,
		ReservedResolutionUnits: gate.ResolutionUnits,
		Status:                  "PRESENT",
	}
	locus.Entries[u.ActorID] = entry
	s.EntryHistory[transitionID] = entry
	gate.OfferStatus = "ENTERED"
	gate.EndTransitionID = transitionID
	locus.Gates[u.ActorID] = gate
	return nil
}

type continuityStepEnvelope struct {
	Schema                     string `json:"schema"`
	ParticipantID              string `json:"participant_id"`
	Ordinal                    uint32 `json:"ordinal"`
	PredecessorStateCommitment string `json:"predecessor_state_commitment"`
	DeltaSHA256                string `json:"delta_sha256"`
}

type continuityPassageEnvelope struct {
	Schema                string           `json:"schema"`
	ParticipantID         string           `json:"participant_id"`
	FromStateCommitment   string           `json:"from_state_commitment"`
	Passage               []ContinuityStep `json:"passage"`
	ResultStateCommitment string           `json:"result_state_commitment"`
}

func continuitySuccessorCommitment(participantID, predecessorStateCommitment, deltaSHA256 string, ordinal uint32) (string, error) {
	envelope := continuityStepEnvelope{
		Schema: continuityStepSchema, ParticipantID: participantID, Ordinal: ordinal,
		PredecessorStateCommitment: predecessorStateCommitment, DeltaSHA256: deltaSHA256,
	}
	bytes, err := json.Marshal(&envelope)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(bytes)
	return hex.EncodeToString(digest[:]), nil
}

func continuityPassageCommitment(participantID, fromStateCommitment string, passage []ContinuityStep, expectedResult string) (string, error) {
	current := fromStateCommitment
	for i, step := range passage {
		expected, err := continuitySuccessorCommitment(participantID, current, step.DeltaSHA256, uint32(i+1))
		if err != nil {
			return "", err
		}
		if step.SuccessorStateCommitment != expected {
			return "", fmt.Errorf("continuity passage step %d does not succeed its named predecessor", i)
		}
		current = step.SuccessorStateCommitment
	}
	if current != expectedResult {
		return "", errors.New("continuity passage result does not equal the required state commitment")
	}
	envelope := continuityPassageEnvelope{
		Schema: continuityPassageSchema, ParticipantID: participantID,
		FromStateCommitment: fromStateCommitment, Passage: passage, ResultStateCommitment: expectedResult,
	}
	bytes, err := json.Marshal(&envelope)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(bytes)
	return hex.EncodeToString(digest[:]), nil
}

func (s *RuntimeState) applyCheckpointDeparture(u *UnsignedTransition, transitionID string) error {
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	entry, exists := locus.Entries[u.ActorID]
	if !exists || entry.Status != "PRESENT" {
		return errors.New("departure checkpoint requires a presently entered participant")
	}
	p, err := decodePayload[CheckpointDeparturePayload](u.Payload)
	if err != nil {
		return err
	}
	if entry.EntryTransitionID != p.EntryTransitionID || entry.PresentationID != p.PresentationID {
		return errors.New("departure checkpoint does not bind the participant's current entry and presentation")
	}
	if entry.DepartureCheckpointID != "" {
		return errors.New("current entry already has a departure checkpoint")
	}
	presentation, exists := s.PresentationHistory[p.PresentationID]
	if !exists || presentation.ParticipantID != u.ActorID {
		return errors.New("departure checkpoint references an unknown participant presentation")
	}
	if p.FromStateCommitment != presentation.StateCommitment {
		return errors.New("departure checkpoint must begin at the state commitment carried on entry")
	}
	passageSHA, err := continuityPassageCommitment(u.ActorID, p.FromStateCommitment, p.Passage, p.DepartureStateCommitment)
	if err != nil {
		return err
	}
	checkpoint := DepartureCheckpointState{
		CheckpointID: transitionID, ParticipantID: u.ActorID, LocusID: u.LocusID,
		PresentationID: p.PresentationID, EntryTransitionID: p.EntryTransitionID,
		FromStateCommitment: p.FromStateCommitment, DepartureStateCommitment: p.DepartureStateCommitment,
		PassageSHA256: passageSHA, PassageStepCount: uint32(len(p.Passage)), Status: "OPEN",
	}
	s.DepartureCheckpoints[transitionID] = checkpoint
	entry.DepartureCheckpointID = transitionID
	entry.DepartureStateCommitment = p.DepartureStateCommitment
	locus.Entries[u.ActorID] = entry
	s.EntryHistory[entry.EntryTransitionID] = entry
	return nil
}

func (s *RuntimeState) applyReenter(u *UnsignedTransition, transitionID string) error {
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	if participantPresent(locus, u.ActorID) {
		return errors.New("actor is already entered")
	}
	p, err := decodePayload[ReenterPayload](u.Payload)
	if err != nil {
		return err
	}
	presentation, exists := locus.Presentations[u.ActorID]
	if !exists || presentation.PresentationID != p.PresentationID {
		return errors.New("re-entry does not address the actor's fresh current presentation")
	}
	gate, exists := locus.Gates[u.ActorID]
	if !exists || gate.PresentationID != p.PresentationID || gate.Disposition != "ADMIT" || gate.OfferStatus != "OPEN" {
		return errors.New("re-entry requires an explicit fresh current ADMIT disposition")
	}
	if u.ObservedAt >= gate.OfferExpiresAt {
		return errors.New("re-entry requires an unexpired admission offer")
	}
	priorEntry, exists := s.EntryHistory[p.PriorEntryTransitionID]
	if !exists || priorEntry.ParticipantID != u.ActorID || priorEntry.Status == "PRESENT" {
		return errors.New("re-entry must name the actor's ended prior entry")
	}
	if priorEntry.DepartureCheckpointID != p.DepartureCheckpointID {
		return errors.New("re-entry checkpoint does not belong to the named prior entry")
	}
	checkpoint, exists := s.DepartureCheckpoints[p.DepartureCheckpointID]
	if !exists || checkpoint.ParticipantID != u.ActorID || checkpoint.EntryTransitionID != p.PriorEntryTransitionID {
		return errors.New("re-entry references an unknown or foreign departure checkpoint")
	}
	if checkpoint.Status != "SEALED" {
		return errors.New("re-entry requires a sealed, unconsumed departure checkpoint")
	}
	priorLocus := s.Loci[checkpoint.LocusID]
	if priorLocus == nil {
		return errors.New("departure checkpoint names an unknown prior locus")
	}
	if priorLocus.Phase == "CLOSED" {
		residue, exists := s.ResidueHistory[p.ResidueID]
		if !exists || residue.ParticipantID != u.ActorID || residue.EntryTransitionID != p.PriorEntryTransitionID || residue.DepartureCheckpointID != p.DepartureCheckpointID {
			return errors.New("re-entry from a closed locus requires its exact addressed residue")
		}
	} else if p.ResidueID != "" {
		return errors.New("same-open-locus re-entry cannot claim a closure residue")
	}
	passageSHA, err := continuityPassageCommitment(u.ActorID, checkpoint.DepartureStateCommitment, p.Passage, presentation.StateCommitment)
	if err != nil {
		return err
	}
	requested, err := safeCapacityAdd(gate.WorkUnits, gate.ResolutionUnits)
	if err != nil {
		return err
	}
	snapshot, err := s.capacitySnapshot()
	if err != nil {
		return err
	}
	if requested > snapshot.AvailableUnits {
		return fmt.Errorf("re-entry requires %d lifecycle units but only %d are presently available", requested, snapshot.AvailableUnits)
	}
	entry := EntryState{
		ParticipantID: u.ActorID, PresentationID: p.PresentationID, EntryTransitionID: transitionID,
		IngressMode: "RENEWED_ENTRY", PriorEntryTransitionID: p.PriorEntryTransitionID,
		PriorDepartureCheckpointID: p.DepartureCheckpointID, PriorResidueID: p.ResidueID,
		IngressPassageSHA256: passageSHA, ReservedWorkUnits: gate.WorkUnits,
		ReservedResolutionUnits: gate.ResolutionUnits, Status: "PRESENT",
	}
	locus.Entries[u.ActorID] = entry
	s.EntryHistory[transitionID] = entry
	checkpoint.Status = "CONSUMED"
	checkpoint.ConsumedByEntryTransitionID = transitionID
	s.DepartureCheckpoints[p.DepartureCheckpointID] = checkpoint
	gate.OfferStatus = "ENTERED"
	gate.EndTransitionID = transitionID
	locus.Gates[u.ActorID] = gate
	return nil
}

func (s *RuntimeState) applyObserveCrossing(u *UnsignedTransition, transitionID string) error {
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	if !s.isActiveAuthority(u.ActorID) && !participantPresent(locus, u.ActorID) {
		return errors.New("crossing observation requires host medium or entered participant")
	}
	p, err := decodePayload[ObserveCrossingPayload](u.Payload)
	if err != nil {
		return err
	}
	if _, exists := locus.Matter[p.MatterID]; exists {
		return errors.New("matter ID has already crossed this locus")
	}
	locus.Matter[p.MatterID] = MatterState{
		MatterID:      p.MatterID,
		ObserverID:    u.ActorID,
		ContentSHA256: p.ContentSHA256,
		MediaType:     p.MediaType,
		Claim:         p.Claim,
		CrossingID:    transitionID,
		Dispositions:  make(map[string]MatterPosture),
	}
	return nil
}

func (s *RuntimeState) applyMatterDisposition(u *UnsignedTransition, transitionID string) error {
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	if !s.isActiveAuthority(u.ActorID) && !participantPresent(locus, u.ActorID) {
		return errors.New("matter disposition requires host medium or entered participant")
	}
	p, err := decodePayload[MatterDispositionPayload](u.Payload)
	if err != nil {
		return err
	}
	if s.derivedBody().Posture == "HOLD_CAPACITY_DEFICIT" && p.Disposition == "ADMIT" {
		return errors.New("capacity deficit allows Matter refusal or HOLD but not new admission")
	}
	matter, exists := locus.Matter[p.MatterID]
	if !exists {
		return errors.New("matter has not crossed this locus")
	}
	matter.Dispositions[u.ActorID] = MatterPosture{
		ActorID:      u.ActorID,
		Disposition:  p.Disposition,
		ReasonSHA256: p.ReasonSHA256,
		TransitionID: transitionID,
	}
	locus.Matter[p.MatterID] = matter
	return nil
}

func (s *RuntimeState) applyRecordEmergence(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capRecordEmergence); err != nil {
		return err
	}
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	p, err := decodePayload[RecordEmergencePayload](u.Payload)
	if err != nil {
		return err
	}
	if _, exists := locus.Emergence[p.EmergenceID]; exists {
		return errors.New("emergence ID has already been used")
	}
	for _, contributorID := range p.ContributorIDs {
		if !participantPresent(locus, contributorID) {
			return fmt.Errorf("contributor %s is not present", contributorID)
		}
		for _, matterID := range p.MatterIDs {
			matter, exists := locus.Matter[matterID]
			if !exists {
				return fmt.Errorf("matter %s has not crossed", matterID)
			}
			posture, exists := matter.Dispositions[contributorID]
			if !exists || posture.Disposition != "ADMIT" {
				return fmt.Errorf("contributor %s has not locally admitted matter %s", contributorID, matterID)
			}
		}
	}
	locus.Emergence[p.EmergenceID] = EmergenceState{
		EmergenceID:       p.EmergenceID,
		ContributorIDs:    sortedCopy(p.ContributorIDs),
		MatterIDs:         sortedCopy(p.MatterIDs),
		Kind:              p.Kind,
		DescriptionSHA256: p.DescriptionSHA256,
		TransitionID:      transitionID,
		Effect:            "FIELD_LOCAL_RECORD_ONLY",
	}
	return nil
}

func (s *RuntimeState) applyCorrection(u *UnsignedTransition, transitionID string) error {
	p, err := decodePayload[CorrectPayload](u.Payload)
	if err != nil {
		return err
	}
	target, exists := s.Events[p.TargetTransitionID]
	if !exists {
		return errors.New("correction target does not exist")
	}
	if target.ActorID != u.ActorID {
		return errors.New("actor may only correct its own attributed transition")
	}
	s.Corrections[p.TargetTransitionID] = append(s.Corrections[p.TargetTransitionID], CorrectionRecord{
		CorrectionTransitionID: transitionID,
		TargetTransitionID:     p.TargetTransitionID,
		ReplacementCommitment:  p.ReplacementCommitment,
		ReasonSHA256:           p.ReasonSHA256,
		ActorID:                u.ActorID,
	})
	return nil
}

func (s *RuntimeState) applyExit(u *UnsignedTransition, transitionID string) error {
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	entry, exists := locus.Entries[u.ActorID]
	if !exists || entry.Status != "PRESENT" {
		return errors.New("actor is not presently entered")
	}
	p, err := decodePayload[ExitPayload](u.Payload)
	if err != nil {
		return err
	}
	if entry.DepartureCheckpointID != p.DepartureCheckpointID {
		return errors.New("exit must name the current entry's departure checkpoint")
	}
	checkpoint, exists := s.DepartureCheckpoints[p.DepartureCheckpointID]
	if !exists || checkpoint.ParticipantID != u.ActorID || checkpoint.EntryTransitionID != entry.EntryTransitionID || checkpoint.Status != "OPEN" {
		return errors.New("exit requires the current entry's open departure checkpoint")
	}
	checkpoint.Status = "SEALED"
	checkpoint.SealTransitionID = transitionID
	s.DepartureCheckpoints[p.DepartureCheckpointID] = checkpoint
	entry.Status = "EXITED"
	entry.EndTransitionID = transitionID
	entry.EndReasonSHA256 = p.ReasonSHA256
	locus.Entries[u.ActorID] = entry
	s.EntryHistory[entry.EntryTransitionID] = entry
	return nil
}

func (s *RuntimeState) applyClose(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capCloseLocus); err != nil {
		return err
	}
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	p, err := decodePayload[ClosePayload](u.Payload)
	if err != nil {
		return err
	}
	for participantID, entry := range locus.Entries {
		if entry.Status != "PRESENT" && entry.Status != "EXITED" {
			continue
		}
		checkpoint, exists := s.DepartureCheckpoints[entry.DepartureCheckpointID]
		if !exists || checkpoint.ParticipantID != participantID || checkpoint.EntryTransitionID != entry.EntryTransitionID {
			return fmt.Errorf("participant %s lacks a departure checkpoint for its current entry", participantID)
		}
		if entry.Status == "PRESENT" {
			if checkpoint.Status != "OPEN" {
				return fmt.Errorf("participant %s departure checkpoint is not open for closure", participantID)
			}
			checkpoint.Status = "SEALED"
			checkpoint.SealTransitionID = transitionID
			s.DepartureCheckpoints[entry.DepartureCheckpointID] = checkpoint
		} else if checkpoint.Status != "SEALED" {
			return fmt.Errorf("participant %s exited without a sealed departure checkpoint", participantID)
		}
		residue, err := makeResidue(s, locus, entry, checkpoint, transitionID)
		if err != nil {
			return err
		}
		locus.Residues[participantID] = residue
		s.ResidueHistory[residue.ResidueID] = residue
		entry.ResidueID = residue.ResidueID
		if entry.Status == "PRESENT" {
			entry.Status = "CLOSED_WITH_FIELD"
			entry.EndTransitionID = transitionID
			entry.EndReasonSHA256 = p.ClosureBasisSHA256
		}
		locus.Entries[participantID] = entry
		s.EntryHistory[entry.EntryTransitionID] = entry
	}
	locus.Phase = "CLOSED"
	locus.CloseTransitionID = transitionID
	locus.ClosureBasisSHA256 = p.ClosureBasisSHA256
	s.ActiveLocusID = ""
	return nil
}

func makeResidue(state *RuntimeState, locus *LocusState, entry EntryState, checkpoint DepartureCheckpointState, closureTransitionID string) (ResidueState, error) {
	presentation, exists := state.PresentationHistory[entry.PresentationID]
	if !exists {
		return ResidueState{}, errors.New("cannot address residue without the entry's preserved presentation")
	}
	envelope := residueEnvelope{
		Schema:                   residueSchema,
		AddressedToLocalityID:    presentation.LocalityReference,
		SourceLocalityID:         state.HostLocalityID,
		SourceLocusID:            locus.LocusID,
		SourceStateCommitment:    state.StateCommitment,
		ClosureTransitionID:      closureTransitionID,
		ParticipantID:            entry.ParticipantID,
		PresentationID:           entry.PresentationID,
		EntryTransitionID:        entry.EntryTransitionID,
		DepartureCheckpointID:    checkpoint.CheckpointID,
		DepartureStateCommitment: checkpoint.DepartureStateCommitment,
		Effect:                   residueEffect,
	}
	residueSHA, residueID, err := residueCommitments(envelope)
	if err != nil {
		return ResidueState{}, err
	}
	return ResidueState{
		ResidueID:                residueID,
		ResidueSHA256:            residueSHA,
		Schema:                   envelope.Schema,
		AddressedToLocalityID:    envelope.AddressedToLocalityID,
		SourceLocalityID:         state.HostLocalityID,
		SourceLocusID:            locus.LocusID,
		SourceStateCommitment:    state.StateCommitment,
		ClosureTransitionID:      closureTransitionID,
		ParticipantID:            entry.ParticipantID,
		PresentationID:           entry.PresentationID,
		EntryTransitionID:        entry.EntryTransitionID,
		DepartureCheckpointID:    checkpoint.CheckpointID,
		DepartureStateCommitment: checkpoint.DepartureStateCommitment,
		Effect:                   residueEffect,
	}, nil
}

func residueCommitments(envelope residueEnvelope) (string, string, error) {
	bytes, err := json.Marshal(&envelope)
	if err != nil {
		return "", "", err
	}
	residueDigest := sha256.Sum256(bytes)
	idInput := append([]byte(residueIDDomain+"\x00"), bytes...)
	idDigest := sha256.Sum256(idInput)
	return hex.EncodeToString(residueDigest[:]), hex.EncodeToString(idDigest[:]), nil
}

func (s *RuntimeState) applyIncorporateResidue(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capIncorporate); err != nil {
		return err
	}
	if u.LocusID != "" {
		return errors.New("residue incorporation is local and must not impersonate a live source locus")
	}
	p, err := decodePayload[IncorporateResiduePayload](u.Payload)
	if err != nil {
		return err
	}
	if p.AddressedToLocalityID != s.HostLocalityID {
		return errors.New("residue is not addressed to this locality")
	}
	envelope := residueEnvelope{
		Schema:                   p.Schema,
		AddressedToLocalityID:    p.AddressedToLocalityID,
		SourceLocalityID:         p.SourceLocalityID,
		SourceLocusID:            p.SourceLocusID,
		SourceStateCommitment:    p.SourceStateCommitment,
		ClosureTransitionID:      p.ClosureTransitionID,
		ParticipantID:            p.ParticipantID,
		PresentationID:           p.PresentationID,
		EntryTransitionID:        p.EntryTransitionID,
		DepartureCheckpointID:    p.DepartureCheckpointID,
		DepartureStateCommitment: p.DepartureStateCommitment,
		Effect:                   p.Effect,
	}
	expectedSHA, expectedID, err := residueCommitments(envelope)
	if err != nil {
		return err
	}
	if p.ResidueSHA256 != expectedSHA || p.ResidueID != expectedID {
		return errors.New("residue commitment does not match its addressed envelope")
	}
	if _, exists := s.IncorporatedResidues[p.ResidueID]; exists {
		return errors.New("residue has already been incorporated")
	}
	s.IncorporatedResidues[p.ResidueID] = IncorporatedResidue{
		ResidueID:                p.ResidueID,
		ResidueSHA256:            p.ResidueSHA256,
		Schema:                   p.Schema,
		AddressedToLocalityID:    p.AddressedToLocalityID,
		SourceLocalityID:         p.SourceLocalityID,
		SourceLocusID:            p.SourceLocusID,
		SourceStateCommitment:    p.SourceStateCommitment,
		ClosureTransitionID:      p.ClosureTransitionID,
		ParticipantID:            p.ParticipantID,
		PresentationID:           p.PresentationID,
		EntryTransitionID:        p.EntryTransitionID,
		DepartureCheckpointID:    p.DepartureCheckpointID,
		DepartureStateCommitment: p.DepartureStateCommitment,
		TransitionID:             transitionID,
		Effect:                   "LOCAL_STATE_INCORPORATION_ONLY",
	}
	return nil
}

func (s *RuntimeState) applyDeclareCapacity(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capDeclareCapacity); err != nil {
		return err
	}
	if u.LocusID != "" {
		return errors.New("capacity declaration is body-local and cannot impersonate a locus")
	}
	p, err := decodePayload[DeclareCapacityPayload](u.Payload)
	if err != nil {
		return err
	}
	if p.ActualUnits > s.Capacity.StructuralCeilingUnits {
		return errors.New("declared actual capacity exceeds the immutable structural ceiling")
	}
	s.Capacity.DeclaredActualUnits = p.ActualUnits
	s.Capacity.ResourceCommitmentSHA256 = p.ResourceCommitmentSHA256
	s.Capacity.LastDeclarationTransitionID = transitionID
	s.Capacity.LastDeclarationObservedAt = u.ObservedAt
	return nil
}

type currentnessEnvelope struct {
	Schema                  string           `json:"schema"`
	PreviousStateCommitment string           `json:"previous_state_commitment"`
	ObservedAt              int64            `json:"observed_at"`
	PresenceCount           uint64           `json:"presence_count"`
	Posture                 string           `json:"posture"`
	Capacity                CapacitySnapshot `json:"capacity"`
	CarrierSetSHA256        string           `json:"carrier_set_sha256"`
}

func (s *RuntimeState) currentnessCommitment(observedAt int64, carrierSetSHA256 string) (string, error) {
	snapshot, err := s.capacitySnapshot()
	if err != nil {
		return "", err
	}
	body := s.derivedBody()
	envelope := currentnessEnvelope{
		Schema: "LOCALITY_CURRENTNESS_003", PreviousStateCommitment: s.StateCommitment,
		ObservedAt: observedAt, PresenceCount: body.PresenceCount, Posture: body.Posture,
		Capacity: snapshot, CarrierSetSHA256: carrierSetSHA256,
	}
	bytes, err := json.Marshal(&envelope)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(bytes)
	return hex.EncodeToString(digest[:]), nil
}

func (s *RuntimeState) applyPulse(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capPulse); err != nil {
		return err
	}
	if u.LocusID != "" {
		return errors.New("pulse is body-local and cannot impersonate a locus")
	}
	p, err := decodePayload[PulsePayload](u.Payload)
	if err != nil {
		return err
	}
	expected, err := s.currentnessCommitment(u.ObservedAt, p.CarrierSetSHA256)
	if err != nil {
		return err
	}
	if p.CurrentnessCommitmentSHA256 != expected {
		return errors.New("pulse currentness commitment does not bind the exact pre-transition state")
	}
	body := s.derivedBody()
	s.Pulses[transitionID] = PulseState{
		TransitionID: transitionID, ActorID: u.ActorID, ObservedAt: u.ObservedAt,
		PresenceCount: body.PresenceCount, Posture: body.Posture,
		CurrentnessCommitmentSHA256: p.CurrentnessCommitmentSHA256,
		CarrierSetSHA256:            p.CarrierSetSHA256,
		Effect:                      "SIGNED_LOCAL_CURRENTNESS_NOT_EXTERNAL_LIVENESS_PROOF",
	}
	s.LastPulseID = transitionID
	return nil
}

func (s *RuntimeState) applyReclaimAdmissionOffer(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capReclaimOffer); err != nil {
		return err
	}
	locus, err := s.activeLocus(u.LocusID)
	if err != nil {
		return err
	}
	p, err := decodePayload[ReclaimAdmissionOfferPayload](u.Payload)
	if err != nil {
		return err
	}
	gate, exists := locus.Gates[p.ParticipantID]
	if !exists || gate.PresentationID != p.PresentationID || gate.TransitionID != p.OfferTransitionID || gate.Disposition != "ADMIT" {
		return errors.New("reclamation does not address the participant's current admission offer")
	}
	if gate.OfferStatus != "OPEN" || u.ObservedAt < gate.OfferExpiresAt {
		return errors.New("only an expired, unentered admission offer can be reclaimed")
	}
	if participantPresent(locus, p.ParticipantID) {
		return errors.New("an entered lifecycle reservation cannot be reclaimed; it requires exit or closure")
	}
	gate.OfferStatus = "EXPIRED"
	gate.EndTransitionID = transitionID
	locus.Gates[p.ParticipantID] = gate
	return nil
}

func (s *RuntimeState) applyRegisterContinuityAuthority(u *UnsignedTransition, transitionID string) error {
	if err := s.requireFormationAuthority(u.ActorID); err != nil {
		return err
	}
	if u.LocusID != "" {
		return errors.New("authority registration is body-local")
	}
	p, err := decodePayload[RegisterContinuityAuthorityPayload](u.Payload)
	if err != nil {
		return err
	}
	publicKey, err := decodeLowerHex("payload.public_key", p.PublicKey, ed25519.PublicKeySize)
	if err != nil {
		return err
	}
	actorID := operationalActorID(ed25519.PublicKey(publicKey))
	if _, exists := s.ActorKeys[actorID]; exists {
		return errors.New("continuity authority key is already bound in this locality")
	}
	for _, authority := range s.Authorities {
		if authority.KeyID == p.KeyID {
			return errors.New("continuity authority key_id has already been used")
		}
	}
	s.Authorities[actorID] = AuthorityState{
		ActorID: actorID, KeyID: p.KeyID, PublicKey: p.PublicKey, Role: "CONTINUITY",
		Capabilities: sortedCopy(p.Capabilities), MandateSHA256: p.MandateSHA256,
		Status: "ACTIVE", RegisteredTransitionID: transitionID,
	}
	s.ActorKeys[actorID] = p.PublicKey
	s.NextNonces[actorID] = 0
	return nil
}

func authorityHasContinuityCore(authority AuthorityState) bool {
	if authority.Role != "CONTINUITY" || authority.Status != "ACTIVE" {
		return false
	}
	for _, capability := range continuityAuthorityCapabilities {
		if !contains(authority.Capabilities, capability) {
			return false
		}
	}
	return true
}

func (s *RuntimeState) applyExhaustFormationAuthority(genesis *Genesis, u *UnsignedTransition, transitionID string) error {
	if err := s.requireFormationAuthority(u.ActorID); err != nil {
		return err
	}
	if u.LocusID != "" || s.ActiveLocusID != "" || s.derivedBody().Posture != "DORMANT_P0" {
		return errors.New("formation authority can exhaust only at a dormant body checkpoint with no open locus")
	}
	if _, err := decodePayload[ExhaustFormationAuthorityPayload](u.Payload); err != nil {
		return err
	}
	var eligible uint32
	for _, authority := range s.Authorities {
		if authorityHasContinuityCore(authority) {
			eligible++
		}
	}
	if eligible < genesis.Continuity.MinimumContinuityAuthorities {
		return fmt.Errorf("formation exhaustion requires %d fully capable continuity authorities; observed %d", genesis.Continuity.MinimumContinuityAuthorities, eligible)
	}
	authority := s.Authorities[s.FormationAuthorityID]
	authority.Status = "EXHAUSTED"
	authority.ExhaustedTransitionID = transitionID
	s.Authorities[s.FormationAuthorityID] = authority
	return nil
}

func successorProposalCommitment(p *ProposeSuccessorPayload) (string, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	domainBytes := append([]byte("LOCALITY_SUCCESSOR_PROPOSAL_003\x00"), bytes...)
	digest := sha256.Sum256(domainBytes)
	return hex.EncodeToString(digest[:]), nil
}

func (s *RuntimeState) applyProposeSuccessor(u *UnsignedTransition, transitionID string) error {
	if err := s.requireCapability(u.ActorID, capProposeSuccessor); err != nil {
		return err
	}
	if u.LocusID != "" {
		return errors.New("successor proposal is body-local")
	}
	p, err := decodePayload[ProposeSuccessorPayload](u.Payload)
	if err != nil {
		return err
	}
	if p.Protocol == vmName || p.VMID == expectedVMID {
		return errors.New("successor must have a distinct protocol and VM identity")
	}
	if _, exists := s.SuccessorProposals[p.ProposalID]; exists {
		return errors.New("successor proposal ID has already been used")
	}
	commitment, err := successorProposalCommitment(p)
	if err != nil {
		return err
	}
	s.SuccessorProposals[p.ProposalID] = &SuccessorProposal{
		ProposalID: p.ProposalID, ProposalCommitment: commitment, Protocol: p.Protocol,
		Version: p.Version, VMID: p.VMID, RuntimeSHA256: p.RuntimeSHA256,
		MigrationSHA256: p.MigrationSHA256, InvariantSetSHA256: p.InvariantSetSHA256,
		SourceReference: p.SourceReference, SourceSHA256: p.SourceSHA256,
		ActivationMode: p.ActivationMode, ProposedBy: u.ActorID,
		ProposalTransitionID: transitionID, Attestations: make(map[string]SuccessorAttestation),
	}
	return nil
}

func (s *RuntimeState) applyAttestSuccessor(u *UnsignedTransition, transitionID string) error {
	if err := s.requireContinuityCapability(u.ActorID, capAttestSuccessor); err != nil {
		return err
	}
	if u.LocusID != "" {
		return errors.New("successor attestation is body-local")
	}
	p, err := decodePayload[AttestSuccessorPayload](u.Payload)
	if err != nil {
		return err
	}
	proposal, exists := s.SuccessorProposals[p.ProposalID]
	if !exists || proposal.ProposalCommitment != p.ProposalCommitment {
		return errors.New("attestation does not bind an exact current successor proposal")
	}
	if _, exists := proposal.Attestations[u.ActorID]; exists {
		return errors.New("continuity authority has already attested this successor proposal")
	}
	proposal.Attestations[u.ActorID] = SuccessorAttestation{u.ActorID, p.BasisSHA256, transitionID}
	return nil
}

func (s *RuntimeState) applyActivateSuccessor(genesis *Genesis, u *UnsignedTransition, transitionID string) error {
	if err := s.requireContinuityCapability(u.ActorID, capActivateSuccessor); err != nil {
		return err
	}
	if u.LocusID != "" || s.ActiveLocusID != "" || s.derivedBody().Posture != "DORMANT_P0" {
		return errors.New("successor activation requires a dormant checkpoint with no open locus or current participant")
	}
	formation := s.Authorities[s.FormationAuthorityID]
	if formation.Status != "EXHAUSTED" {
		return errors.New("successor activation requires explicit exhaustion of formation authority")
	}
	p, err := decodePayload[ActivateSuccessorPayload](u.Payload)
	if err != nil {
		return err
	}
	proposal, exists := s.SuccessorProposals[p.ProposalID]
	if !exists || proposal.ProposalCommitment != p.ProposalCommitment {
		return errors.New("activation does not bind an exact current successor proposal")
	}
	var attestations uint32
	for actorID := range proposal.Attestations {
		if authority, exists := s.Authorities[actorID]; exists && authority.Role == "CONTINUITY" && authority.Status == "ACTIVE" {
			attestations++
		}
	}
	if attestations < genesis.Continuity.SuccessorAttestationThreshold {
		return fmt.Errorf("successor activation requires %d active continuity attestations; observed %d", genesis.Continuity.SuccessorAttestationThreshold, attestations)
	}
	s.SuccessorBoundary = &SuccessorBoundary{
		ProposalID: proposal.ProposalID, ProposalCommitment: proposal.ProposalCommitment,
		Protocol: proposal.Protocol, Version: proposal.Version, VMID: proposal.VMID,
		RuntimeSHA256: proposal.RuntimeSHA256, MigrationSHA256: proposal.MigrationSHA256,
		InvariantSetSHA256: proposal.InvariantSetSHA256, SourceReference: proposal.SourceReference,
		SourceSHA256: proposal.SourceSHA256, ActivationTransitionID: transitionID,
		FrozenAtRevision: u.Revision,
		Effect:           "PREDECESSOR_FROZEN_AWAITING_EXPLICIT_EXTERNAL_BINARY_HANDOFF",
	}
	return nil
}

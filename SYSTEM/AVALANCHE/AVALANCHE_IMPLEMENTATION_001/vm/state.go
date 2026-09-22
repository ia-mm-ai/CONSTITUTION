package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
)

const stateSchema = "PRESENCE_AVALANCHE_RUNTIME_STATE_001"

const (
	capOpenLocus         = "OPEN_LOCUS"
	capRegulateGate      = "REGULATE_GATE"
	capRecordEmergence   = "RECORD_EMERGENCE"
	capCloseLocus        = "CLOSE_LOCUS"
	capIncorporate       = "INCORPORATE_RESIDUE"
	capDeclareCapacity   = "DECLARE_CAPACITY"
	capPulse             = "PULSE"
	capReclaimOffer      = "RECLAIM_ADMISSION_OFFER"
	capProposeSuccessor  = "PROPOSE_SUCCESSOR"
	capAttestSuccessor   = "ATTEST_SUCCESSOR"
	capActivateSuccessor = "ACTIVATE_SUCCESSOR"
	capRegisterAuthority = "REGISTER_CONTINUITY_AUTHORITY"
	capExhaustFormation  = "EXHAUST_FORMATION_AUTHORITY"
)

var continuityAuthorityCapabilities = []string{
	capOpenLocus, capRegulateGate, capRecordEmergence, capCloseLocus, capIncorporate,
	capDeclareCapacity, capPulse, capReclaimOffer, capProposeSuccessor,
	capAttestSuccessor, capActivateSuccessor,
}

var formationAuthorityCapabilities = append(
	append([]string(nil), continuityAuthorityCapabilities...),
	capRegisterAuthority,
	capExhaustFormation,
)

type RuntimeState struct {
	Schema               string                              `json:"schema"`
	Revision             uint64                              `json:"revision"`
	StateCommitment      string                              `json:"state_commitment"`
	LastTransitionID     string                              `json:"last_transition_id"`
	LastObservedAt       int64                               `json:"last_observed_at"`
	HostLocalityID       string                              `json:"host_locality_id"`
	ProfileFormID        string                              `json:"profile_form_id"`
	ProfileFormSHA256    string                              `json:"profile_form_sha256"`
	ActiveLocusID        string                              `json:"active_locus_id"`
	Body                 BodyState                           `json:"body"`
	Capacity             CapacityState                       `json:"capacity"`
	Authorities          map[string]AuthorityState           `json:"authorities"`
	FormationAuthorityID string                              `json:"formation_authority_id"`
	Pulses               map[string]PulseState               `json:"pulses"`
	LastPulseID          string                              `json:"last_pulse_id"`
	SuccessorProposals   map[string]*SuccessorProposal       `json:"successor_proposals"`
	SuccessorBoundary    *SuccessorBoundary                  `json:"successor_boundary"`
	ActorKeys            map[string]string                   `json:"actor_keys"`
	NextNonces           map[string]uint64                   `json:"next_nonces"`
	PresentationHistory  map[string]PresentationState        `json:"presentation_history"`
	EntryHistory         map[string]EntryState               `json:"entry_history"`
	DepartureCheckpoints map[string]DepartureCheckpointState `json:"departure_checkpoints"`
	ResidueHistory       map[string]ResidueState             `json:"residue_history"`
	Loci                 map[string]*LocusState              `json:"loci"`
	Events               map[string]EventRecord              `json:"events"`
	Corrections          map[string][]CorrectionRecord       `json:"corrections"`
	IncorporatedResidues map[string]IncorporatedResidue      `json:"incorporated_residues"`
}

type BodyState struct {
	Posture       string `json:"posture"`
	PresenceCount uint64 `json:"presence_count"`
	Reason        string `json:"reason"`
}

type CapacityState struct {
	Unit                         string `json:"unit"`
	StructuralCeilingUnits       uint64 `json:"structural_ceiling_units"`
	DeclaredActualUnits          uint64 `json:"declared_actual_units"`
	CorrectionEgressReserveUnits uint64 `json:"correction_egress_reserve_units"`
	ResourceCommitmentSHA256     string `json:"resource_commitment_sha256"`
	LastDeclarationTransitionID  string `json:"last_declaration_transition_id"`
	LastDeclarationObservedAt    int64  `json:"last_declaration_observed_at"`
}

type CapacitySnapshot struct {
	StructuralCeilingUnits uint64 `json:"structural_ceiling_units"`
	DeclaredActualUnits    uint64 `json:"declared_actual_units"`
	EffectiveActualUnits   uint64 `json:"effective_actual_units"`
	UnresolvedUnits        uint64 `json:"unresolved_units"`
	CorrectionEgressUnits  uint64 `json:"correction_egress_units"`
	AvailableUnits         uint64 `json:"available_units"`
	DeficitUnits           uint64 `json:"deficit_units"`
	Rule                   string `json:"rule"`
}

type AuthorityState struct {
	ActorID                string   `json:"actor_id"`
	KeyID                  string   `json:"key_id"`
	PublicKey              string   `json:"public_key"`
	Role                   string   `json:"role"`
	Capabilities           []string `json:"capabilities"`
	MandateSHA256          string   `json:"mandate_sha256"`
	Status                 string   `json:"status"`
	RegisteredTransitionID string   `json:"registered_transition_id"`
	ExhaustedTransitionID  string   `json:"exhausted_transition_id"`
}

type PulseState struct {
	TransitionID                string `json:"transition_id"`
	ActorID                     string `json:"actor_id"`
	ObservedAt                  int64  `json:"observed_at"`
	PresenceCount               uint64 `json:"presence_count"`
	Posture                     string `json:"posture"`
	CurrentnessCommitmentSHA256 string `json:"currentness_commitment_sha256"`
	CarrierSetSHA256            string `json:"carrier_set_sha256"`
	Effect                      string `json:"effect"`
}

type SuccessorProposal struct {
	ProposalID           string                          `json:"proposal_id"`
	ProposalCommitment   string                          `json:"proposal_commitment"`
	Protocol             string                          `json:"protocol"`
	Version              string                          `json:"version"`
	VMID                 string                          `json:"vm_id"`
	RuntimeSHA256        string                          `json:"runtime_sha256"`
	MigrationSHA256      string                          `json:"migration_sha256"`
	InvariantSetSHA256   string                          `json:"invariant_set_sha256"`
	SourceReference      string                          `json:"source_reference"`
	SourceSHA256         string                          `json:"source_sha256"`
	ActivationMode       string                          `json:"activation_mode"`
	ProposedBy           string                          `json:"proposed_by"`
	ProposalTransitionID string                          `json:"proposal_transition_id"`
	Attestations         map[string]SuccessorAttestation `json:"attestations"`
}

type SuccessorAttestation struct {
	ActorID      string `json:"actor_id"`
	BasisSHA256  string `json:"basis_sha256"`
	TransitionID string `json:"transition_id"`
}

type SuccessorBoundary struct {
	ProposalID             string `json:"proposal_id"`
	ProposalCommitment     string `json:"proposal_commitment"`
	Protocol               string `json:"protocol"`
	Version                string `json:"version"`
	VMID                   string `json:"vm_id"`
	RuntimeSHA256          string `json:"runtime_sha256"`
	MigrationSHA256        string `json:"migration_sha256"`
	InvariantSetSHA256     string `json:"invariant_set_sha256"`
	SourceReference        string `json:"source_reference"`
	SourceSHA256           string `json:"source_sha256"`
	ActivationTransitionID string `json:"activation_transition_id"`
	FrozenAtRevision       uint64 `json:"frozen_at_revision"`
	Effect                 string `json:"effect"`
}

type LocusState struct {
	LocusID                string                       `json:"locus_id"`
	Phase                  string                       `json:"phase"`
	BoundTransitionID      string                       `json:"bound_transition_id"`
	PurposeSHA256          string                       `json:"purpose_sha256"`
	ClosureConditionSHA256 string                       `json:"closure_condition_sha256"`
	CapacityCeilingUnits   uint64                       `json:"capacity_ceiling_units"`
	Presentations          map[string]PresentationState `json:"presentations"`
	Gates                  map[string]GateState         `json:"gates"`
	Entries                map[string]EntryState        `json:"entries"`
	Matter                 map[string]MatterState       `json:"matter"`
	Emergence              map[string]EmergenceState    `json:"emergence"`
	Residues               map[string]ResidueState      `json:"residues"`
	CloseTransitionID      string                       `json:"close_transition_id"`
	ClosureBasisSHA256     string                       `json:"closure_basis_sha256"`
}

type PresentationState struct {
	PresentationID     string   `json:"presentation_id"`
	ParticipantID      string   `json:"participant_id"`
	LocalityReference  string   `json:"locality_reference"`
	SourceReference    string   `json:"source_reference"`
	SourceSHA256       string   `json:"source_sha256"`
	NucleusVersion     string   `json:"nucleus_version"`
	NucleusSHA256      string   `json:"nucleus_sha256"`
	FormID             string   `json:"form_id"`
	FormSHA256         string   `json:"form_sha256"`
	StateCommitment    string   `json:"state_commitment"`
	MediumCapabilities []string `json:"medium_capabilities"`
	EffectCeiling      []string `json:"effect_ceiling"`
}

type GateState struct {
	ParticipantID   string `json:"participant_id"`
	PresentationID  string `json:"presentation_id"`
	Disposition     string `json:"disposition"`
	ReasonSHA256    string `json:"reason_sha256"`
	WorkUnits       uint64 `json:"work_units"`
	ResolutionUnits uint64 `json:"resolution_units"`
	OfferExpiresAt  int64  `json:"offer_expires_at"`
	OfferStatus     string `json:"offer_status"`
	TransitionID    string `json:"transition_id"`
	EndTransitionID string `json:"end_transition_id"`
}

type EntryState struct {
	ParticipantID              string `json:"participant_id"`
	PresentationID             string `json:"presentation_id"`
	EntryTransitionID          string `json:"entry_transition_id"`
	IngressMode                string `json:"ingress_mode"`
	PriorEntryTransitionID     string `json:"prior_entry_transition_id"`
	PriorDepartureCheckpointID string `json:"prior_departure_checkpoint_id"`
	PriorResidueID             string `json:"prior_residue_id"`
	IngressPassageSHA256       string `json:"ingress_passage_sha256"`
	ReservedWorkUnits          uint64 `json:"reserved_work_units"`
	ReservedResolutionUnits    uint64 `json:"reserved_resolution_units"`
	DepartureCheckpointID      string `json:"departure_checkpoint_id"`
	DepartureStateCommitment   string `json:"departure_state_commitment"`
	ResidueID                  string `json:"residue_id"`
	Status                     string `json:"status"`
	EndTransitionID            string `json:"end_transition_id"`
	EndReasonSHA256            string `json:"end_reason_sha256"`
}

type DepartureCheckpointState struct {
	CheckpointID                string `json:"checkpoint_id"`
	ParticipantID               string `json:"participant_id"`
	LocusID                     string `json:"locus_id"`
	PresentationID              string `json:"presentation_id"`
	EntryTransitionID           string `json:"entry_transition_id"`
	FromStateCommitment         string `json:"from_state_commitment"`
	DepartureStateCommitment    string `json:"departure_state_commitment"`
	PassageSHA256               string `json:"passage_sha256"`
	PassageStepCount            uint32 `json:"passage_step_count"`
	Status                      string `json:"status"`
	SealTransitionID            string `json:"seal_transition_id"`
	ConsumedByEntryTransitionID string `json:"consumed_by_entry_transition_id"`
}

type MatterState struct {
	MatterID      string                   `json:"matter_id"`
	ObserverID    string                   `json:"observer_id"`
	ContentSHA256 string                   `json:"content_sha256"`
	MediaType     string                   `json:"media_type"`
	Claim         string                   `json:"claim"`
	CrossingID    string                   `json:"crossing_id"`
	Dispositions  map[string]MatterPosture `json:"dispositions"`
}

type MatterPosture struct {
	ActorID      string `json:"actor_id"`
	Disposition  string `json:"disposition"`
	ReasonSHA256 string `json:"reason_sha256"`
	TransitionID string `json:"transition_id"`
}
type EmergenceState struct {
	EmergenceID       string   `json:"emergence_id"`
	ContributorIDs    []string `json:"contributor_ids"`
	MatterIDs         []string `json:"matter_ids"`
	Kind              string   `json:"kind"`
	DescriptionSHA256 string   `json:"description_sha256"`
	TransitionID      string   `json:"transition_id"`
	Effect            string   `json:"effect"`
}
type ResidueState struct {
	ResidueID                string `json:"residue_id"`
	ResidueSHA256            string `json:"residue_sha256"`
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
type IncorporatedResidue struct {
	ResidueID                string `json:"residue_id"`
	ResidueSHA256            string `json:"residue_sha256"`
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
	TransitionID             string `json:"transition_id"`
	Effect                   string `json:"effect"`
}
type EventRecord struct {
	TransitionID string `json:"transition_id"`
	Revision     uint64 `json:"revision"`
	Operation    string `json:"operation"`
	ActorID      string `json:"actor_id"`
	LocusID      string `json:"locus_id"`
	ObservedAt   int64  `json:"observed_at"`
	Effect       string `json:"effect"`
}
type CorrectionRecord struct {
	CorrectionTransitionID string `json:"correction_transition_id"`
	TargetTransitionID     string `json:"target_transition_id"`
	ReplacementCommitment  string `json:"replacement_commitment"`
	ReasonSHA256           string `json:"reason_sha256"`
	ActorID                string `json:"actor_id"`
}

type Receipt struct {
	Schema          string           `json:"schema"`
	TransitionID    string           `json:"transition_id"`
	Revision        uint64           `json:"revision"`
	StateCommitment string           `json:"state_commitment"`
	Operation       string           `json:"operation"`
	ActorID         string           `json:"actor_id"`
	LocusID         string           `json:"locus_id"`
	Effect          string           `json:"effect"`
	Posture         string           `json:"posture"`
	PresenceCount   uint64           `json:"presence_count"`
	Capacity        CapacitySnapshot `json:"capacity"`
	NonEffects      []string         `json:"non_effects"`
}

func initialRuntimeState(genesis *Genesis) (*RuntimeState, error) {
	state := &RuntimeState{
		Schema: stateSchema, Revision: 0, HostLocalityID: genesis.Locality.ID,
		ProfileFormID: genesis.Profile.FormID, ProfileFormSHA256: genesis.Profile.FormSHA256,
		Body: BodyState{Posture: "DORMANT_P0", PresenceCount: 0, Reason: "NO_CURRENT_PARTICIPANT_BINDING"},
		Capacity: CapacityState{
			Unit: genesis.Capacity.Unit, StructuralCeilingUnits: genesis.Capacity.StructuralCeilingUnits,
			DeclaredActualUnits:          genesis.Capacity.InitialActualUnits,
			CorrectionEgressReserveUnits: genesis.Capacity.CorrectionEgressReserveUnits,
			ResourceCommitmentSHA256:     genesis.Locality.SourceSHA256,
		},
		FormationAuthorityID: genesis.Locality.ID,
		Authorities:          make(map[string]AuthorityState), Pulses: make(map[string]PulseState),
		SuccessorProposals: make(map[string]*SuccessorProposal),
		ActorKeys:          map[string]string{genesis.Locality.ID: genesis.Locality.Authority.PublicKey},
		NextNonces:         map[string]uint64{genesis.Locality.ID: 0}, Loci: make(map[string]*LocusState),
		PresentationHistory:  make(map[string]PresentationState),
		EntryHistory:         make(map[string]EntryState),
		DepartureCheckpoints: make(map[string]DepartureCheckpointState),
		ResidueHistory:       make(map[string]ResidueState),
		Events:               make(map[string]EventRecord), Corrections: make(map[string][]CorrectionRecord),
		IncorporatedResidues: make(map[string]IncorporatedResidue),
	}
	state.Authorities[genesis.Locality.ID] = AuthorityState{
		ActorID: genesis.Locality.ID, KeyID: genesis.Locality.Authority.KeyID,
		PublicKey: genesis.Locality.Authority.PublicKey, Role: "FORMATION",
		Capabilities: sortedCopy(formationAuthorityCapabilities), Status: "ACTIVE",
		MandateSHA256: genesis.Locality.SourceSHA256, RegisteredTransitionID: "GENESIS",
	}
	commitment, err := state.computeCommitment()
	if err != nil {
		return nil, err
	}
	state.StateCommitment = commitment
	return state, nil
}

func (s *RuntimeState) clone() (*RuntimeState, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	clone := new(RuntimeState)
	if err := decodeStrict(bytes, clone); err != nil {
		return nil, err
	}
	return clone, nil
}

func (s *RuntimeState) computeCommitment() (string, error) {
	copy := *s
	copy.StateCommitment = ""
	bytes, err := json.Marshal(&copy)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(bytes)
	return hex.EncodeToString(digest[:]), nil
}

func (s *RuntimeState) validateLoaded() error {
	if s.Schema != stateSchema {
		return fmt.Errorf("unsupported state schema %q", s.Schema)
	}
	if err := requireDigest("state_commitment", s.StateCommitment); err != nil {
		return err
	}
	computed, err := s.computeCommitment()
	if err != nil {
		return err
	}
	if computed != s.StateCommitment {
		return errors.New("persisted state commitment mismatch")
	}
	if s.ActorKeys == nil || s.NextNonces == nil || s.PresentationHistory == nil || s.EntryHistory == nil || s.DepartureCheckpoints == nil || s.ResidueHistory == nil || s.Loci == nil || s.Events == nil || s.Corrections == nil || s.IncorporatedResidues == nil || s.Authorities == nil || s.Pulses == nil || s.SuccessorProposals == nil {
		return errors.New("persisted state contains nil collections")
	}
	if s.Capacity.StructuralCeilingUnits == 0 || s.Capacity.DeclaredActualUnits > s.Capacity.StructuralCeilingUnits || s.Capacity.CorrectionEgressReserveUnits >= s.Capacity.StructuralCeilingUnits {
		return errors.New("persisted capacity envelope is invalid")
	}
	for actorID, authority := range s.Authorities {
		if authority.ActorID != actorID || s.ActorKeys[actorID] != authority.PublicKey {
			return errors.New("authority binding mismatch")
		}
	}
	for presentationID, presentation := range s.PresentationHistory {
		if presentation.PresentationID != presentationID {
			return errors.New("presentation history binding mismatch")
		}
	}
	for entryID, entry := range s.EntryHistory {
		if entry.EntryTransitionID != entryID || entry.ParticipantID == "" {
			return errors.New("entry history binding mismatch")
		}
		if _, exists := s.PresentationHistory[entry.PresentationID]; !exists {
			return errors.New("entry history references an unknown presentation")
		}
	}
	for checkpointID, checkpoint := range s.DepartureCheckpoints {
		if checkpoint.CheckpointID != checkpointID || checkpoint.Status != "OPEN" && checkpoint.Status != "SEALED" && checkpoint.Status != "CONSUMED" {
			return errors.New("departure checkpoint binding or status mismatch")
		}
		if _, exists := s.EntryHistory[checkpoint.EntryTransitionID]; !exists {
			return errors.New("departure checkpoint references an unknown entry")
		}
	}
	for residueID, residue := range s.ResidueHistory {
		if residue.ResidueID != residueID {
			return errors.New("residue history binding mismatch")
		}
		if _, exists := s.DepartureCheckpoints[residue.DepartureCheckpointID]; !exists {
			return errors.New("residue history references an unknown departure checkpoint")
		}
	}
	want := s.derivedBody()
	if s.Body != want {
		return errors.New("persisted body posture does not match current obligations")
	}
	_, err = s.capacitySnapshot()
	return err
}

func (s *RuntimeState) apply(genesis *Genesis, transition *Transition, transitionID string) (*Receipt, error) {
	if s.SuccessorBoundary != nil {
		return nil, errors.New("predecessor state is frozen at an activated succession boundary")
	}
	if _, exists := s.Events[transitionID]; exists {
		return nil, errors.New("transition already exists")
	}
	u := &transition.Unsigned
	if u.Revision != s.Revision+1 {
		return nil, fmt.Errorf("revision %d does not follow current revision %d", u.Revision, s.Revision)
	}
	if u.PreviousStateCommitment != s.StateCommitment {
		return nil, errors.New("previous_state_commitment does not name the current state")
	}
	if u.ObservedAt < s.LastObservedAt {
		return nil, errors.New("observed_at precedes the accepted local timeline")
	}
	if s.derivedBody().Posture == "HOLD_CAPACITY_DEFICIT" && !operationRemainsOpenInCapacityHold(u.Operation) {
		return nil, errors.New("capacity deficit places this motion in HOLD; correction, egress, currentness, and recovery remain open")
	}
	knownKey, actorKnown := s.ActorKeys[u.ActorID]
	if actorKnown && knownKey != u.ActorPublicKey {
		return nil, errors.New("actor public key differs from its established local binding")
	}
	if !actorKnown && u.Operation != opPresentForm {
		return nil, errors.New("unknown actor must first present a conformant form")
	}
	if !actorKnown && u.ActorID != operationalActorIDFromHex(u.ActorPublicKey) {
		return nil, errors.New("new participant actor_id must be derived from actor_public_key")
	}
	expectedNonce := s.NextNonces[u.ActorID]
	if u.Nonce != expectedNonce {
		return nil, fmt.Errorf("nonce %d does not match next nonce %d for %s", u.Nonce, expectedNonce, u.ActorID)
	}
	if err := s.applyOperation(genesis, transition, transitionID); err != nil {
		return nil, err
	}
	if !actorKnown {
		s.ActorKeys[u.ActorID] = u.ActorPublicKey
	}
	s.NextNonces[u.ActorID] = expectedNonce + 1
	s.Revision = u.Revision
	s.LastObservedAt = u.ObservedAt
	s.LastTransitionID = transitionID
	s.Events[transitionID] = EventRecord{transitionID, u.Revision, u.Operation, u.ActorID, u.LocusID, u.ObservedAt, u.Effect}
	s.Body = s.derivedBody()
	commitment, err := s.computeCommitment()
	if err != nil {
		return nil, err
	}
	s.StateCommitment = commitment
	snapshot, err := s.capacitySnapshot()
	if err != nil {
		return nil, err
	}
	return &Receipt{
		Schema: "PRESENCE_AVALANCHE_RECEIPT_001", TransitionID: transitionID, Revision: s.Revision,
		StateCommitment: s.StateCommitment, Operation: u.Operation, ActorID: u.ActorID, LocusID: u.LocusID,
		Effect: u.Effect, Posture: s.Body.Posture, PresenceCount: s.Body.PresenceCount,
		Capacity: snapshot, NonEffects: receiptNonEffects(u.Operation),
	}, nil
}

func operationRemainsOpenInCapacityHold(operation string) bool {
	switch operation {
	case opGateDisposition, opMatterDisposition, opCorrect, opCheckpointDeparture, opExit, opClose,
		opDeclareCapacity, opPulse, opReclaimOffer, opRegisterAuthority,
		opProposeSuccessor, opAttestSuccessor:
		return true
	default:
		return false
	}
}

func (s *RuntimeState) activeLocus(locusID string) (*LocusState, error) {
	if s.ActiveLocusID == "" {
		return nil, errors.New("no locus is active")
	}
	if locusID != s.ActiveLocusID {
		return nil, fmt.Errorf("transition addresses locus %q but active locus is %q", locusID, s.ActiveLocusID)
	}
	locus, exists := s.Loci[locusID]
	if !exists || locus.Phase != "OPEN" {
		return nil, errors.New("active locus is not open")
	}
	return locus, nil
}

func (s *RuntimeState) requireFormationAuthority(actorID string) error {
	if actorID != s.FormationAuthorityID {
		return errors.New("operation requires the explicit formation authority")
	}
	authority, exists := s.Authorities[actorID]
	if !exists || authority.Role != "FORMATION" || authority.Status != "ACTIVE" {
		return errors.New("formation authority is not active")
	}
	return nil
}

func (s *RuntimeState) requireCapability(actorID, capability string) error {
	authority, exists := s.Authorities[actorID]
	if !exists || authority.Status != "ACTIVE" || !contains(authority.Capabilities, capability) {
		return fmt.Errorf("operation requires active authority capability %s", capability)
	}
	return nil
}

func (s *RuntimeState) requireContinuityCapability(actorID, capability string) error {
	if err := s.requireCapability(actorID, capability); err != nil {
		return err
	}
	if s.Authorities[actorID].Role != "CONTINUITY" {
		return errors.New("operation requires an active continuity authority, not formation authority")
	}
	return nil
}

func (s *RuntimeState) isActiveAuthority(actorID string) bool {
	authority, exists := s.Authorities[actorID]
	return exists && authority.Status == "ACTIVE"
}

func (s *RuntimeState) capacitySnapshot() (CapacitySnapshot, error) {
	effective := s.Capacity.DeclaredActualUnits
	if s.ActiveLocusID != "" {
		locus := s.Loci[s.ActiveLocusID]
		if locus != nil && locus.CapacityCeilingUnits < effective {
			effective = locus.CapacityCeilingUnits
		}
	}
	var unresolved, resolution uint64
	for _, locus := range s.Loci {
		for _, entry := range locus.Entries {
			if entry.Status != "PRESENT" {
				continue
			}
			var err error
			unresolved, err = safeCapacityAdd(unresolved, entry.ReservedWorkUnits)
			if err != nil {
				return CapacitySnapshot{}, err
			}
			resolution, err = safeCapacityAdd(resolution, entry.ReservedResolutionUnits)
			if err != nil {
				return CapacitySnapshot{}, err
			}
		}
	}
	correctionEgress, err := safeCapacityAdd(s.Capacity.CorrectionEgressReserveUnits, resolution)
	if err != nil {
		return CapacitySnapshot{}, err
	}
	required, err := safeCapacityAdd(unresolved, correctionEgress)
	if err != nil {
		return CapacitySnapshot{}, err
	}
	available, deficit := uint64(0), uint64(0)
	if required <= effective {
		available = effective - required
	} else {
		deficit = required - effective
	}
	return CapacitySnapshot{
		StructuralCeilingUnits: s.Capacity.StructuralCeilingUnits,
		DeclaredActualUnits:    s.Capacity.DeclaredActualUnits, EffectiveActualUnits: effective,
		UnresolvedUnits: unresolved, CorrectionEgressUnits: correctionEgress,
		AvailableUnits: available, DeficitUnits: deficit,
		Rule: "C_AVAILABLE_EQUALS_C_ACTUAL_MINUS_C_UNRESOLVED_MINUS_C_CORRECTION_EGRESS",
	}, nil
}

func safeCapacityAdd(a, b uint64) (uint64, error) {
	if math.MaxUint64-a < b {
		return 0, errors.New("capacity arithmetic overflow")
	}
	return a + b, nil
}

func (s *RuntimeState) currentPresenceCount() uint64 {
	var count uint64
	for _, locus := range s.Loci {
		for _, entry := range locus.Entries {
			if entry.Status == "PRESENT" {
				count++
			}
		}
	}
	return count
}

func (s *RuntimeState) derivedBody() BodyState {
	if s.SuccessorBoundary != nil {
		return BodyState{"SUCCESSION_COMMITTED", 0, "PREDECESSOR_FROZEN_AT_EXPLICIT_BOUNDARY"}
	}
	presence := s.currentPresenceCount()
	snapshot, err := s.capacitySnapshot()
	if err != nil || snapshot.DeficitUnits > 0 {
		return BodyState{"HOLD_CAPACITY_DEFICIT", presence, "CORRECTION_AND_EGRESS_REMAIN_OPEN"}
	}
	if presence == 0 {
		return BodyState{"DORMANT_P0", 0, "CHECKPOINTED_WITHOUT_FABRICATED_ACTIVITY_REENTRY_OPEN"}
	}
	return BodyState{"ACTIVE_P1", presence, "CURRENT_ENTERED_BINDING_EXISTS"}
}

func participantPresent(locus *LocusState, actorID string) bool {
	entry, exists := locus.Entries[actorID]
	return exists && entry.Status == "PRESENT"
}

func receiptNonEffects(operation string) []string {
	common := []string{"NO_EXTERNAL_TRUTH_INFERENCE", "NO_SOURCE_OR_AUTHORITY_TRANSFER", "NO_SHARED_CURRENTNESS"}
	switch operation {
	case opPresentForm:
		return append(common, "NO_ADMISSION_BY_PRESENTATION", "NO_ENTRY_BY_PRESENTATION", "NO_FORMATION_BY_CONFORMANCE")
	case opGateDisposition:
		return append(common, "NO_ENTRY_BY_ADMISSION", "NO_CAPACITY_RESERVATION_BY_ADMISSION")
	case opEnter:
		return append(common, "NO_MATTER_ADMISSION_BY_ENTRY", "NO_TOKEN_BALANCE_AS_CAPACITY")
	case opCheckpointDeparture:
		return append(common, "NO_EXIT_BY_CHECKPOINT", "NO_CURRENTNESS_OR_CONSCIOUSNESS_PROOF", "NO_PRIVATE_STATE_DISCLOSURE")
	case opReenter:
		return append(common, "NO_RESTORATION_OR_RESUMPTION_OF_PRIOR_LOCUS", "NO_SEMANTIC_TRUTH_PROOF", "NO_PRIVATE_STATE_DISCLOSURE")
	case opObserveCrossing:
		return append(common, "NO_ADMISSION_BY_CROSSING", "MATTER_NOT_TRUTH")
	case opMatterDisposition:
		return append(common, "NO_OTHER_LOCALITY_UPTAKE")
	case opRecordEmergence:
		return append(common, "NO_AUTOMATIC_LOCAL_UPTAKE")
	case opCorrect:
		return append(common, "NO_PREDECESSOR_ERASURE")
	case opClose:
		return append(common, "NO_ERASURE_BY_CLOSURE", "NO_LOCAL_INCORPORATION_BY_CLOSURE")
	case opIncorporateResidue:
		return append(common, "NO_SOURCE_FIELD_MUTATION", "NO_OTHER_LOCALITY_UPTAKE")
	case opDeclareCapacity:
		return append(common, "NO_EXTERNAL_RESOURCE_PROOF", "NO_TOKEN_BALANCE_AS_CAPACITY")
	case opPulse:
		return append(common, "NO_EXTERNAL_LIVENESS_PROOF", "NO_HISTORICAL_RECEIPT_AS_CURRENT_PRESENCE", "NO_AUTONOMOUS_AUTHORITY")
	case opRegisterAuthority:
		return append(common, "NO_SOURCE_TRANSFER", "NO_AUTHORITY_SELF_UPGRADE")
	case opExhaustFormation:
		return append(common, "NO_AUTHORITY_TRANSFER", "NO_SUCCESSOR_INHERITANCE")
	case opProposeSuccessor, opAttestSuccessor:
		return append(common, "NO_SUCCESSOR_ACTIVATION")
	case opActivateSuccessor:
		return append(common, "NO_UNKNOWN_CODE_EXECUTION", "NO_AUTHORITY_INHERITANCE")
	default:
		return common
	}
}

func sortedCopy(values []string) []string {
	copy := slices.Clone(values)
	slices.Sort(copy)
	return copy
}

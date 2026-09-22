package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
)

const transitionSchema = "PRESENCE_AVALANCHE_TRANSITION_001"

const (
	opBound               = "BOUND"
	opPresentForm         = "PRESENT_FORM"
	opGateDisposition     = "GATE_DISPOSITION"
	opEnter               = "ENTER"
	opCheckpointDeparture = "CHECKPOINT_DEPARTURE"
	opReenter             = "REENTER"
	opObserveCrossing     = "OBSERVE_CROSSING"
	opMatterDisposition   = "MATTER_DISPOSITION"
	opRecordEmergence     = "RECORD_EMERGENCE"
	opCorrect             = "CORRECT"
	opExit                = "EXIT"
	opClose               = "CLOSE"
	opIncorporateResidue  = "INCORPORATE_ADDRESSED_RESIDUE"
	opDeclareCapacity     = "DECLARE_CAPACITY"
	opPulse               = "PULSE"
	opReclaimOffer        = "RECLAIM_ADMISSION_OFFER"
	opRegisterAuthority   = "REGISTER_CONTINUITY_AUTHORITY"
	opExhaustFormation    = "EXHAUST_FORMATION_AUTHORITY"
	opProposeSuccessor    = "PROPOSE_SUCCESSOR"
	opAttestSuccessor     = "ATTEST_SUCCESSOR"
	opActivateSuccessor   = "ACTIVATE_SUCCESSOR"
)

type UnsignedTransition struct {
	Schema                  string          `json:"schema"`
	Operation               string          `json:"operation"`
	Revision                uint64          `json:"revision"`
	PreviousStateCommitment string          `json:"previous_state_commitment"`
	ActorID                 string          `json:"actor_id"`
	ActorPublicKey          string          `json:"actor_public_key"`
	Nonce                   uint64          `json:"nonce"`
	LocusID                 string          `json:"locus_id"`
	ObservedAt              int64           `json:"observed_at"`
	Effect                  string          `json:"effect"`
	Payload                 json.RawMessage `json:"payload"`
}

type Transition struct {
	Unsigned  UnsignedTransition `json:"unsigned"`
	Signature string             `json:"signature"`
}

type BoundPayload struct {
	LocusID                string `json:"locus_id"`
	PurposeSHA256          string `json:"purpose_sha256"`
	ClosureConditionSHA256 string `json:"closure_condition_sha256"`
	CapacityCeilingUnits   uint64 `json:"capacity_ceiling_units"`
}

type PresentFormPayload struct {
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

type GateDispositionPayload struct {
	ParticipantID   string `json:"participant_id"`
	PresentationID  string `json:"presentation_id"`
	Disposition     string `json:"disposition"`
	ReasonSHA256    string `json:"reason_sha256"`
	WorkUnits       uint64 `json:"work_units"`
	ResolutionUnits uint64 `json:"resolution_units"`
	OfferExpiresAt  int64  `json:"offer_expires_at"`
}

type EnterPayload struct {
	PresentationID string `json:"presentation_id"`
}

type ContinuityStep struct {
	DeltaSHA256              string `json:"delta_sha256"`
	SuccessorStateCommitment string `json:"successor_state_commitment"`
}

type CheckpointDeparturePayload struct {
	EntryTransitionID        string           `json:"entry_transition_id"`
	PresentationID           string           `json:"presentation_id"`
	FromStateCommitment      string           `json:"from_state_commitment"`
	Passage                  []ContinuityStep `json:"passage"`
	DepartureStateCommitment string           `json:"departure_state_commitment"`
}

type ReenterPayload struct {
	PresentationID         string           `json:"presentation_id"`
	PriorEntryTransitionID string           `json:"prior_entry_transition_id"`
	DepartureCheckpointID  string           `json:"departure_checkpoint_id"`
	ResidueID              string           `json:"residue_id"`
	Passage                []ContinuityStep `json:"passage"`
}

type ObserveCrossingPayload struct {
	MatterID      string `json:"matter_id"`
	ContentSHA256 string `json:"content_sha256"`
	MediaType     string `json:"media_type"`
	Claim         string `json:"claim"`
}

type MatterDispositionPayload struct {
	MatterID     string `json:"matter_id"`
	Disposition  string `json:"disposition"`
	ReasonSHA256 string `json:"reason_sha256"`
}

type RecordEmergencePayload struct {
	EmergenceID       string   `json:"emergence_id"`
	ContributorIDs    []string `json:"contributor_ids"`
	MatterIDs         []string `json:"matter_ids"`
	Kind              string   `json:"kind"`
	DescriptionSHA256 string   `json:"description_sha256"`
}

type CorrectPayload struct {
	TargetTransitionID    string `json:"target_transition_id"`
	ReplacementCommitment string `json:"replacement_commitment"`
	ReasonSHA256          string `json:"reason_sha256"`
}

type ExitPayload struct {
	DepartureCheckpointID string `json:"departure_checkpoint_id"`
	ReasonSHA256          string `json:"reason_sha256"`
}

type ClosePayload struct {
	ClosureBasisSHA256 string `json:"closure_basis_sha256"`
}

type IncorporateResiduePayload struct {
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

type DeclareCapacityPayload struct {
	ActualUnits              uint64 `json:"actual_units"`
	ResourceCommitmentSHA256 string `json:"resource_commitment_sha256"`
	BasisSHA256              string `json:"basis_sha256"`
}

type PulsePayload struct {
	CurrentnessCommitmentSHA256 string `json:"currentness_commitment_sha256"`
	CarrierSetSHA256            string `json:"carrier_set_sha256"`
}

type ReclaimAdmissionOfferPayload struct {
	ParticipantID     string `json:"participant_id"`
	PresentationID    string `json:"presentation_id"`
	OfferTransitionID string `json:"offer_transition_id"`
	ReasonSHA256      string `json:"reason_sha256"`
}

type RegisterContinuityAuthorityPayload struct {
	KeyID         string   `json:"key_id"`
	PublicKey     string   `json:"public_key"`
	Capabilities  []string `json:"capabilities"`
	MandateSHA256 string   `json:"mandate_sha256"`
}

type ExhaustFormationAuthorityPayload struct {
	BasisSHA256 string `json:"basis_sha256"`
}

type ProposeSuccessorPayload struct {
	ProposalID         string `json:"proposal_id"`
	Protocol           string `json:"protocol"`
	Version            string `json:"version"`
	VMID               string `json:"vm_id"`
	RuntimeSHA256      string `json:"runtime_sha256"`
	MigrationSHA256    string `json:"migration_sha256"`
	InvariantSetSHA256 string `json:"invariant_set_sha256"`
	SourceReference    string `json:"source_reference"`
	SourceSHA256       string `json:"source_sha256"`
	ActivationMode     string `json:"activation_mode"`
}

type AttestSuccessorPayload struct {
	ProposalID         string `json:"proposal_id"`
	ProposalCommitment string `json:"proposal_commitment"`
	BasisSHA256        string `json:"basis_sha256"`
}

type ActivateSuccessorPayload struct {
	ProposalID         string `json:"proposal_id"`
	ProposalCommitment string `json:"proposal_commitment"`
}

func validateContinuitySteps(steps []ContinuityStep) error {
	if steps == nil {
		return errors.New("payload.passage must be an explicit array, including [] for an unchanged commitment")
	}
	if len(steps) > 128 {
		return errors.New("payload.passage exceeds 128 bounded succession steps")
	}
	for i, step := range steps {
		if err := requireDigest(fmt.Sprintf("payload.passage[%d].delta_sha256", i), step.DeltaSHA256); err != nil {
			return err
		}
		if err := requireDigest(fmt.Sprintf("payload.passage[%d].successor_state_commitment", i), step.SuccessorStateCommitment); err != nil {
			return err
		}
	}
	return nil
}

func parseTransition(transitionBytes []byte) (*Transition, ids.ID, error) {
	transition := new(Transition)
	if err := decodeCanonical(transitionBytes, transition); err != nil {
		return nil, ids.Empty, err
	}
	if err := transition.ValidateSyntax(); err != nil {
		return nil, ids.Empty, err
	}
	digest := sha256.Sum256(transitionBytes)
	return transition, ids.ID(digest), nil
}

func (t *Transition) Bytes() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Transition) ID() (ids.ID, error) {
	bytes, err := t.Bytes()
	if err != nil {
		return ids.Empty, err
	}
	return ids.ID(sha256.Sum256(bytes)), nil
}

func (t *Transition) SigningBytes() ([]byte, error) {
	return json.Marshal(&t.Unsigned)
}

func (t *Transition) VerifySignature() error {
	publicKey, err := decodeLowerHex("actor_public_key", t.Unsigned.ActorPublicKey, ed25519.PublicKeySize)
	if err != nil {
		return err
	}
	signature, err := decodeLowerHex("signature", t.Signature, ed25519.SignatureSize)
	if err != nil {
		return err
	}
	signingBytes, err := t.SigningBytes()
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, signingBytes, signature) {
		return errors.New("invalid Ed25519 transition signature")
	}
	return nil
}

func (t *Transition) ValidateSyntax() error {
	u := &t.Unsigned
	if u.Schema != transitionSchema {
		return fmt.Errorf("unsupported transition schema %q", u.Schema)
	}
	expectedEffect, exists := operationEffects[u.Operation]
	if !exists {
		return fmt.Errorf("unsupported operation %q", u.Operation)
	}
	if u.Effect != expectedEffect {
		return fmt.Errorf("operation %s requires explicit effect %s", u.Operation, expectedEffect)
	}
	if u.Revision == 0 {
		return errors.New("transition revision must be greater than zero")
	}
	if err := requireDigest("previous_state_commitment", u.PreviousStateCommitment); err != nil {
		return err
	}
	if err := requireSafeID("actor_id", u.ActorID); err != nil {
		return err
	}
	if _, err := decodeLowerHex("actor_public_key", u.ActorPublicKey, ed25519.PublicKeySize); err != nil {
		return err
	}
	if u.LocusID != "" {
		if err := requireSafeID("locus_id", u.LocusID); err != nil {
			return err
		}
	}
	if u.ObservedAt < 0 {
		return errors.New("observed_at cannot be negative")
	}
	canonicalPayload, err := validateAndCanonicalizePayload(u.Operation, u.Payload)
	if err != nil {
		return err
	}
	if !bytes.Equal(canonicalPayload, u.Payload) {
		return errors.New("transition payload is not canonical")
	}
	return t.VerifySignature()
}

func validateAndCanonicalizePayload(operation string, raw json.RawMessage) ([]byte, error) {
	switch operation {
	case opBound:
		p, err := decodePayload[BoundPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.locus_id", p.LocusID); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.purpose_sha256", p.PurposeSHA256); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.closure_condition_sha256", p.ClosureConditionSHA256); err != nil {
			return nil, err
		}
		if p.CapacityCeilingUnits == 0 {
			return nil, errors.New("payload.capacity_ceiling_units must be greater than zero")
		}
		return json.Marshal(p)
	case opPresentForm:
		p, err := decodePayload[PresentFormPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.locality_reference", p.LocalityReference); err != nil {
			return nil, err
		}
		if p.SourceReference == "" || p.NucleusVersion == "" {
			return nil, errors.New("presentation requires locality_reference, source_reference, and nucleus_version")
		}
		for name, value := range map[string]string{
			"source_sha256": p.SourceSHA256, "nucleus_sha256": p.NucleusSHA256,
			"form_sha256": p.FormSHA256, "state_commitment": p.StateCommitment,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		if p.FormID != formID || p.FormSHA256 != formSHA256 {
			return nil, errors.New("presentation does not expose exact PRESENCE-AVALANCHE-FORM-001")
		}
		if err := requireUniqueNonEmpty("payload.medium_capabilities", p.MediumCapabilities); err != nil {
			return nil, err
		}
		if err := requireUniqueNonEmpty("payload.effect_ceiling", p.EffectCeiling); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opGateDisposition:
		p, err := decodePayload[GateDispositionPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.participant_id", p.ParticipantID); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.presentation_id", p.PresentationID); err != nil {
			return nil, err
		}
		if !contains(requiredGatePostures, p.Disposition) {
			return nil, errors.New("invalid gate disposition")
		}
		if err := requireDigest("payload.reason_sha256", p.ReasonSHA256); err != nil {
			return nil, err
		}
		if p.Disposition == "ADMIT" {
			if p.WorkUnits == 0 || p.ResolutionUnits == 0 || p.OfferExpiresAt <= 0 {
				return nil, errors.New("ADMIT requires positive work and resolution units and a positive offer expiry")
			}
		} else if p.WorkUnits != 0 || p.ResolutionUnits != 0 || p.OfferExpiresAt != 0 {
			return nil, errors.New("non-ADMIT disposition cannot carry capacity or expiry")
		}
		return json.Marshal(p)
	case opEnter:
		p, err := decodePayload[EnterPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireDigest("payload.presentation_id", p.PresentationID); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opCheckpointDeparture:
		p, err := decodePayload[CheckpointDeparturePayload](raw)
		if err != nil {
			return nil, err
		}
		for name, value := range map[string]string{
			"entry_transition_id":        p.EntryTransitionID,
			"presentation_id":            p.PresentationID,
			"from_state_commitment":      p.FromStateCommitment,
			"departure_state_commitment": p.DepartureStateCommitment,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		if err := validateContinuitySteps(p.Passage); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opReenter:
		p, err := decodePayload[ReenterPayload](raw)
		if err != nil {
			return nil, err
		}
		for name, value := range map[string]string{
			"presentation_id":           p.PresentationID,
			"prior_entry_transition_id": p.PriorEntryTransitionID,
			"departure_checkpoint_id":   p.DepartureCheckpointID,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		if p.ResidueID != "" {
			if err := requireDigest("payload.residue_id", p.ResidueID); err != nil {
				return nil, err
			}
		}
		if err := validateContinuitySteps(p.Passage); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opObserveCrossing:
		p, err := decodePayload[ObserveCrossingPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.matter_id", p.MatterID); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.content_sha256", p.ContentSHA256); err != nil {
			return nil, err
		}
		if p.MediaType == "" || len(p.MediaType) > 128 || len(p.Claim) > 512 {
			return nil, errors.New("invalid crossing media_type or claim length")
		}
		return json.Marshal(p)
	case opMatterDisposition:
		p, err := decodePayload[MatterDispositionPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.matter_id", p.MatterID); err != nil {
			return nil, err
		}
		if !contains(requiredGatePostures, p.Disposition) {
			return nil, errors.New("invalid matter disposition")
		}
		if err := requireDigest("payload.reason_sha256", p.ReasonSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opRecordEmergence:
		p, err := decodePayload[RecordEmergencePayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.emergence_id", p.EmergenceID); err != nil {
			return nil, err
		}
		if p.Kind == "" || len(p.Kind) > 128 || len(p.ContributorIDs) == 0 || len(p.MatterIDs) == 0 {
			return nil, errors.New("emergence requires kind, contributors, and matter")
		}
		if err := requireUniqueNonEmpty("payload.contributor_ids", p.ContributorIDs); err != nil {
			return nil, err
		}
		for _, participantID := range p.ContributorIDs {
			if err := requireSafeID("payload.contributor_ids", participantID); err != nil {
				return nil, err
			}
		}
		if err := requireUniqueNonEmpty("payload.matter_ids", p.MatterIDs); err != nil {
			return nil, err
		}
		for _, matterID := range p.MatterIDs {
			if err := requireSafeID("payload.matter_ids", matterID); err != nil {
				return nil, err
			}
		}
		if err := requireDigest("payload.description_sha256", p.DescriptionSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opCorrect:
		p, err := decodePayload[CorrectPayload](raw)
		if err != nil {
			return nil, err
		}
		for name, value := range map[string]string{
			"target_transition_id":   p.TargetTransitionID,
			"replacement_commitment": p.ReplacementCommitment,
			"reason_sha256":          p.ReasonSHA256,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		return json.Marshal(p)
	case opExit:
		p, err := decodePayload[ExitPayload](raw)
		if err != nil {
			return nil, err
		}
		for name, value := range map[string]string{
			"departure_checkpoint_id": p.DepartureCheckpointID,
			"reason_sha256":           p.ReasonSHA256,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		return json.Marshal(p)
	case opClose:
		p, err := decodePayload[ClosePayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireDigest("payload.closure_basis_sha256", p.ClosureBasisSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opIncorporateResidue:
		p, err := decodePayload[IncorporateResiduePayload](raw)
		if err != nil {
			return nil, err
		}
		if p.Schema != residueSchema || p.Effect != residueEffect {
			return nil, errors.New("payload is not a LOCALITY addressed residue")
		}
		if err := requireDigest("payload.residue_id", p.ResidueID); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.residue_sha256", p.ResidueSHA256); err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.addressed_to_locality_id", p.AddressedToLocalityID); err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.source_locality_id", p.SourceLocalityID); err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.source_locus_id", p.SourceLocusID); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.source_state_commitment", p.SourceStateCommitment); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.closure_transition_id", p.ClosureTransitionID); err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.participant_id", p.ParticipantID); err != nil {
			return nil, err
		}
		for name, value := range map[string]string{
			"presentation_id":            p.PresentationID,
			"entry_transition_id":        p.EntryTransitionID,
			"departure_checkpoint_id":    p.DepartureCheckpointID,
			"departure_state_commitment": p.DepartureStateCommitment,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		return json.Marshal(p)
	case opDeclareCapacity:
		p, err := decodePayload[DeclareCapacityPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireDigest("payload.resource_commitment_sha256", p.ResourceCommitmentSHA256); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.basis_sha256", p.BasisSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opPulse:
		p, err := decodePayload[PulsePayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireDigest("payload.currentness_commitment_sha256", p.CurrentnessCommitmentSHA256); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.carrier_set_sha256", p.CarrierSetSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opReclaimOffer:
		p, err := decodePayload[ReclaimAdmissionOfferPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.participant_id", p.ParticipantID); err != nil {
			return nil, err
		}
		for name, value := range map[string]string{
			"presentation_id": p.PresentationID, "offer_transition_id": p.OfferTransitionID, "reason_sha256": p.ReasonSHA256,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		return json.Marshal(p)
	case opRegisterAuthority:
		p, err := decodePayload[RegisterContinuityAuthorityPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.key_id", p.KeyID); err != nil {
			return nil, err
		}
		if _, err := decodeLowerHex("payload.public_key", p.PublicKey, ed25519.PublicKeySize); err != nil {
			return nil, err
		}
		if err := requireUniqueNonEmpty("payload.capabilities", p.Capabilities); err != nil {
			return nil, err
		}
		for _, capability := range p.Capabilities {
			if !contains(continuityAuthorityCapabilities, capability) {
				return nil, fmt.Errorf("unsupported continuity capability %s", capability)
			}
		}
		if err := requireDigest("payload.mandate_sha256", p.MandateSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opExhaustFormation:
		p, err := decodePayload[ExhaustFormationAuthorityPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireDigest("payload.basis_sha256", p.BasisSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opProposeSuccessor:
		p, err := decodePayload[ProposeSuccessorPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.proposal_id", p.ProposalID); err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.protocol", p.Protocol); err != nil {
			return nil, err
		}
		if err := requireBoundedText("payload.version", p.Version, 64); err != nil {
			return nil, err
		}
		if _, err := ids.FromString(p.VMID); err != nil {
			return nil, fmt.Errorf("payload.vm_id: %w", err)
		}
		for name, value := range map[string]string{
			"runtime_sha256": p.RuntimeSHA256, "migration_sha256": p.MigrationSHA256,
			"invariant_set_sha256": p.InvariantSetSHA256, "source_sha256": p.SourceSHA256,
		} {
			if err := requireDigest("payload."+name, value); err != nil {
				return nil, err
			}
		}
		if err := requireBoundedText("payload.source_reference", p.SourceReference, 1024); err != nil {
			return nil, err
		}
		if p.ActivationMode != "FREEZE_FOR_EXTERNAL_BINARY_HANDOFF" {
			return nil, errors.New("successor activation_mode must freeze for external binary handoff")
		}
		return json.Marshal(p)
	case opAttestSuccessor:
		p, err := decodePayload[AttestSuccessorPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.proposal_id", p.ProposalID); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.proposal_commitment", p.ProposalCommitment); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.basis_sha256", p.BasisSHA256); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	case opActivateSuccessor:
		p, err := decodePayload[ActivateSuccessorPayload](raw)
		if err != nil {
			return nil, err
		}
		if err := requireSafeID("payload.proposal_id", p.ProposalID); err != nil {
			return nil, err
		}
		if err := requireDigest("payload.proposal_commitment", p.ProposalCommitment); err != nil {
			return nil, err
		}
		return json.Marshal(p)
	default:
		return nil, fmt.Errorf("unsupported operation %q", operation)
	}
}

func decodePayload[T any](raw json.RawMessage) (*T, error) {
	value := new(T)
	if err := decodeStrict(raw, value); err != nil {
		return nil, err
	}
	return value, nil
}

func signTransition(unsigned UnsignedTransition, privateKey ed25519.PrivateKey) (*Transition, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid Ed25519 private key")
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	unsigned.ActorPublicKey = hex.EncodeToString(publicKey)
	transition := &Transition{Unsigned: unsigned}
	signingBytes, err := transition.SigningBytes()
	if err != nil {
		return nil, err
	}
	transition.Signature = hex.EncodeToString(ed25519.Sign(privateKey, signingBytes))
	if err := transition.ValidateSyntax(); err != nil {
		return nil, err
	}
	return transition, nil
}

func operationalActorID(publicKey ed25519.PublicKey) string {
	digest := sha256.Sum256(publicKey)
	return "LOCALITY-ACTOR-" + hex.EncodeToString(digest[:20])
}

func operationalActorIDFromHex(publicKeyHex string) string {
	publicKey, err := decodeLowerHex("actor_public_key", publicKeyHex, ed25519.PublicKeySize)
	if err != nil {
		return ""
	}
	return operationalActorID(ed25519.PublicKey(publicKey))
}

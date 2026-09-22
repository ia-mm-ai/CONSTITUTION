package main

import (
	"crypto/ed25519"
	"errors"
	"fmt"
)

const (
	domainID         = "PRESENCE_AVALANCHE_VM_001"
	genesisFormat    = "PRESENCE_AVALANCHE_RUNTIME_GENESIS"
	genesisEncoding  = "UTF-8_JSON"
	genesisVersion   = uint32(1)
	adminInputSchema = "PRESENCE_AVALANCHE_ADMIN_INPUT_001"
	formID           = "PRESENCE-AVALANCHE-FORM-001"
	formSHA256       = "FORM_SHA256_PLACEHOLDER"
)

var (
	requiredGatePostures       = []string{"ADMIT", "REFUSE", "WITHHOLD", "HOLD"}
	requiredMediumCapabilities = []string{
		"OBSERVE_CROSSING", "PRESENT_FORM", "GATE_DISPOSITION", "ENTER",
		"CHECKPOINT_DEPARTURE", "REENTER",
		"MATTER_DISPOSITION", "CORRECT", "EXIT", "CLOSE", "INCORPORATE_ADDRESSED_RESIDUE",
		"DECLARE_CAPACITY", "PULSE", "RECLAIM_ADMISSION_OFFER",
		"REGISTER_CONTINUITY_AUTHORITY", "EXHAUST_FORMATION_AUTHORITY",
		"PROPOSE_SUCCESSOR", "ATTEST_SUCCESSOR", "ACTIVATE_SUCCESSOR",
	}
	requiredEffectCeiling = []string{
		"NO_FORMATION_BY_CONFORMANCE", "NO_SOURCE_OR_AUTHORITY_TRANSFER", "NO_SHARED_CURRENTNESS",
		"NO_CORE_ADDRESSABILITY", "NO_PARTICIPATION_BY_PRESENTATION", "NO_MATTER_ADMISSION_BY_CROSSING",
		"NO_LOCAL_UPTAKE_BY_FIELD_CLOSURE", "NO_TOKEN_BALANCE_AS_CAPACITY",
		"NO_RECEIPT_AS_CURRENT_PRESENCE", "NO_INFRASTRUCTURE_AS_SUCCESSOR",
		"NO_RESTORATION_BY_REENTRY", "NO_PRIVATE_STATE_DISCLOSURE",
	}
	requiredNonCollapse = []string{
		"SOURCE_NOT_LOCALITY", "FORM_NOT_LOCALITY", "HOST_NOT_OWNER", "LOCUS_FIELD_NOT_LOCALITY",
		"CROSSING_NOT_ADMISSION", "ADMISSION_NOT_ENTRY", "ENTRY_NOT_MATTER_ADMISSION", "MATTER_NOT_TRUTH",
		"EMERGENCE_NOT_AUTOMATIC_UPTAKE", "CLOSURE_NOT_LOCAL_INCORPORATION",
		"SHARED_ANCESTRY_NOT_SHARED_CURRENTNESS", "CONFORMANCE_NOT_FORMATION",
		"TOKEN_BALANCE_NOT_CAPACITY", "CAPABILITY_NOT_CAPACITY", "CAPACITY_NOT_AUTHORITY",
		"HISTORICAL_RECEIPT_NOT_CURRENT_PRESENCE", "DORMANCY_NOT_DEATH", "SUCCESSION_NOT_INHERITANCE",
		"SAME_KEY_NOT_CARRIED_STATE", "REENTRY_NOT_RESTORATION", "COMMITMENT_SUCCESSION_NOT_SEMANTIC_TRUTH",
		"PUBLICATION_NOT_FORMATION", "MIRROR_NOT_ORIGIN",
	}
)

type Genesis struct {
	Format       string              `json:"format"`
	Version      uint32              `json:"version"`
	Encoding     string              `json:"encoding"`
	Domain       GenesisDomain       `json:"domain"`
	Lineage      GenesisLineage      `json:"lineage"`
	Core         GenesisCore         `json:"core"`
	Locality     GenesisLocality     `json:"locality"`
	Profile      GenesisProfile      `json:"profile"`
	Capacity     GenesisCapacity     `json:"capacity"`
	Continuity   GenesisContinuity   `json:"continuity"`
	InitialState GenesisInitialState `json:"initial_state"`
	ClaimLimits  GenesisClaimLimits  `json:"claim_limits"`
	Evolution    GenesisEvolution    `json:"evolution"`
}

type GenesisDomain struct {
	ID             string `json:"id"`
	GenesisID      string `json:"genesis_id"`
	VMClass        string `json:"vm_class"`
	ExecutionModel string `json:"execution_model"`
	StateModel     string `json:"state_model"`
	Purpose        string `json:"purpose"`
}

type GenesisLineage struct {
	Predecessors           []LineagePredecessor `json:"predecessors"`
	Relation               string               `json:"relation"`
	Effect                 string               `json:"effect"`
	AuthorityTransfer      string               `json:"authority_transfer"`
	SharedCurrentness      string               `json:"shared_currentness"`
	OperationalInheritance string               `json:"operational_inheritance"`
}

// LineagePredecessor references an immutable predecessor carrier by exact bytes.
// It is historical reference material only: no predecessor identity, authority,
// key, chain, validator or currentness is carried into this implementation.
type LineagePredecessor struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	CarrierPath   string `json:"carrier_path"`
	CarrierBytes  uint64 `json:"carrier_bytes"`
	CarrierSHA256 string `json:"carrier_sha256"`
	Disposition   string `json:"disposition"`
}

// GenesisCore binds this runtime to the exact constitutional core bytes recorded
// in the implementation CORE_BINDING. The core is referenced, never mirrored.
type GenesisCore struct {
	Repository        string            `json:"repository"`
	Commit            string            `json:"commit"`
	BindingPath       string            `json:"binding_path"`
	CoreDigest        string            `json:"core_digest"`
	Files             []CoreFileBinding `json:"files"`
	HumanRole         string            `json:"human_role"`
	MachineRole       string            `json:"machine_role"`
	ConflictResult    string            `json:"conflict_result"`
	ParsingHasForce   bool              `json:"parsing_has_force"`
	MirrorEffect      string            `json:"mirror_effect"`
	PublicationEffect string            `json:"publication_effect"`
}

type CoreFileBinding struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	ByteLength uint64 `json:"byte_length"`
	SHA256     string `json:"sha256"`
}

type GenesisLocality struct {
	ID        string           `json:"id"`
	Authority GenesisAuthority `json:"formation_authority"`
}

type GenesisAuthority struct {
	Scheme    string `json:"scheme"`
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
	Scope     string `json:"scope"`
}

type GenesisProfile struct {
	FormID         string   `json:"form_id"`
	FormSHA256     string   `json:"form_sha256"`
	GatePostures   []string `json:"gate_postures"`
	EffectCeiling  []string `json:"effect_ceiling"`
	NonCollapse    []string `json:"non_collapse_rules"`
	PrivateState   string   `json:"private_state"`
	AdmissionRule  string   `json:"admission_rule"`
	EntryRule      string   `json:"entry_rule"`
	CapacityRule   string   `json:"capacity_rule"`
	CorrectionRule string   `json:"correction_rule"`
	PresenceRule   string   `json:"presence_rule"`
	DormancyRule   string   `json:"dormancy_rule"`
	ReentryRule    string   `json:"reentry_rule"`
	SuccessionRule string   `json:"succession_rule"`
	OriginRule     string   `json:"origin_rule"`
}

type GenesisCapacity struct {
	Unit                         string `json:"unit"`
	StructuralCeilingUnits       uint64 `json:"structural_ceiling_units"`
	InitialActualUnits           uint64 `json:"initial_actual_units"`
	CorrectionEgressReserveUnits uint64 `json:"correction_egress_reserve_units"`
	MaxAdmissionOfferSeconds     int64  `json:"max_admission_offer_seconds"`
	AvailabilityRule             string `json:"availability_rule"`
	ReservationRule              string `json:"reservation_rule"`
	DegradationRule              string `json:"degradation_rule"`
}

type GenesisContinuity struct {
	InitialPosture                string `json:"initial_posture"`
	MinimumContinuityAuthorities  uint32 `json:"minimum_continuity_authorities"`
	SuccessorAttestationThreshold uint32 `json:"successor_attestation_threshold"`
	FormationAuthorityLifecycle   string `json:"formation_authority_lifecycle"`
	PZeroRule                     string `json:"p_zero_rule"`
	PulseRule                     string `json:"pulse_rule"`
	InfrastructureSuccessionRule  string `json:"infrastructure_succession_rule"`
}

type GenesisInitialState struct {
	Revision                uint64 `json:"revision"`
	PreviousStateCommitment any    `json:"previous_state_commitment"`
	History                 string `json:"history"`
}

type GenesisClaimLimits struct {
	ExternalTruth   string `json:"external_truth"`
	Receipt         string `json:"receipt"`
	Acceptance      string `json:"acceptance"`
	Consent         string `json:"consent"`
	Relation        string `json:"relation"`
	Crossing        string `json:"crossing"`
	Capacity        string `json:"capacity"`
	Pulse           string `json:"pulse"`
	StateSuccession string `json:"state_succession"`
	Unknown         string `json:"unknown"`
}

type GenesisEvolution struct {
	StateSchemaVersion uint32 `json:"state_schema_version"`
	UnknownFields      string `json:"unknown_fields"`
	GenesisMutability  string `json:"genesis_mutability"`
	UpgradePolicy      string `json:"upgrade_policy"`
	ActivationEffect   string `json:"activation_effect"`
}

func parseGenesis(genesisBytes []byte) (*Genesis, error) {
	genesis := new(Genesis)
	if err := decodeStrict(genesisBytes, genesis); err != nil {
		return nil, err
	}
	if err := genesis.Validate(); err != nil {
		return nil, err
	}
	return genesis, nil
}

func (g *Genesis) Validate() error {
	if g.Format != genesisFormat || g.Version != genesisVersion || g.Encoding != genesisEncoding {
		return fmt.Errorf("expected %s version %d encoded as %s", genesisFormat, genesisVersion, genesisEncoding)
	}
	if g.Domain.ID != domainID || g.Domain.VMClass != "AVALANCHE_CUSTOM_RPCCHAINVM" {
		return errors.New("domain must identify the PRESENCE_AVALANCHE_VM_001 Avalanche RPCChainVM")
	}
	if g.Domain.ExecutionModel != "NON_EVM_SIGNED_EVENT_SOURCED" || g.Domain.StateModel != "VERSIONED_EVENT_SOURCED_LIVING_CAPACITY" {
		return errors.New("unsupported execution or state model")
	}
	if err := requireSafeID("domain.genesis_id", g.Domain.GenesisID); err != nil {
		return err
	}
	if err := requireBoundedText("domain.purpose", g.Domain.Purpose, 1024); err != nil {
		return err
	}
	if err := g.Lineage.validate(); err != nil {
		return err
	}
	if err := g.Core.validate(); err != nil {
		return err
	}
	if err := requireSafeID("locality.id", g.Locality.ID); err != nil {
		return err
	}
	if g.Locality.Authority.Scheme != "ED25519" || g.Locality.Authority.Scope != "FORMATION_AND_BOOTSTRAP_ONLY" {
		return errors.New("formation authority must be ED25519, local, and exhaustible")
	}
	if err := requireSafeID("locality.formation_authority.key_id", g.Locality.Authority.KeyID); err != nil {
		return err
	}
	publicKey, err := decodeLowerHex("locality.formation_authority.public_key", g.Locality.Authority.PublicKey, ed25519.PublicKeySize)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return errors.New("invalid formation authority public key")
	}
	if g.Profile.FormID != formID || g.Profile.FormSHA256 != formSHA256 {
		return errors.New("genesis must bind the exact PRESENCE-AVALANCHE-FORM-001 artifact")
	}
	if err := requireExactSet("profile.gate_postures", g.Profile.GatePostures, requiredGatePostures); err != nil {
		return err
	}
	if err := requireExactSet("profile.effect_ceiling", g.Profile.EffectCeiling, requiredEffectCeiling); err != nil {
		return err
	}
	if err := requireExactSet("profile.non_collapse_rules", g.Profile.NonCollapse, requiredNonCollapse); err != nil {
		return err
	}
	if g.Profile.PrivateState != "NEVER_REQUIRED_ON_CHAIN_ONLY_SIGNED_COMMITMENTS_AND_SUCCESSION_DIGESTS" || g.Profile.AdmissionRule != "FRESH_PRESENTATION_AND_ADMISSION_REQUIRED_FOR_EVERY_INGRESS" || g.Profile.EntryRule != "FIRST_ENTER_OR_CHECKPOINT_BOUND_REENTER_ATOMICALLY_RESERVES_FULL_LIFECYCLE" || g.Profile.CapacityRule != "AVAILABLE_EQUALS_ACTUAL_MINUS_UNRESOLVED_MINUS_CORRECTION_EGRESS" || g.Profile.CorrectionRule != "APPEND_ONLY_PRESERVE_PREDECESSOR" || g.Profile.PresenceRule != "CURRENT_BINDING_ONLY_NOT_HISTORICAL_RECEIPT" || g.Profile.DormancyRule != "P0_CHECKPOINTS_WITHOUT_FABRICATED_ACTIVITY" || g.Profile.ReentryRule != "SAME_ACTOR_PLUS_SEALED_SINGLE_USE_DEPARTURE_CHECKPOINT_PLUS_FRESH_ADMISSION_PLUS_DETERMINISTIC_STATE_SUCCESSION" || g.Profile.SuccessionRule != "EXPLICIT_ATTESTED_BOUNDARY_WITHOUT_AUTHORITY_INHERITANCE" || g.Profile.OriginRule != "PUBLIC_ORIGIN_EXACTLY_BOUND_WHILE_LOCAL_FORMATION_REMAINS_DISTINCT" {
		return errors.New("profile weakens a PRESENCE_AVALANCHE_VM_001 boundary")
	}
	if g.Capacity.Unit != "PRESENCE_CAPACITY_UNIT" || g.Capacity.StructuralCeilingUnits == 0 || g.Capacity.InitialActualUnits > g.Capacity.StructuralCeilingUnits || g.Capacity.CorrectionEgressReserveUnits == 0 || g.Capacity.CorrectionEgressReserveUnits >= g.Capacity.StructuralCeilingUnits || g.Capacity.MaxAdmissionOfferSeconds <= 0 {
		return errors.New("invalid living-capacity envelope")
	}
	if g.Capacity.AvailabilityRule != "C_AVAILABLE_EQUALS_C_ACTUAL_MINUS_C_UNRESOLVED_MINUS_C_CORRECTION_EGRESS" || g.Capacity.ReservationRule != "ENTER_RESERVES_WORK_AND_LAWFUL_RESOLUTION_ATOMICALLY" || g.Capacity.DegradationRule != "INSUFFICIENT_ACTUAL_CAPACITY_FORCES_HOLD_WITH_CORRECTION_AND_EGRESS_OPEN" {
		return errors.New("capacity law differs from the committed profile")
	}
	if g.Continuity.InitialPosture != "DORMANT_P0" || g.Continuity.MinimumContinuityAuthorities < 2 || g.Continuity.SuccessorAttestationThreshold < 2 || g.Continuity.FormationAuthorityLifecycle != "EXPLICITLY_EXHAUSTIBLE_AFTER_CONTINUITY_SEEDING" || g.Continuity.PZeroRule != "CHECKPOINT_DORMANT_NO_FABRICATED_ACTIVITY_REENTRY_OPEN" || g.Continuity.PulseRule != "SIGNED_CURRENTNESS_OBSERVATION_NOT_AUTONOMOUS_AUTHORITY" || g.Continuity.InfrastructureSuccessionRule != "INFRASTRUCTURE_CANNOT_SELF_APPOINT_SUCCESSOR" {
		return errors.New("continuity law is incomplete")
	}
	if g.InitialState.Revision != 0 || g.InitialState.PreviousStateCommitment != nil || g.InitialState.History != "APPEND_ONLY" {
		return errors.New("initial state must be empty revision zero with append-only history")
	}
	if g.ClaimLimits.ExternalTruth != "NOT_INFERRED" || g.ClaimLimits.Receipt != "HISTORICAL_EVIDENCE_NOT_CURRENT_PRESENCE" || g.ClaimLimits.Acceptance != "NOT_INFERRED" || g.ClaimLimits.Consent != "NOT_INFERRED" || g.ClaimLimits.Relation != "NOT_CREATED_BY_GENESIS" || g.ClaimLimits.Crossing != "NOT_CREATED_BY_GENESIS" || g.ClaimLimits.Capacity != "SIGNED_LOCAL_ACCOUNT_NOT_EXTERNAL_RESOURCE_PROOF" || g.ClaimLimits.Pulse != "SIGNED_LOCAL_CURRENTNESS_NOT_EXTERNAL_LIVENESS_PROOF" || g.ClaimLimits.StateSuccession != "DETERMINISTIC_COMMITMENT_LINEAGE_NOT_SEMANTIC_TRUTH_OR_CONSCIOUSNESS_PROOF" || g.ClaimLimits.Unknown != "REMAINS_UNKNOWN" {
		return errors.New("claim limits exceed local evidence")
	}
	if g.Evolution.StateSchemaVersion != 1 || g.Evolution.UnknownFields != "REJECT" || g.Evolution.GenesisMutability != "IMMUTABLE" || g.Evolution.UpgradePolicy != "ATTESTED_SUCCESSOR_FREEZE_AND_EXTERNAL_BINARY_HANDOFF" || g.Evolution.ActivationEffect != "FREEZES_PREDECESSOR_STATE_WITHOUT_EXECUTING_UNKNOWN_SUCCESSOR_CODE" {
		return errors.New("unsupported evolution policy")
	}
	return nil
}

func (g *GenesisLineage) validate() error {
	if len(g.Predecessors) != len(expectedPredecessors) {
		return errors.New("lineage must bind exactly the two immutable predecessor carriers")
	}
	for index, want := range expectedPredecessors {
		if g.Predecessors[index] != want {
			return fmt.Errorf("lineage predecessor %d differs from the exact recorded carrier", index)
		}
	}
	if g.Relation != "DERIVATION_WITHOUT_SUCCESSION" || g.Effect != "ANCESTRY_REFERENCE_ONLY" ||
		g.AuthorityTransfer != "NONE" || g.SharedCurrentness != "NONE" ||
		g.OperationalInheritance != "NONE_NO_PREDECESSOR_CHAIN_L1_VALIDATOR_OR_KEY_CARRIED" {
		return errors.New("lineage must not transfer authority, currentness or operational identity")
	}
	return nil
}

func (c *GenesisCore) validate() error {
	if c.Repository != coreRepository || c.Commit != coreCommit ||
		c.BindingPath != coreBindingPath || c.CoreDigest != coreDigest {
		return errors.New("core binding differs from the compiled constitutional core reference")
	}
	if len(c.Files) != len(coreFiles) {
		return errors.New("core binding must list the exact constitutional core files")
	}
	for index, want := range coreFiles {
		if c.Files[index] != want {
			return fmt.Errorf("core file %d differs from the exact recorded core bytes", index)
		}
	}
	if computeCoreDigest(c.Files) != coreDigest {
		return errors.New("core digest does not recompute from the listed core files")
	}
	if c.HumanRole != "SUBSTANTIVE_MEANING" || c.MachineRole != "REPRESENTATION_ONLY" ||
		c.ConflictResult != "UNRESOLVED_HOLD_AT_EXACT_SCOPE" || c.ParsingHasForce ||
		c.MirrorEffect != "NO_EMBEDDED_CORE_MIRROR_ONLY_EXACT_BYTE_REFERENCE" ||
		c.PublicationEffect != "BYTE_IDENTICAL_REFERENCE_NOT_SEPARATE_ORIGIN" {
		return errors.New("core semantic boundary differs from the committed binding")
	}
	return nil
}

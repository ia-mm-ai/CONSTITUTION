package main

import (
	"crypto/ed25519"
	"errors"
	"fmt"
)

const (
	domainID                   = "PRESENCE_AVALANCHE_VM"
	genesisFormat              = "PRESENCE_RUNTIME_GENESIS"
	genesisEncoding            = "UTF-8_JSON"
	genesisVersion             = uint32(4)
	deploymentDescriptorSchema = "PRESENCE_DEPLOYMENT_DESCRIPTOR_001"
	formID                     = "PRESENCE-AVALANCHE-FORM-001"
	formSHA256                 = "10848ac3fc2f68f6fdc041a4c9862d8fb186f4c3d0ebd4edde3f3933fe3ce893"

	directPredecessorProtocol         = "LOCALITY_VM_002"
	directPredecessorVersion          = "2.1.0"
	directPredecessorVMID             = "2JnfZqeNUW34DmiSznMp1oJSpX5Fpqv9Ms6BYEyGF3LYVx7jQ5"
	directPredecessorRepositorySHA256 = "bffc125a5873a42b8ad81ede898ae7d65e36666870af3f0e6e6e5d7bdf2ddacb"
	directPredecessorSidecarSHA256    = "a7b5b4b922943cef87e04e140b63c430cb90f795dfdb40f8d72aec4b5e02ce7c"
	directPredecessorReleaseSHA256    = "7586bdd9f31bb120b608414d5428d2435855d76c7dfd011d9f036eec91590add"
	directPredecessorEvidenceSHA256   = "78d2403fa9847cee82235ded7d6aca32572c8da854a55ea7d95169fe21393a84"
	directPredecessorStateCommitment  = "38fd03c9d2bbb0ccd18585d71ce7f0573f2967a2289dbd60fd513cd782cf7d3f"
	directPredecessorStateSHA256      = "6990892e7d7d80d7bec72e03caddb74d6fe0c5e916aa732aca8e98cf1b4ec8e3"

	predecessorChainID         = "LB6wwV4JNxr8fwjUPBHMzf3PiW2d4hsTc4v63uX1MjY6oZ9Wb"
	predecessorL1ID            = "21mJfY4QpDeVykBaeG8nwn7k7w7b5oPpPhaYcJWqumht8SvaK"
	predecessorVMID            = "pJHx1NU8ghWsiwg1vrqwaqE5uKh1k4EJRBkpB1tKV1QuQFhMi"
	predecessorValidationID    = "ttR2sBgWUEq3e6dupCxYcyiGTv49wnYW3wZ6TUs7XnDkTeZjh"
	predecessorValidatorNodeID = "NodeID-AD66psMnQ257UAF7Jz9FmNAyRdibVnkt"
	predecessorGenesisID       = "KENTRA-CUSTOM-VM-GENESIS-001"
	predecessorBodySHA256      = "58bf3daa9c74aac496451179f053152ed50943b8e8c4dbdac8f34d7172ab4f72"

	constitutionHumanSHA256   = "5dea16e2400c652eb063129711a6bf26c8485a2cbad25ce261beda6f7fee5206"
	constitutionMachineSHA256 = "5a0d6ffec23d09b203212b295d86e45e141a33a7f35f525d56172fb848d1a73c"
	constitutionBindingSHA256 = "cc2f4f2a1fdfab3287ee661dfbcc6e9db72c8e2ac918a94c288598f27916e758"
	constitutionOriginRepo    = "https://github.com/ia-mm-ai/CONSTITUTION"
	constitutionOriginCommit  = "4cf5a926d5fece4e3cccfdfcab40e16431dd332b"
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
	Constitution GenesisConstitution `json:"constitution"`
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
	DirectPredecessor DirectPredecessor `json:"direct_predecessor"`
	FormationAncestor FormationAncestor `json:"formation_ancestor"`
	Relation          string            `json:"relation"`
	Effect            string            `json:"effect"`
	AuthorityTransfer string            `json:"authority_transfer"`
	SharedCurrentness string            `json:"shared_currentness"`
}

type DirectPredecessor struct {
	Protocol                        string `json:"protocol"`
	Version                         string `json:"version"`
	VMID                            string `json:"vm_id"`
	RepositoryPackSHA256            string `json:"repository_pack_sha256"`
	RepositoryChecksumSidecarSHA256 string `json:"repository_checksum_sidecar_sha256"`
	LinuxReleaseSHA256              string `json:"linux_release_sha256"`
	QualificationEvidenceSHA256     string `json:"qualification_evidence_sha256"`
	FinalStateCommitment            string `json:"final_state_commitment"`
	FinalStateSHA256                string `json:"final_state_sha256"`
	Disposition                     string `json:"disposition"`
}

type FormationAncestor struct {
	Name                    string `json:"name"`
	BlockchainID            string `json:"blockchain_id"`
	L1ID                    string `json:"l1_id"`
	VMID                    string `json:"vm_id"`
	ValidationID            string `json:"validation_id"`
	ValidatorNodeID         string `json:"validator_node_id"`
	GenesisID               string `json:"genesis_id"`
	BodySHA256              string `json:"body_sha256"`
	Disposition             string `json:"disposition"`
	CredentialCustody       string `json:"credential_custody"`
	ContinuationRequirement string `json:"continuation_requirement"`
}

type GenesisConstitution struct {
	HumanPath         string                    `json:"human_path"`
	HumanSHA256       string                    `json:"human_sha256"`
	MachinePath       string                    `json:"machine_path"`
	MachineSHA256     string                    `json:"machine_sha256"`
	FormBindingPath   string                    `json:"form_binding_path"`
	FormBindingSHA256 string                    `json:"form_binding_sha256"`
	PublicOrigin      GenesisConstitutionOrigin `json:"public_origin"`
	HumanRole         string                    `json:"human_role"`
	MachineRole       string                    `json:"machine_role"`
	ConflictResult    string                    `json:"conflict_result"`
	ParsingHasForce   bool                      `json:"parsing_has_force"`
}

type GenesisConstitutionOrigin struct {
	Repository                string `json:"repository"`
	Commit                    string `json:"commit"`
	HumanRepositoryPath       string `json:"human_repository_path"`
	MachineRepositoryPath     string `json:"machine_repository_path"`
	FormBindingRepositoryPath string `json:"form_binding_repository_path"`
	HumanBytes                uint64 `json:"human_bytes"`
	MachineBytes              uint64 `json:"machine_bytes"`
	FormBindingBytes          uint64 `json:"form_binding_bytes"`
	Binding                   string `json:"binding"`
	MirrorEffect              string `json:"mirror_effect"`
	PublicationEffect         string `json:"publication_effect"`
}

type DeploymentDescriptor struct {
	Schema          string `json:"schema"`
	GenesisID       string `json:"genesis_id"`
	LocalityID      string `json:"locality_id"`
	Purpose         string `json:"purpose"`
	SourceReference string `json:"source_reference"`
	SourceSHA256    string `json:"source_sha256"`
}

type GenesisLocality struct {
	ID              string           `json:"id"`
	SourceReference string           `json:"source_reference"`
	SourceSHA256    string           `json:"source_sha256"`
	Authority       GenesisAuthority `json:"formation_authority"`
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
		return errors.New("domain must identify the LOCALITY_VM Avalanche RPCChainVM")
	}
	if g.Domain.ExecutionModel != "NON_EVM_SIGNED_EVENT_SOURCED" || g.Domain.StateModel != "VERSIONED_EVENT_SOURCED_LIVING_CAPACITY" {
		return errors.New("unsupported LOCALITY execution or state model")
	}
	if err := requireSafeID("domain.genesis_id", g.Domain.GenesisID); err != nil {
		return err
	}
	if err := requireBoundedText("domain.purpose", g.Domain.Purpose, 1024); err != nil {
		return err
	}
	wantDirect := DirectPredecessor{
		directPredecessorProtocol, directPredecessorVersion, directPredecessorVMID,
		directPredecessorRepositorySHA256, directPredecessorSidecarSHA256,
		directPredecessorReleaseSHA256, directPredecessorEvidenceSHA256,
		directPredecessorStateCommitment, directPredecessorStateSHA256,
		"PRESERVED_IMMUTABLE_PREDECESSOR",
	}
	if g.Lineage.DirectPredecessor != wantDirect {
		return errors.New("lineage must bind the exact immutable LOCALITY_VM_002 v2.1.0 predecessor")
	}
	wantAncestor := FormationAncestor{"KENTRA", predecessorChainID, predecessorL1ID, predecessorVMID, predecessorValidationID, predecessorValidatorNodeID, predecessorGenesisID, predecessorBodySHA256, "FORMATION_COMPLETE_REFERENCE_ONLY", "UNAVAILABLE_TO_CURRENT_OPERATOR", "NEW_LOCALITY_REQUIRES_NEW_VALIDATOR_AND_AUTHORITY_CUSTODY"}
	if g.Lineage.FormationAncestor != wantAncestor {
		return errors.New("lineage must preserve the exact KENTRA formation ancestor")
	}
	if g.Lineage.Relation != "SUCCESSOR_DERIVED_WITHOUT_REWRITE" || g.Lineage.Effect != "ANCESTRY_REFERENCE_ONLY" || g.Lineage.AuthorityTransfer != "NONE" || g.Lineage.SharedCurrentness != "NONE" {
		return errors.New("lineage must not transfer authority or currentness")
	}
	wantOrigin := GenesisConstitutionOrigin{
		constitutionOriginRepo, constitutionOriginCommit,
		"SOURCE/CONSTITUTION_0()1.md", "SOURCE/CONSTITUTION_0()1.json", "SOURCE/FORM_BINDING.json",
		41268, 139114, 1901,
		"EXACT_COMMIT_PATH_LENGTH_SHA256_AND_REPOSITORY_FORM_BINDING",
		"BYTE_IDENTICAL_EVIDENCE_NOT_SEPARATE_ORIGIN",
		"FORM_BINDING_NON_EFFECTS_APPLY",
	}
	if g.Constitution.HumanPath != "source/CONSTITUTION_0()1.md" || g.Constitution.HumanSHA256 != constitutionHumanSHA256 || g.Constitution.MachinePath != "source/CONSTITUTION_0()1.json" || g.Constitution.MachineSHA256 != constitutionMachineSHA256 || g.Constitution.FormBindingPath != "source/FORM_BINDING.json" || g.Constitution.FormBindingSHA256 != constitutionBindingSHA256 || g.Constitution.PublicOrigin != wantOrigin || g.Constitution.HumanRole != "SUBSTANTIVE_MEANING" || g.Constitution.MachineRole != "REPRESENTATION_ONLY" || g.Constitution.ConflictResult != "UNRESOLVED_HOLD_EXACT_SCOPE" || g.Constitution.ParsingHasForce {
		return errors.New("constitution boundary differs from the committed source surface")
	}
	if err := requireSafeID("locality.id", g.Locality.ID); err != nil {
		return err
	}
	if err := requireSafeID("locality.source_reference", g.Locality.SourceReference); err != nil {
		return err
	}
	if err := requireDigest("locality.source_sha256", g.Locality.SourceSHA256); err != nil {
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
	if g.Evolution.StateSchemaVersion != 4 || g.Evolution.UnknownFields != "REJECT" || g.Evolution.GenesisMutability != "IMMUTABLE" || g.Evolution.UpgradePolicy != "ATTESTED_SUCCESSOR_FREEZE_AND_EXTERNAL_BINARY_HANDOFF" || g.Evolution.ActivationEffect != "FREEZES_PREDECESSOR_STATE_WITHOUT_EXECUTING_UNKNOWN_SUCCESSOR_CODE" {
		return errors.New("unsupported evolution policy")
	}
	return nil
}

func (d *DeploymentDescriptor) Validate() error {
	if d.Schema != deploymentDescriptorSchema {
		return fmt.Errorf("expected deployment descriptor schema %s", deploymentDescriptorSchema)
	}
	if err := requireSafeID("deployment.genesis_id", d.GenesisID); err != nil {
		return err
	}
	if err := requireSafeID("deployment.locality_id", d.LocalityID); err != nil {
		return err
	}
	if d.GenesisID == predecessorGenesisID || d.LocalityID == "KENTRA" {
		return errors.New("deployment identity must be distinct from the formation ancestor")
	}
	if err := requireBoundedText("deployment.purpose", d.Purpose, 1024); err != nil {
		return err
	}
	if err := requireSafeID("deployment.source_reference", d.SourceReference); err != nil {
		return err
	}
	return requireDigest("deployment.source_sha256", d.SourceSHA256)
}

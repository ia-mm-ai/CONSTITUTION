package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

type testAuthorities struct {
	host        ed25519.PrivateKey
	participant ed25519.PrivateKey
	continuity1 ed25519.PrivateKey
	continuity2 ed25519.PrivateKey
}

func deterministicKey(fill byte) ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed([]byte(strings.Repeat(string([]byte{fill}), ed25519.SeedSize)))
}

func digestText(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func testGenesis(t *testing.T) ([]byte, *Genesis, testAuthorities) {
	t.Helper()
	keys := testAuthorities{host: deterministicKey(1), participant: deterministicKey(2), continuity1: deterministicKey(3), continuity2: deterministicKey(4)}
	genesis := &Genesis{
		Format:   genesisFormat,
		Version:  genesisVersion,
		Encoding: genesisEncoding,
		Domain: GenesisDomain{
			ID:             domainID,
			GenesisID:      "LOCALITY-RUNTIME-TEST-001",
			VMClass:        "AVALANCHE_CUSTOM_RPCCHAINVM",
			ExecutionModel: "NON_EVM_SIGNED_EVENT_SOURCED",
			StateModel:     "VERSIONED_EVENT_SOURCED_LIVING_CAPACITY",
			Purpose:        "Exercise the LOCALITY executable grammar.",
		},
		Lineage: GenesisLineage{
			DirectPredecessor: DirectPredecessor{
				Protocol: directPredecessorProtocol, Version: directPredecessorVersion,
				VMID: directPredecessorVMID, RepositoryPackSHA256: directPredecessorRepositorySHA256,
				RepositoryChecksumSidecarSHA256: directPredecessorSidecarSHA256,
				LinuxReleaseSHA256:              directPredecessorReleaseSHA256,
				QualificationEvidenceSHA256:     directPredecessorEvidenceSHA256,
				FinalStateCommitment:            directPredecessorStateCommitment,
				FinalStateSHA256:                directPredecessorStateSHA256,
				Disposition:                     "PRESERVED_IMMUTABLE_PREDECESSOR",
			},
			FormationAncestor: FormationAncestor{
				Name: "KENTRA", BlockchainID: predecessorChainID, L1ID: predecessorL1ID, VMID: predecessorVMID,
				ValidationID: predecessorValidationID, ValidatorNodeID: predecessorValidatorNodeID,
				GenesisID: predecessorGenesisID, BodySHA256: predecessorBodySHA256,
				Disposition: "FORMATION_COMPLETE_REFERENCE_ONLY", CredentialCustody: "UNAVAILABLE_TO_CURRENT_OPERATOR",
				ContinuationRequirement: "NEW_LOCALITY_REQUIRES_NEW_VALIDATOR_AND_AUTHORITY_CUSTODY",
			},
			Relation: "SUCCESSOR_DERIVED_WITHOUT_REWRITE", Effect: "ANCESTRY_REFERENCE_ONLY",
			AuthorityTransfer: "NONE", SharedCurrentness: "NONE",
		},
		Constitution: GenesisConstitution{
			HumanPath: "source/CONSTITUTION_0()1.md", HumanSHA256: constitutionHumanSHA256,
			MachinePath: "source/CONSTITUTION_0()1.json", MachineSHA256: constitutionMachineSHA256,
			FormBindingPath: "source/FORM_BINDING.json", FormBindingSHA256: constitutionBindingSHA256,
			PublicOrigin: GenesisConstitutionOrigin{
				Repository: constitutionOriginRepo, Commit: constitutionOriginCommit,
				HumanRepositoryPath: "SOURCE/CONSTITUTION_0()1.md", MachineRepositoryPath: "SOURCE/CONSTITUTION_0()1.json",
				FormBindingRepositoryPath: "SOURCE/FORM_BINDING.json",
				HumanBytes:                41268, MachineBytes: 139114, FormBindingBytes: 1901,
				Binding:           "EXACT_COMMIT_PATH_LENGTH_SHA256_AND_REPOSITORY_FORM_BINDING",
				MirrorEffect:      "BYTE_IDENTICAL_EVIDENCE_NOT_SEPARATE_ORIGIN",
				PublicationEffect: "FORM_BINDING_NON_EFFECTS_APPLY",
			},
			HumanRole: "SUBSTANTIVE_MEANING", MachineRole: "REPRESENTATION_ONLY",
			ConflictResult: "UNRESOLVED_HOLD_EXACT_SCOPE", ParsingHasForce: false,
		},
		Locality: GenesisLocality{
			ID:              "LOCALITY-RUNTIME-TEST",
			SourceReference: "CONTINUITY",
			SourceSHA256:    digestText("test continuity source"),
			Authority: GenesisAuthority{
				Scheme:    "ED25519",
				KeyID:     "locality-host-test-001",
				PublicKey: hex.EncodeToString(keys.host.Public().(ed25519.PublicKey)),
				Scope:     "FORMATION_AND_BOOTSTRAP_ONLY",
			},
		},
		Profile: GenesisProfile{
			FormID:         formID,
			FormSHA256:     formSHA256,
			GatePostures:   append([]string(nil), requiredGatePostures...),
			EffectCeiling:  append([]string(nil), requiredEffectCeiling...),
			NonCollapse:    append([]string(nil), requiredNonCollapse...),
			PrivateState:   "NEVER_REQUIRED_ON_CHAIN_ONLY_SIGNED_COMMITMENTS_AND_SUCCESSION_DIGESTS",
			AdmissionRule:  "FRESH_PRESENTATION_AND_ADMISSION_REQUIRED_FOR_EVERY_INGRESS",
			EntryRule:      "FIRST_ENTER_OR_CHECKPOINT_BOUND_REENTER_ATOMICALLY_RESERVES_FULL_LIFECYCLE",
			CapacityRule:   "AVAILABLE_EQUALS_ACTUAL_MINUS_UNRESOLVED_MINUS_CORRECTION_EGRESS",
			CorrectionRule: "APPEND_ONLY_PRESERVE_PREDECESSOR",
			PresenceRule:   "CURRENT_BINDING_ONLY_NOT_HISTORICAL_RECEIPT",
			DormancyRule:   "P0_CHECKPOINTS_WITHOUT_FABRICATED_ACTIVITY",
			ReentryRule:    "SAME_ACTOR_PLUS_SEALED_SINGLE_USE_DEPARTURE_CHECKPOINT_PLUS_FRESH_ADMISSION_PLUS_DETERMINISTIC_STATE_SUCCESSION",
			SuccessionRule: "EXPLICIT_ATTESTED_BOUNDARY_WITHOUT_AUTHORITY_INHERITANCE",
			OriginRule:     "PUBLIC_ORIGIN_EXACTLY_BOUND_WHILE_LOCAL_FORMATION_REMAINS_DISTINCT",
		},
		Capacity: GenesisCapacity{
			Unit: "PRESENCE_CAPACITY_UNIT", StructuralCeilingUnits: 100,
			InitialActualUnits: 100, CorrectionEgressReserveUnits: 10,
			MaxAdmissionOfferSeconds: 3600,
			AvailabilityRule:         "C_AVAILABLE_EQUALS_C_ACTUAL_MINUS_C_UNRESOLVED_MINUS_C_CORRECTION_EGRESS",
			ReservationRule:          "ENTER_RESERVES_WORK_AND_LAWFUL_RESOLUTION_ATOMICALLY",
			DegradationRule:          "INSUFFICIENT_ACTUAL_CAPACITY_FORCES_HOLD_WITH_CORRECTION_AND_EGRESS_OPEN",
		},
		Continuity: GenesisContinuity{
			InitialPosture: "DORMANT_P0", MinimumContinuityAuthorities: 2, SuccessorAttestationThreshold: 2,
			FormationAuthorityLifecycle:  "EXPLICITLY_EXHAUSTIBLE_AFTER_CONTINUITY_SEEDING",
			PZeroRule:                    "CHECKPOINT_DORMANT_NO_FABRICATED_ACTIVITY_REENTRY_OPEN",
			PulseRule:                    "SIGNED_CURRENTNESS_OBSERVATION_NOT_AUTONOMOUS_AUTHORITY",
			InfrastructureSuccessionRule: "INFRASTRUCTURE_CANNOT_SELF_APPOINT_SUCCESSOR",
		},
		InitialState: GenesisInitialState{Revision: 0, PreviousStateCommitment: nil, History: "APPEND_ONLY"},
		ClaimLimits: GenesisClaimLimits{
			ExternalTruth:   "NOT_INFERRED",
			Receipt:         "HISTORICAL_EVIDENCE_NOT_CURRENT_PRESENCE",
			Acceptance:      "NOT_INFERRED",
			Consent:         "NOT_INFERRED",
			Relation:        "NOT_CREATED_BY_GENESIS",
			Crossing:        "NOT_CREATED_BY_GENESIS",
			Capacity:        "SIGNED_LOCAL_ACCOUNT_NOT_EXTERNAL_RESOURCE_PROOF",
			Pulse:           "SIGNED_LOCAL_CURRENTNESS_NOT_EXTERNAL_LIVENESS_PROOF",
			StateSuccession: "DETERMINISTIC_COMMITMENT_LINEAGE_NOT_SEMANTIC_TRUTH_OR_CONSCIOUSNESS_PROOF",
			Unknown:         "REMAINS_UNKNOWN",
		},
		Evolution: GenesisEvolution{
			StateSchemaVersion: 4,
			UnknownFields:      "REJECT",
			GenesisMutability:  "IMMUTABLE",
			UpgradePolicy:      "ATTESTED_SUCCESSOR_FREEZE_AND_EXTERNAL_BINARY_HANDOFF",
			ActivationEffect:   "FREEZES_PREDECESSOR_STATE_WITHOUT_EXECUTING_UNKNOWN_SUCCESSOR_CODE",
		},
	}
	genesisBytes, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseGenesis(genesisBytes)
	if err != nil {
		t.Fatalf("test genesis is invalid: %v", err)
	}
	return genesisBytes, parsed, keys
}

func testEffectCeiling() []string {
	return append([]string(nil), requiredEffectCeiling...)
}

func testMediumCapabilities() []string {
	return append([]string(nil), requiredMediumCapabilities...)
}

func makeTransition(t *testing.T, state *RuntimeState, actorID string, key ed25519.PrivateKey, operation, locusID string, payload any) *Transition {
	t.Helper()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	transition, err := signTransition(UnsignedTransition{
		Schema:                  transitionSchema,
		Operation:               operation,
		Revision:                state.Revision + 1,
		PreviousStateCommitment: state.StateCommitment,
		ActorID:                 actorID,
		Nonce:                   state.NextNonces[actorID],
		LocusID:                 locusID,
		ObservedAt:              int64(state.Revision + 1000),
		Effect:                  operationEffects[operation],
		Payload:                 payloadBytes,
	}, key)
	if err != nil {
		t.Fatalf("sign %s: %v", operation, err)
	}
	return transition
}

func transitionHexID(t *testing.T, transition *Transition) string {
	t.Helper()
	id, err := transition.ID()
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(id[:])
}

func applyToClone(state *RuntimeState, genesis *Genesis, transition *Transition) (*RuntimeState, *Receipt, error) {
	clone, err := state.clone()
	if err != nil {
		return nil, nil, err
	}
	id, err := transition.ID()
	if err != nil {
		return nil, nil, err
	}
	receipt, err := clone.apply(genesis, transition, hex.EncodeToString(id[:]))
	return clone, receipt, err
}

func mustApply(t *testing.T, state *RuntimeState, genesis *Genesis, transition *Transition) (*RuntimeState, *Receipt, string) {
	t.Helper()
	next, receipt, err := applyToClone(state, genesis, transition)
	if err != nil {
		t.Fatalf("apply %s: %v", transition.Unsigned.Operation, err)
	}
	return next, receipt, receipt.TransitionID
}

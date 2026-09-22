package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// The constitutional core this implementation is bound to. These values are the
// exact bytes recorded in the implementation CORE_BINDING.json. They are compiled
// into the runtime so that a materialized genesis cannot silently drift from the
// core it claims to carry. The core is referenced by exact bytes and never
// mirrored into the runtime.
const (
	coreRepository  = "https://github.com/ia-mm-ai/PRESENCE"
	coreCommit      = "a0cbb1080540bbe83f8678e81240247585dba060"
	coreBindingPath = "SYSTEM/AVALANCHE/AVALANCHE_IMPLEMENTATION_001/CORE_BINDING.json"
	coreDigest      = "20f263e71cef4c0a0735303701c91539ee7847730bd3f1dc0eb54153bd36c00c"

	coreHumanRole       = "SUBSTANTIVE_MEANING"
	coreMachineRole     = "REPRESENTATION_ONLY"
	coreConflictResult  = "UNRESOLVED_HOLD_AT_EXACT_SCOPE"
	coreMirrorEffect    = "NO_EMBEDDED_CORE_MIRROR_ONLY_EXACT_BYTE_REFERENCE"
	corePublishEffect   = "BYTE_IDENTICAL_REFERENCE_NOT_SEPARATE_ORIGIN"
	implementationID    = "AVALANCHE_IMPLEMENTATION_001"
	implementationLevel = "1.0.0"
)

// coreFiles lists the exact constitutional core bytes in ascending path order.
var coreFiles = []CoreFileBinding{
	{"SOURCE_MACHINE", "SOURCE/CONSTITUTION_0()1.json", 146048, "519a81d2a26d5bf32afd77566bbb35b4d28b22af85ac45a7055fccd1b45222f8"},
	{"SOURCE_HUMAN", "SOURCE/CONSTITUTION_0()1.md", 44133, "affeb5738cdfeea7ee4fe985652bf78b15eb6dfbba5871d82f0f2213a81832d0"},
	{"CSC_MACHINE", "STATE/CONTINUITY-STATE-CAPABILITY/CONTINUITY-STATE-CAPABILITY.json", 23930, "8fcc25f3144b362a356735ddbd7e3f253e72789ef7b43671b27dea3c595cb308"},
	{"CSC_HUMAN", "STATE/CONTINUITY-STATE-CAPABILITY/CONTINUITY-STATE-CAPABILITY.md", 16126, "f826d86472a97a6db891437d1324750dcb103e786dc021c8ace095f4e8a6ca7f"},
	{"DCR_MACHINE", "STATE/DYNAMIC-CAPACITY-REGULATION/DYNAMIC-CAPACITY-REGULATION.json", 24561, "cac11fcf704a3afac523da6eff340acdb8286eebe8a199646ebb5bffb5eed897"},
	{"DCR_HUMAN", "STATE/DYNAMIC-CAPACITY-REGULATION/DYNAMIC-CAPACITY-REGULATION.md", 13923, "36e2aa067b7e9fa0e72ff54b01cec7d1e200722f9dfddcd681ef7f5d4edd8abe"},
	{"LINEAGE_MACHINE", "STATE/LINEAGE/LINEAGE.json", 38360, "bf61fb31bf4703b238029550f76e078cefb393a0227bf5e766fc03bab64042fa"},
	{"LINEAGE_HUMAN", "STATE/LINEAGE/LINEAGE.md", 21178, "560d1689e5858b192dde23bb1a0fa7300d90fb0f0814cf86561dcba47a485485"},
	{"STATE_MACHINE", "STATE/STATE.json", 4605, "324752d635a858e0d0ccdc238c7852b93422d947fddaa1fde327a611e6702eb0"},
	{"STATE_HUMAN", "STATE/STATE.md", 10418, "4fb429cef3cc965f654ee9bf014e6a0b3573a2db2e40806c70c9483f399867ae"},
}

// expectedPredecessors records the two immutable predecessor carriers this
// implementation was derived from. Nothing operational is inherited from them.
var expectedPredecessors = []LineagePredecessor{
	{
		ID:            "LOCALITY_VM_003",
		Version:       "3.0.0",
		CarrierPath:   "STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip",
		CarrierBytes:  6201060,
		CarrierSHA256: "073949b8b74e7407a433b74efab71f22e2ee93c4d9ed80eb0aeb07b164c08d2b",
		Disposition:   "HISTORICAL_REFERENCE_ONLY",
	},
	{
		ID:            "LOCALITY_MEDIUM_001",
		Version:       "2.0.0",
		CarrierPath:   "STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip",
		CarrierBytes:  260923,
		CarrierSHA256: "61353cddc93d62ac28d60e0c7357203ecfe01235e7d249e973ffb16356990989",
		Disposition:   "HISTORICAL_REFERENCE_ONLY",
	},
}

// reservedHistoricalIdentities are predecessor and formation-ancestor names that
// must never be reused as a new operative identity. They are refused here
// precisely because they belong to closed historical records.
var reservedHistoricalIdentities = []string{
	"KENTRA",
	"KENTRA-CUSTOM-VM-GENESIS-001",
	"LOCALITY_VM_001",
	"LOCALITY_VM_002",
	"LOCALITY_VM_003",
	"LOCALITY_MEDIUM_001",
}

func computeCoreDigest(files []CoreFileBinding) string {
	builder := new(strings.Builder)
	for _, file := range files {
		fmt.Fprintf(builder, "%s  %d  %s\n", file.SHA256, file.ByteLength, file.Path)
	}
	digest := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(digest[:])
}

func requireDistinctFromPredecessors(field, value string) error {
	for _, reserved := range reservedHistoricalIdentities {
		if strings.EqualFold(value, reserved) {
			return fmt.Errorf("%s must not reuse the closed historical identity %s", field, reserved)
		}
	}
	return nil
}

// currentCore returns the compiled core binding for materialization. Core
// references are injected from here and are never accepted as operator input.
func currentCore() GenesisCore {
	files := make([]CoreFileBinding, len(coreFiles))
	copy(files, coreFiles)
	return GenesisCore{
		Repository:        coreRepository,
		Commit:            coreCommit,
		BindingPath:       coreBindingPath,
		CoreDigest:        coreDigest,
		Files:             files,
		HumanRole:         coreHumanRole,
		MachineRole:       coreMachineRole,
		ConflictResult:    coreConflictResult,
		ParsingHasForce:   false,
		MirrorEffect:      coreMirrorEffect,
		PublicationEffect: corePublishEffect,
	}
}

func currentLineage() GenesisLineage {
	predecessors := make([]LineagePredecessor, len(expectedPredecessors))
	copy(predecessors, expectedPredecessors)
	return GenesisLineage{
		Predecessors:           predecessors,
		Relation:               "DERIVATION_WITHOUT_SUCCESSION",
		Effect:                 "ANCESTRY_REFERENCE_ONLY",
		AuthorityTransfer:      "NONE",
		SharedCurrentness:      "NONE",
		OperationalInheritance: "NONE_NO_PREDECESSOR_CHAIN_L1_VALIDATOR_OR_KEY_CARRIED",
	}
}

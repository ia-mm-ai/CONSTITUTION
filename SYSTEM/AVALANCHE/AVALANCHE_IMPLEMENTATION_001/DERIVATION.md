# Derivation record

## Successor identity

- Implementation: `AVALANCHE_IMPLEMENTATION_001`
- Version: `1.0.0`
- VM protocol identity: `PRESENCE_AVALANCHE_VM_001`
- Source origin commit: `a0cbb1080540bbe83f8678e81240247585dba060`

The implementation was materialized as readable source from these immutable,
read-only repository inputs:

| Input | Bytes | SHA-256 |
| --- | ---: | --- |
| `STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip` | 6201060 | `073949b8b74e7407a433b74efab71f22e2ee93c4d9ed80eb0aeb07b164c08d2b` |
| `STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip` | 260923 | `61353cddc93d62ac28d60e0c7357203ecfe01235e7d249e973ffb16356990989` |

The VM predecessor supplies the current VM, genesis, compatibility overlay,
build/preflight, and rehearsal basis. The medium predecessor supplies the FIELD
runtime, profiles, and tests. Historical predecessor and ancestor identifiers
remain only where lineage or regression evidence requires them.

Generated archives, predecessor release manifests, predecessor workflows,
private authority material, and predecessor Source mirrors were not copied into
the successor. `SOURCE/`, canonical `STATE/`, both ZIP inputs, and predecessor
evidence were not rewritten.

## Structural repair

The predecessor FIELD required nearly every operation other than `BOUND` to use
the accepted active locus, while the VM required body-local
`DECLARE_CAPACITY` to use an empty `locus_id`. The successor replaces the
divergent tables and any one-off exception with the exhaustive normative
`protocol/operations.json`.

Go embeds and validates that file in `operation_contract.go`. FIELD reads the
same file in `field/src/constants.js`. Both runtimes fail closed on unknown,
duplicate, missing, or unclassified operations, and regression tests compare
operation, effect, and scope inventories.

## Official VM-ID derivation

The canonical derivation input is exactly:

```text
PRESENCE_AVALANCHE_VM_001
```

No alternate CLI-safe alias is used. `main.go` derives the ID through the
pinned AvalancheGo `v1.15.0` Go API:

```text
ids.ID(hashing.ComputeHash256Array([]byte("PRESENCE_AVALANCHE_VM_001"))).String()
```

Result:

```text
cQYXygFUVutQucm4s8pr8M51sRdS1UfrTbpMdEgpRYt2JvEzr
```

This derives an executable protocol identifier only. It does not create a
blockchain, L1, Authority, or deployment.


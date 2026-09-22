# AVALANCHE_IMPLEMENTATION_001

First self-contained source implementation of the PRESENCE Avalanche runtime.

- Implementation identity: `AVALANCHE_IMPLEMENTATION_001`
- Version: `1.0.0`
- VM protocol identity: `PRESENCE_AVALANCHE_VM_001`
- VM ID (official AvalancheGo `ids.ToID` derivation over the canonical
  identity): `cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8`
- FIELD protocol identity: `PRESENCE_AVALANCHE_FIELD_001`

Statuses: `SOURCE_IMPLEMENTED`, `FUNCTIONALLY_TESTED`,
`QUALIFICATION_INCOMPLETE`, `NOT_RELEASED`, `NOT_DEPLOYED`,
`NOT_READY_FOR_MAINNET` — see `records/QUALIFICATION_STATUS.json` for the
exact passed/unresolved boundaries.

This is an **implementation-source crossing**. It is not release
qualification, deployment, L1 creation, adoption, formation, succession,
Authority or Presence.

## Layout

| Path | Content |
| --- | --- |
| `*.go`, `go.mod`, `go.sum` | VM source (pinned AvalancheGo v1.15.0, RPCChainVM protocol 46) |
| `contract/OPERATION_SCOPE_CONTRACT_001.json` | The single shared normative operation-scope contract |
| `field/` | FIELD runtime source (`package.json`, zero third-party dependencies) |
| `genesis/` | Genesis schema/template; materializer lives in the VM CLI |
| `binding/` | Current Source/STATE bindings (`FORM_BINDING.json`, `SOURCE_STATE_BINDING.json`) |
| `admin/ADMIN_INPUT.schema.json` | Administrative deployment-descriptor schema |
| `docs/DEPLOYMENT_ARCHITECTURE.md` | Custom non-EVM deployment architecture (not performed) |
| `records/` | Derivation record, qualification status, coupled-lifecycle transcript |
| `scripts/` | Build scripts and validator preflight |

## The operation-scope contract

FIELD (`src/engine.js`) formerly required every operation except `BOUND` to
name the accepted active locus, while the VM required `DECLARE_CAPACITY` to
use an empty `locus_id`. That predecessor coupling defect is repaired
structurally: `contract/OPERATION_SCOPE_CONTRACT_001.json` classifies every
supported operation into `NEW_LOCUS`, `ACTIVE_LOCUS`, `BODY_LOCAL`,
`DORMANT_BODY` or `TARGET_SCOPED`. The VM embeds the file at compile time
(`contract.go`); the FIELD loads the identical file at runtime
(`field/src/scope-contract.js`). Neither runtime declares its own scope
table, unknown or unclassified operations fail closed on both sides, and the
mirrored regression suites (`scope_contract_test.go`,
`field/test/scope-contract.test.js`) prove the inventories agree.

## Build and test

```sh
GOTOOLCHAIN=auto go build -o presence-avalanche-vm .
GOTOOLCHAIN=auto go vet ./...
GOTOOLCHAIN=auto go test ./...
cd field && node --test "test/*.test.js"
node scripts/verify.mjs && node scripts/qualify.mjs   # from field/
```

## Coupled lifecycle rehearsal (loopback only)

```sh
cd field
node scripts/coupled-lifecycle.mjs /path/to/presence-avalanche-vm /tmp/coupled-run
```

Drives the real FIELD engine against the real VM (`--serve-rehearsal`,
in-memory, loopback-only) through
`read → BOUND → receipt → renewed read → DECLARE_CAPACITY (empty locus) →
receipt → renewed read → PULSE → CORRECT → final read` — beyond the former
predecessor failure point. Transcript evidence:
`records/COUPLED_LIFECYCLE_TRANSCRIPT.json`. Disposable keys stay in the
temporary work directory and are never persisted here.

## Derivation

Derived read-only from the byte-identical predecessor archives
`STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip` and
`STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip`; see
`records/DERIVATION_RECORD.json`. Historical identifiers
(`LOCALITY_VM_003`, `LOCALITY_MEDIUM_001`, `VM001`, `VM002`, `KENTRA`)
remain only as derivation/lineage evidence, never as the operative successor
identity.

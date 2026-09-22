# AVALANCHE_IMPLEMENTATION_001

This directory is the single coherent, unreleased source implementation of
`AVALANCHE_IMPLEMENTATION_001` version `1.0.0`. Its custom RPCChainVM protocol
identity is `PRESENCE_AVALANCHE_VM_001`, and its canonical VM ID — derived by
right-padding the UTF-8 name bytes with zero bytes to `ids.IDLen` and
converting through the pinned AvalancheGo v1.15.0 `ids.ToID` — is:

```
cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8
```

This crossing is source implementation and functional testing only. It does
not create a blockchain, L1, Validator Manager, Authority, Presence, release,
deployment, adoption, formation, succession, or public-network effect.

This form was consolidated from three completed crossings (accepted main at
`5e09d1f`, PR #16 at `0ab3800`, and PR #17 at `d5e6703`); their lineage and
disposition are recorded in `records/COHERENCE_RECORD_001.json` and
`records/DERIVATION_RECORD.json`. The first crossing's hashed VM ID remains
addressable only there and in Git history; it has no operative use.

## Layout

- `vm/`: the sole Go RPCChainVM module and metabolism, with the bounded
  `vm/rehearsal/` runner submodule.
- `vm/protocol/`: the exact shared contract bytes embedded by Go and read by
  FIELD — `operations.json` (the single exhaustive operation/effect/scope
  contract with fail-closed scope classes) and `binding.json` (the exact
  Source/whole-STATE/predecessor-archive byte binding).
- `vm/genesis/`: genesis schema, template and administrative-input schema.
- `field/`: the sole dependency-free Node FIELD runtime — the locality-facing
  admission/capability membrane — bound to the single canonical VM identity.
- `bindings/`: Source, STATE, and predecessor archive binding record.
- `integration/`: coupled FIELD/real-VM lifecycle and disposable private
  network harnesses.
- `qualification/`: the single current qualification record.
- `records/`: derivation and coherence records.
- `compat/`: compatibility declaration and the required
  `PRESENCE_AVALANCHE_SECURITY_OVERLAY_001` AvalancheGo overlay.
- `docs/`: operational deployment boundary.
- `scripts/`: build, preflight, binding verification, and test entry points.

## Identity

| Item | Value |
| --- | --- |
| VM protocol | `PRESENCE_AVALANCHE_VM_001` |
| VM ID | `cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8` |
| State schema | `PRESENCE_AVALANCHE_RUNTIME_STATE_001` |
| Transition schema | `PRESENCE_AVALANCHE_TRANSITION_001` |
| Receipt schema | `PRESENCE_AVALANCHE_RECEIPT_001` |
| Actor namespace | `PRESENCE-AVALANCHE-ACTOR-` |
| FIELD protocol | `PRESENCE_AVALANCHE_FIELD_001` |

FIELD verifies the accepted state schema and the exact embedded
operation-contract digest before every draft and submission; there is exactly
one operation contract, one VM ID, one state-schema family, and one actor
namespace.

## Verify

From this directory:

```sh
./scripts/verify-bindings.py
python3 scripts/verify-binding.py
./scripts/test.sh
./scripts/build.sh /tmp/presence-avalanche-build
/tmp/presence-avalanche-build/presence-avalanche-vm --version-json
./integration/run.sh
```

`scripts/test.sh` validates the shared contract, the VM and rehearsal Go
modules, and the FIELD runtime. It does not contact or write to a network. The
disposable private network rehearsal (`integration/network.sh`) requires an
operator-supplied compatible AvalancheGo binary and is deliberately separate.

## Materialize deployment-specific genesis

Keep private authority material outside the repository:

```sh
/tmp/presence-avalanche-build/presence-avalanche-vm --generate-authority /tmp/presence-authority
/tmp/presence-avalanche-build/presence-avalanche-vm --materialize-genesis \
  vm/genesis/PRESENCE_AVALANCHE_RUNTIME_GENESIS_TEMPLATE_001.json \
  /path/to/deployment-descriptor.json \
  /tmp/presence-authority/locality-authority.public.json \
  /tmp/presence-avalanche-genesis.json
```

The deployment descriptor must satisfy
`vm/genesis/administrative-input.schema.json`. The output is checked by
`--check-genesis`; it is deployment-specific material, not repository source.

## Status

See `qualification/QUALIFICATION_STATUS.json` for the current, regenerated
qualification record and its unresolved boundaries. This implementation is
`QUALIFICATION_INCOMPLETE`, `NOT_RELEASED`, `NOT_DEPLOYED`, and
`NOT_READY_FOR_MAINNET`.

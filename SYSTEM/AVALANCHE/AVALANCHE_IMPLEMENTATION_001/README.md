# AVALANCHE_IMPLEMENTATION_001

This directory is the first ordinary-file source implementation of
`AVALANCHE_IMPLEMENTATION_001` version `1.0.0`. Its custom RPCChainVM protocol
identity is `PRESENCE_AVALANCHE_VM_001`.

This crossing is source implementation and functional testing only. It does
not create a blockchain, L1, Validator Manager, Authority, Presence, release,
deployment, adoption, formation, succession, or public-network effect.

## Layout

- `protocol/operations.json`: the single normative operation/effect/scope contract.
- root Go files: custom Avalanche VM, state machine, genesis materializer, and tests.
- `field/`: dependency-free Node FIELD runtime and tests.
- `genesis/` and `schemas/`: genesis template and administrative interfaces.
- `bindings/`: exact Source, STATE, and predecessor archive binding.
- `compat/` and `dependencies/`: pinned runtime and dependency declarations.
- `rehearsal/`: disposable private three-validator lifecycle harness.
- `scripts/`: build, preflight, binding verification, and test entry points.
- `DERIVATION.md`, `DEPLOYMENT_ARCHITECTURE.md`, and
  `QUALIFICATION_STATUS.md`: lineage, operational boundary, and observed status.

## Verify

From this directory:

```sh
./scripts/verify-bindings.py
./scripts/test.sh
./scripts/build.sh /tmp/presence-avalanche-build
/tmp/presence-avalanche-build/presence-avalanche-vm --version-json
```

`scripts/test.sh` does not contact or write to a network. The network rehearsal
requires an operator-supplied compatible AvalancheGo binary and is deliberately
separate.

## Materialize deployment-specific genesis

Keep private authority material outside the repository:

```sh
/tmp/presence-avalanche-build/presence-avalanche-vm --generate-authority /tmp/presence-authority
/tmp/presence-avalanche-build/presence-avalanche-vm --materialize-genesis \
  genesis/PRESENCE_AVALANCHE_GENESIS_TEMPLATE_001.json \
  /path/to/PRESENCE_AVALANCHE_ADMIN_INPUT_001.json \
  /tmp/presence-authority/presence-avalanche-authority.public.json \
  /tmp/presence-avalanche-genesis.json
```

The administrative input must satisfy
`schemas/PRESENCE_AVALANCHE_ADMIN_INPUT_001.schema.json`. The output is checked
by `--check-genesis`; it is deployment-specific material, not repository source.

# AVALANCHE_IMPLEMENTATION_001

`AVALANCHE_IMPLEMENTATION_001` version `1.0.0` is one unreleased
implementation-source crossing:

- VM protocol: `PRESENCE_AVALANCHE_VM_001`
- FIELD protocol: `PRESENCE_AVALANCHE_FIELD_001`
- VM ID: `cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8`
- runtime state: `PRESENCE_AVALANCHE_RUNTIME_STATE_001`
- transition: `PRESENCE_AVALANCHE_TRANSITION_001`
- receipt: `PRESENCE_AVALANCHE_RECEIPT_001`
- actor namespace: `PRESENCE-AVALANCHE-ACTOR-`

The VM ID is `ids.ToID` applied to the UTF-8 bytes of
`PRESENCE_AVALANCHE_VM_001`, right-padded with zero bytes to `ids.IDLen`.
There is no alternative runtime identity, migration, or predecessor-state
continuity in this source tree.

## Operative topology

- `vm/`: the sole Go RPCChainVM module, metabolism, genesis, and bounded
  rehearsal submodule.
- `vm/protocol/operations.json`: the one exhaustive operation/effect/scope
  contract, embedded verbatim by Go and read directly by FIELD.
- `field/`: the sole locality-facing admission and capability membrane.
- `bindings/`: the exact Source, canonical STATE, and predecessor-archive
  binding.
- `integration/`: coupled FIELD/real-VM and optional private-network checks.
- `qualification/`: current observations and explicit ceilings.
- `records/`: derivation, coherence, and bounded historical observations.
- `compat/`: pinned AvalancheGo and security-overlay compatibility.
- `docs/`: operational boundaries.
- `scripts/`: the one default build and verification path.

`vm/rehearsal/` is a bounded disposable-network support module, not a second VM.

## Verify and build

```sh
./scripts/test.sh
./scripts/build.sh /tmp/presence-avalanche-build
/tmp/presence-avalanche-build/presence-avalanche-vm --version-json
```

The default build always compiles `vm/`. The default FIELD configuration in
`field/examples/config.template.json` always targets the same cNhh VM ID.

## Materialize disposable genesis

Keep all private material outside the repository:

```sh
/tmp/presence-avalanche-build/presence-avalanche-vm --generate-authority /tmp/presence-authority
/tmp/presence-avalanche-build/presence-avalanche-vm --materialize-genesis \
  vm/genesis/PRESENCE_AVALANCHE_RUNTIME_GENESIS_TEMPLATE_001.json \
  /path/to/PRESENCE_AVALANCHE_DEPLOYMENT_DESCRIPTOR_001.json \
  /tmp/presence-authority/locality-authority.public.json \
  /tmp/presence-avalanche-genesis.json
/tmp/presence-avalanche-build/presence-avalanche-vm --check-genesis \
  /tmp/presence-avalanche-genesis.json
```

This source does not create a release, deployment, chain, L1, Validator
Manager, Authority, formation, adoption, succession, Presence claim, tag, or
public-network mutation.

# Deployment architecture — AVALANCHE_IMPLEMENTATION_001

Status: `NOT_DEPLOYED`, `NOT_READY_FOR_MAINNET`. This document describes the
required deployment shape; nothing in it has been performed against any public
network.

## VM shape

- Custom **non-EVM** virtual machine: `PRESENCE_AVALANCHE_VM_001` v1.0.0.
- Built against pinned AvalancheGo `v1.15.0`
  (`70bd6d063b7343fd2cd8217200aaf77b57f19f68`), Go `1.25.13`, Linux amd64,
  **RPCChainVM protocol 46**.
- VM ID (official `ids.ToID` derivation over the canonical identity):
  `cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8`.
- The same binary is both the plugin and the operator CLI
  (`--version-json`, `--check-genesis`, `--materialize-genesis`,
  `--generate-authority`, `--sign-unsigned`, `--serve-rehearsal`).

## Supported-in-principle architecture

Arbitrary-VM blockchain creation is available on Avalanche; the following
constraints govern this custom plugin:

1. **Every validator must receive the custom executable.** The VM binary must
   be installed in each validator's AvalancheGo `plugins/` directory under the
   exact VM ID filename (`cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8`).
   There is no on-chain code distribution for RPCChainVM plugins.
2. **External Validator Manager.** An L1 validator manager contract may reside
   on the C-Chain or another Warp-capable EVM chain. The manager governs the
   validator set; the custom chain itself carries no EVM and cannot host it.
3. **Builder Console limitation.** Builder Console's stock Docker generation
   assumes Subnet-EVM. It cannot provision this plugin. Deployment therefore
   requires a **manual or CLI-aware path**: distribute the binary, place it in
   `plugins/`, configure `track-subnets`, create the chain with this VM ID and
   a genesis materialized by `--materialize-genesis`.
4. **Genesis materialization.** Administrators supply a deployment descriptor
   (see `admin/ADMIN_INPUT.schema.json`) and a locality authority public file.
   `--materialize-genesis TEMPLATE DESCRIPTOR PUBLIC_AUTHORITY OUTPUT`
   produces the canonical genesis; `--check-genesis` re-verifies it.
5. **Key custody.** Locality authority private keys are generated with
   `--generate-authority` into a `0700` directory and must never be committed
   or persisted beyond their administrative purpose. Actor signing is external
   (ed25519); the FIELD never holds locality keys.

## Rehearsal-only execution

`--serve-rehearsal GENESIS 127.0.0.1:PORT` serves the real VM HTTP surface on
an explicit loopback address with in-memory state, accepting transitions
through the real block path. It refuses non-loopback addresses. It exists for
coupled FIELD⇄VM lifecycle rehearsal (`field/scripts/coupled-lifecycle.mjs`)
and is not a deployment mode.

## Not performed

- No Mainnet, Fuji or public-network write.
- No chain, subnet or L1 creation.
- No release, tag or binary distribution.
- No three-validator restart/rejoin qualification (recorded `UNRESOLVED` in
  `records/QUALIFICATION_STATUS.json`).

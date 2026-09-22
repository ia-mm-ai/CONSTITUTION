# Deployment architecture and boundary

## Architecture

`PRESENCE_AVALANCHE_VM_001` is a custom, non-EVM RPCChainVM. Every validator of
a chain using it must receive the same executable, installed under the derived
VM ID, and must pass `scripts/preflight-node.sh` before any chain-creation
proposal. Genesis must be materialized from the committed template using
deployment-specific non-secret administrative input and a newly generated
formation-authority public key.

The stock Builder Console Docker path assumes Subnet-EVM and is not the
provisioning path for this custom plugin. An operator must use a manual or
CLI-aware process that distributes the executable, verifies checksums and
version/protocol output on every node, installs the custom genesis, and confirms
unanimous preflight results before creating a private test chain.

An external Validator Manager, if later required, may reside on C-Chain or
another Warp-capable EVM chain. It is outside this VM, is not supplied here, and
must not be inferred from source implementation.

## Administrative and custody boundary

`vm/genesis/administrative-input.schema.json` defines the non-secret
deployment descriptor. Formation authority generation writes the private file
with mode `0600`, refuses overwrite, and emits a separate public document.
Production custody must be operator-controlled and outside this repository.
Disposable rehearsal credentials must stay under `/tmp` and be removed with
the rehearsal network.

The FIELD server defaults to loopback. Plain HTTP is rejected for non-loopback
endpoints. Expected blockchain ID, locality ID, VM ID, version, protocol, and
derived actor ID are pinned in its deployment-specific configuration.

## Explicit non-effects

No command in the ordinary build or deterministic test path creates or modifies
a blockchain, L1, Validator Manager, validator set, Authority, Presence,
formation, succession, deployment, release, or public-network resource.
Mainnet, Fuji, and every other public network are outside this crossing.


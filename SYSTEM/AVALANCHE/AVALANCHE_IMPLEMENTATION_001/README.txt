AVALANCHE_IMPLEMENTATION_001 — 1.0.0
VM protocol: PRESENCE_AVALANCHE_VM_001

This directory is the first readable implementation-source crossing. It is
derivative SYSTEM material, not a new root jurisdiction or canonical CSC, DCR,
LINEAGE or Source. See qualification.json for observed results and boundaries.
Compilation, test acceptance and chain receipts do not establish adoption,
formation, succession, Authority or Presence. No release or deployment is made.

Source layout
  vm/                Go RPCChainVM, operation law, CLI and regression tests
  vm/protocol/       single operation/effect/scope contract and current binding
  vm/genesis/        genesis template and schema; CLI materializes exact inputs
  field/             Node FIELD runtime, protected-signature boundary and tests
  integration/       coupled FIELD/real-VM lifecycle testing
  scripts/           reproducible build, binding check, offline validator check
  administrative-input.schema.json   public administrative input contract
  compatibility.json                 pinned dependency/transport declarations
  derivation.json                    read-only predecessor input accounting
  DEPLOYMENT.txt                     architecture, prerequisites and non-effects
  qualification.json                 results, missing observations and ceilings

Build from the repository checkout with Go toolchain downloads available:
  bash SYSTEM/AVALANCHE/AVALANCHE_IMPLEMENTATION_001/scripts/build.sh /tmp/presence-build vm

The merge retains the independently added root-level VM documented in README.md.
build.sh defaults to that root implementation; its second argument "vm" selects
this branch's VM. Their genesis, state schemas and actor namespaces are distinct;
do not reuse state or qualification evidence across them. FIELD selects the
exact identity from expected_vm_id and checks its schema and contract digest.
Both embedded operation tables must agree on every operation/effect/scope entry.
The shared scripts/test.sh validates both source trees and FIELD.

Tests (working directories are material; use the exact successor directories):
  cd SYSTEM/AVALANCHE/AVALANCHE_IMPLEMENTATION_001/vm
  GOTOOLCHAIN=go1.25.13 go test ./...
  GOTOOLCHAIN=go1.25.13 go vet ./...
  cd ../field
  npm test
  cd ..
  bash integration/run.sh

The additional integration/network.sh accepts a pinned node executable, this
VM executable, the vm/rehearsal executable and a public result path. It creates
only a disposable loopback three-validator network, checks the critical FIELD
lifecycle, bounces a validator, restarts all three, and removes temporary keys.
See integration/network-result.json for the observed run and its exact scope.

The FIELD package is private and requires both protocol/ and vm/protocol/.
Do not independently install or distribute it without the same contract bytes.
The contract is an exhaustive array of unique operation/effect/scope entries.
Both runtimes consume it, reject missing/unknown entries and verify their
inventories. Scope is necessary, not sufficient: signatures, actor authority,
payload law, capacity, nonces and accepted-state commitments remain mandatory.
CORRECT uses its target event's exact locus, not incidental present activity.

Current bindings:
  python3 SYSTEM/AVALANCHE/AVALANCHE_IMPLEMENTATION_001/scripts/verify-binding.py
This verifies the stored Source pair, its internal exact-human-byte binding,
all 51 canonical STATE files, and both unchanged predecessor ZIPs.
Source is read at the repository root; no stale duplicate Constitution is
promoted from the archives. Existing SYSTEM/conformance.py remains the canonical
representation evaluator, independent of this implementation.

The VM CLI --materialize-genesis accepts TEMPLATE, ADMINISTRATIVE_INPUT,
PUBLIC_AUTHORITY and OUTPUT; --check-genesis validates the result. The blank
template is deliberately not an instantiable genesis. Use public keys only in
administrative inputs; private signing material belongs in protected external
custody. Disposable test credentials must remain in memory or be removed from
private temporary directories. Never put them in the repository.

Build artifacts go to ignored build/ or an explicitly supplied temporary path.
No install, create-chain, public RPC write, release or tagging command is part of
the build. Source implementation must not be confused with release qualification.

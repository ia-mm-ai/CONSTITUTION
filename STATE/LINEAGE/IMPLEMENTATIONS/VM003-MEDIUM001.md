# VM003–MEDIUM001: bounded historical implementation account

## 1. Status and effect ceiling

This is **one integrated historical account** of `LOCALITY_VM_003` v3.0.0
and its `LOCALITY_MEDIUM_001` v2.0.0 companion, not two independent capability
claims. The VM attempted accepted-state continuity and capacity accounting;
the Medium attempted state-bound access to explicitly routed capabilities and
signed VM transitions. Their source-level connection and historical
compatibility evidence are narrower than a demonstrated end-to-end deployment.

This account establishes **none** of: present CSC, present DCR, live deployment,
a successor L1, current compatibility, Authority, Agency, identity, adoption,
succession, Presence, constitutional validity or Source transfer. It is not a
canonical module, constitutional amendment, implementation standard, release,
formation or privileged embodiment. Predecessor labels such as `PRESENT`,
authority and succession describe their own data and operations, not findings
that those constitutional conditions actually obtained.

The [unsigned machine claim](VM003-MEDIUM001.json) uses the existing
[implementation-claim schema](../SCHEMAS/IMPLEMENTATION-CLAIM.schema.json).
Its only claimed effect is `BOUNDED_IMPLEMENTATION_CLAIM`; its ceiling is
`NO_CONSTITUTIONAL_EFFECT_FROM_CLAIM`. `DECLARED_DEMONSTRATION` is deliberately
absent: `LINEAGE_IMPLEMENTATION_PRESENTNESS` does not allow that effect with
`HISTORICAL`. Attribution is only `ia-mm-ai/PRESENCE repository record`, not
a manufactured Actor, signer, keeper or Authority.

## 2. Exact inputs and reference notation

Starting live `main`, fetched and verified against the agent HEAD before
inspection: **`f03009d259852eb5dc6dc789f209efe6072b5de5`**. This includes
the uploads; the earlier prompt anchor is not substituted for it.

| Carrier: repository-relative path and filename | Bytes | SHA-256 |
|---|---:|---|
| `STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip` | 6201060 | `073949b8b74e7407a433b74efab71f22e2ee93c4d9ed80eb0aeb07b164c08d2b` |
| `STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip` | 260923 | `61353cddc93d62ac28d60e0c7357203ecfe01235e7d249e973ffb16356990989` |

The following **archive locators**, not extracted directories, make citations
exact. Append the text after an alias's colon to its complete prefix; `!/`
crosses a container boundary and `./` is retained where present in TAR names.
All prefixes in this human account are repository-root-relative.

- `V:` = `STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip!/continuous_state/LOCALITY_VM_003/LOCALITY_VM_003-v3.0.0-GITHUB_REPOSITORY_PACK.zip!/`
- `VQ:` = `STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip!/continuous_state/LOCALITY_VM_003/LOCALITY_VM_003-v3.0.0-qualification-evidence.tar.gz!/./`
- `VB:` = `STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip!/continuous_state/LOCALITY_VM_003/LOCALITY_VM_003-v3.0.0-linux-amd64.tar.gz!/LOCALITY_VM_003-v3.0.0-linux-amd64/`
- `VE:` = `VQ:rehearsal/rehearsal-003-20260921T153110Z-I4rgzX/`
- `M:` = `STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip!/local_medium/LOCALITY_MEDIUM_001/LOCALITY_MEDIUM_001-v2.0.0-SOURCE.tar.gz!/`
- `MR:` = `STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip!/local_medium/LOCALITY_MEDIUM_001/LOCALITY_MEDIUM_001-v2.0.0-RUNTIME_PACK.tgz!/package/`
- `MQ:` = `STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip!/local_medium/LOCALITY_MEDIUM_001/LOCALITY_MEDIUM_001-v2.0.0-QUALIFICATION_EVIDENCE.tar.gz!/./`

Machine evidence locators are relative to the machine claim. Its carrier
`identity_refs` use fragment metadata `sha256=…&byte_length=…` to bind the
original ZIP bytes, not to assert authenticity. `core.source` uses the exact
repository-root paths required by its existing schema.

### Internal identities, manifests and evidence

VM identity is `LOCALITY_VM_003`, release `3.0.0`, executable version
`locality-vm/3.0.0`, module `locality.vm/runtime`. `V:RELEASE_MANIFEST.json`
is 5023 bytes, SHA-256
`f3a6e0f5d904b6a1f39924bafe852d5d8e527609a68e07d27839c5d4621def97`.
The nested repository pack is 207315 bytes, SHA-256
`5597eab61802b1df92636b6df88fea46b88f460c4e5f2c51e393c5c6ce7a214f`.
`V:MANIFEST.sha256` covers 65 members; its exact bytes equal
`VQ:SOURCE_MANIFEST.sha256` and `VB:SOURCE_MANIFEST.sha256`.
`VB:PACKAGE_SHA256SUMS` and `VQ:SHA256SUMS` separately bind package/evidence
members. The qualification TAR is 51492 bytes, SHA-256
`170eafe5836ee5d43bd0def635c048a29d9b30937e596a4336378695500aac64`;
`VQ:QUALIFICATION_REPORT.json` is 1653 bytes, SHA-256
`21d28f7cbcc4eb416fbdb73eb2ed6ccbd01866d3409eec3f25ba37506b8dd593`.

Medium identity is `LOCALITY_MEDIUM_001`, release `2.0.0`, private Node ESM
package `locality-medium`. `M:RELEASE_MANIFEST.json` is 4591 bytes, SHA-256
`feb7804a9755eb0ead41b3053c7b25ec3e10259b15e4a53a0d2c857b5087155a`.
`M:MANIFEST.sha256` covers 54 members; the complete source pack has 55 files.
The nested repository pack has the same source paths/bytes, is 103702 bytes,
and has SHA-256
`0962405082e8e07f428605a3091ac2307c921d2632f95ad62dea713103c84e70`.
The outer `local_medium/LOCALITY_MEDIUM_001/LOCALITY_MEDIUM_001-v2.0.0-SHA256SUMS`
binds five artifacts; `MQ:SHA256SUMS` binds six evidence members.
The qualification TAR is 4174 bytes, SHA-256
`c7aba43baafa41374d75e7563e6079a1feff097904c08d8c84e7792d256a521c`;
`MQ:LOCALITY_MEDIUM_001-v2.0.0-qualification.json` is 2688 bytes, SHA-256
`ed51b4e16b835aaed5c8d2f7e2c17900aa9a100186496b6262ab48c437ef6922`.
`M:lineage/PREDECESSOR_LOCK.json` and `M:RELEASE_MANIFEST.json` explicitly
bind the VM repository-pack digest above: this is a concrete cross-carrier
relation, not merely similar names.

### Current Source versus historical copies

Current semantic authority remains the unchanged repository `SOURCE/` pair:

| Repository-relative current Source | Bytes | SHA-256 |
|---|---:|---|
| `SOURCE/CONSTITUTION_0()1.md` | 44133 | `affeb5738cdfeea7ee4fe985652bf78b15eb6dfbba5871d82f0f2213a81832d0` |
| `SOURCE/CONSTITUTION_0()1.json` | 146048 | `519a81d2a26d5bf32afd77566bbb35b4d28b22af85ac45a7055fccd1b45222f8` |

Its machine Form's `binding.human_path`, exact-byte length and SHA-256 resolve
to and match the human Form. No Form was regenerated or normalized.

Both predecessors instead carry the following identical **historical copies**
at `V:source/…` and `M:source/…` (also in their packaged duplicates):

| Archived filename | Bytes | SHA-256 |
|---|---:|---|
| `CONSTITUTION_0()1.md` | 41268 | `5dea16e2400c652eb063129711a6bf26c8485a2cbad25ce261beda6f7fee5206` |
| `CONSTITUTION_0()1.json` | 139114 | `5a0d6ffec23d09b203212b295d86e45e141a33a7f35f525d56172fb848d1a73c` |
| `FORM_BINDING.json` | 1901 | `cc2f4f2a1fdfab3287ee661dfbcc6e9db72c8e2ac918a94c288598f27916e758` |

Their internal binding matches those older human bytes, not current Source.
The archived origin declaration names `ia-mm-ai/CONSTITUTION` commit
`4cf5a926d5fece4e3cccfdfcab40e16431dd332b`; publication/authenticity was not
remotely reverified. Current-only stable references include `section.13.7`,
`section.16.7`, `section.16.8`, `section.19.3`, `section.19.4`,
`article.4.table.row.22`, `article.4.table.row.23`,
`article.5.table.row.27` and `article.5.table.row.28`; the old machine copy
instead includes `trailer.1`. Consequently the historical copy cannot silently
supply the later core's full CSC/DCR basis.

The unchanged first-pass indexes are
[VM003-TO-CSC](../RECORDS/VM003-TO-CSC.json) and
[MEDIUM-TO-DCR](../RECORDS/MEDIUM-TO-DCR.json). All 109 and 156 indexed byte
bindings, respectively, were independently checked against the carriers.
Their inventories and conclusions were not regenerated.

## 3. Concrete composition

### VM and accepted execution

`V:main.go` serves an Avalanche RPCChainVM process; `V:vm.go` implements
Snowman ChainVM. The VM-ID domain is `LOCALITY_VM_003`, yielding Avalanche ID
`25tZjky6SecZA1Gwc6VwD2ouy64xAUdgNLo1C8dkXfbsFNxaTk`.
Plugin installation names a binary with a supplied VM ID; installation and
preflight do not themselves compare that filename with its embedded VM ID
(`V:scripts/install-plugin.sh`, `V:scripts/preflight-node.sh`).

`V:genesis.go` and `V:genesis/LOCALITY_RUNTIME_GENESIS_TEMPLATE_003.json`
describe a non-EVM signed-event state machine. Genesis contains domain,
lineage, constitution, locality, profile, capacity, continuity, initial-state,
claim-limit and evolution objects. Release/form version `3.0.0`, `_003`
identifiers and numeric genesis/evolution-state version **4** are different
version axes. The template's locality/source/public-authority deployment
fields remain unmaterialized. Its defaults are structural/actual capacity
1000, correction-egress reserve 100, maximum admission offer 86400 seconds,
and continuity-authority minimum/successor threshold 2. Materialization
requires a descriptor and public authority; no genesis was materialized here.

`V:block.go` uses `LOCALITY_BLOCK_003` with parent ID, Unix-second timestamp,
uint64 height and 1–64 transitions, rejecting duplicate transition IDs.
Genesis block ID hashes exact genesis bytes. Other block IDs hash exact canonical
block bytes. Avalanche block-ID string encoding is distinct from hexadecimal
state commitments and transition IDs.

`V:state.go`, `V:operations.go` and `V:transition.go` distinguish presentation,
host gate disposition, entry, crossing, matter disposition, emergence,
correction, departure checkpoint, exit, closure/residue and renewed entry.
Presentation is not admission; an offer is not entry. Entry reserves work
and lawful resolution. Correction appends rather than rewriting its targeted
event. Closure preserves addressed residue; incorporation is a separate,
validated operation.

Capacity accounting subtracts unresolved work and reserved correction/egress
from effective actual capacity, bounded by the active locus. A deficit
changes posture to `HOLD_CAPACITY_DEFICIT`; selected correction/egress
operations remain possible. Pulses bind a reported current pre-state;
neither a pulse nor the predecessor's `PRESENT` count proves constitutional
Presence. Formation authority can register continuity authorities; exhaustion
and threshold-attested successor freeze have explicit dormant/no-open-locus
conditions. Freeze neither executes unknown successor code nor transfers
keys, State, identity or Authority.

### Medium, profiles and controlled surfaces

`M:src/runtime.js` assembles configuration, profile, journal, VM client and
engine. `M:package.json` declares Node ESM, Node >=22 and no third-party
runtime/package dependencies. There is no independent consensus engine here.
The four profiles in `M:profiles/` are operational allowlists:
`HUMAN_MEDIUM_001.json` exposes a local channel;
`AI_MEDIUM_001.json` adds a bounded tool broker;
`DEVICE_MEDIUM_001.json` exposes command/sensor-commitment channels;
`LOCALITY_MEDIUM_001.json` adds gateway/local-effect and host operations.
They are not ontological classifications or control of an entire person,
AI, device or locality.

`M:src/state.js` derives phases from accepted VM state: no locus,
presentation, admission/reentry availability, entry, checkpoint, ending,
capacity deficit and successor freeze. It checks selected capacity fields
as nonnegative safe integers, but **does not recompute VM capacity arithmetic**.
An open offer must not have expired; reentry requires a unique unconsumed
sealed checkpoint. Capacity deficit with an active locus overrides ordinary
entry phase. Profile HOLD allowlists retain correction/checkpoint/exit; the
host profile also retains selected gate/close/capacity/pulse/reclaim actions.
Other restrictions, such as pending transitions or wrong keys, still apply.

`M:src/broker.js` can call a registered handler once after a fresh guard.
HTTP authorization alone does not execute or constrain an external tool.
`M:examples/ai-broker.mjs` deliberately throws until a real handler is supplied.
No continuous supervision, cancellation of an already-running handler, or
atomic accepted-state/external-effect transaction is implemented.

## 4. Actual coupling, boundary by boundary

These are source-observed invocation paths, **not a recovered live integrated
trace**. Tests and historical reports are bounded separately in section 8.

1. **Encounter/request → Medium decision.** Producer: a loopback HTTP caller
   or registered broker caller. Consumer: `M:src/server.js`/`M:src/engine.js`.
   `POST /v1/authorize` carries `{capability, expected_state_commitment?}`;
   broker invocation carries a capability and handler input. The HTTP server
   also exposes `GET /health`, `GET /v1/context`, `POST /v1/drafts`,
   `POST /v1/transitions`, `GET /v1/receipts/<id>`. It rejects queries,
   unknown routes and non-object/empty/over-1-MiB request bodies. Authorization
   returns a public decision (200/403), not a transition or grant of Authority.
   Handler absence or failed guard prevents invocation. Evidence:
   `M:src/server.js`, `M:src/broker.js`, `M:test/server.test.js`
   (stub engine), `M:test/engine.test.js` (mocked VM).

2. **Accepted VM views → Medium capacity/phase gate.** Producer: VM handlers
   `GET /status` and `GET /state`; consumer: `M:src/engine.js` through
   `M:src/vm-client.js` native HTTP `fetch`. Views are fetched concurrently.
   The engine checks health, frozen version/protocol/form/locality,
   revision/commitment, active locus, body posture/count, successor-boundary
   truthiness and safe numeric fields. It intersects phase/profile capability
   and VM-operation allowlists with actor/key, pending and optional expected
   commitment restrictions. It journals `CAPABILITY_DECISION`, including
   denial, before returning. Transport, incoherent snapshots and journal
   failure propagate; there is no atomic snapshot/retry protocol. The engine
   neither recomputes state commitments, compares both capacity objects,
   authenticates chain identity nor checks authority roles. The configured VM
   ID is compared with a constant, not remote attestation; its genesis client
   method is unused by the engine. Evidence: `M:src/engine.js`,
   `M:src/state.js`, `V:api.go`, `M:test/state.test.js`,
   `M:test/engine.test.js`, `M:test/vm-client.test.js`.

3. **Authorized draft request → VM unsigned envelope → Medium draft.**
   Caller supplies exactly `{operation,locus_id,observed_at,payload}` to
   `POST /v1/drafts`. Medium obtains fresh authorization, checks locus/time/
   object shape, then sends `{operation,actor_id,actor_public_key,locus_id,
   observed_at,payload}` to VM `POST /drafts`. VM derives unsigned context
   from preferred state plus pending preview, not necessarily accepted state
   alone. VM drafting validates syntax/identity/time but does not apply the
   requested operation. Medium compares returned operation, actor/key, locus,
   time, effect, payload, next revision and previous accepted commitment,
   validates unsigned shape, stores a process-local draft and journals
   `DRAFT_PREPARED`. It returns the unsigned object, draft commitment,
   signing-byte hash and `DRAFT_ONLY_NO_STATE_CHANGE`. No accepted State
   changes. Mismatch refuses; draft insertion precedes journal append and can
   survive an append failure in memory. Medium does not independently check
   nonce against State or operation-specific payload law. Evidence:
   `M:src/engine.js`, `M:src/signature.js`, `V:signing.go`, `V:api.go`,
   `M:test/engine.test.js`, `V:vm_test.go`. There is no separate proposal
   service: `PROPOSE_SUCCESSOR` is a generic host operation, not automatically
   generated by a capacity decision.

4. **External signer → Medium → VM pending submission.** The signer is
   outside the Medium runtime. It supplies `{unsigned,signature}` to
   `POST /v1/transitions`. The unsigned field order is `schema, operation,
   revision, previous_state_commitment, actor_id, actor_public_key, nonce,
   locus_id, observed_at, effect, payload`. The schema is
   `LOCALITY_TRANSITION_003`; no explicit chain/network-ID field is included.
   Ed25519 signs the serialized unsigned bytes, not its displayed hash.
   Public keys are 32 bytes and signatures 64 bytes, represented as lowercase
   hex. Participant IDs derive from the first 20 SHA-256 bytes of public keys.
   Medium verifies signature/configured actor and a matching in-memory draft,
   then repeats authorization at the predecessor commitment. It posts to
   VM `POST /transitions` and requires the locally computed transition ID and
   `PENDING_CONSENSUS`; only then it deletes the draft and journals
   `TRANSITION_SUBMITTED`. VM returns 202 with
   `NO_ACCEPTANCE_UNTIL_BLOCK_ACCEPTED`. Invalid signatures, stale drafts,
   transport errors or mismatched responses fail; submission can already
   have occurred before a response/journal failure. Restart loses drafts;
   no time-based draft-expiry mechanism is implemented. Evidence:
   `M:src/signature.js`, `M:src/engine.js`, `M:src/vm-client.js`,
   `V:transition.go`, `V:api.go`, `M:test/engine.test.js`,
   `V:state_test.go`, `V:api_test.go`.

5. **VM pending transition → consensus block → accepted application State.**
   Producer: VM pending pool/block builder; consumer: Avalanche consensus
   callbacks and `V:vm.go`/`V:state.go`. Canonical signed bytes and operation
   payloads must match Go serialization. Validation checks signature,
   operation effect, next global revision, predecessor commitment, monotonic
   time, actor-key/actor-specific nonce, operation authority, capacity and
   operation-specific state. Block verification checks parent/height/time and
   applies transitions to the parent snapshot. Acceptance stages block,
   State, height, tip, transitions and receipts through `versiondb` into the
   node-provided database. Successful application advances revision/nonce,
   appends event/receipt and recomputes posture/commitment. Invalid proposals
   do not establish accepted State. A build/reject/requeue path exists, but
   commit-error recovery is not universally atomic: pending entries are
   removed before storage commit without restoration on commit failure.
   Evidence: `V:canonical.go`, `V:transition.go`, `V:operations.go`,
   `V:state.go`, `V:block.go`, `V:vm.go`, `V:vm_test.go`,
   `V:continuity_reentry_test.go`, `V:metabolism_test.go`.

6. **Accepted receipt → Medium observer → journal.** VM `GET /receipts/<id>`
   returns a direct `LOCALITY_RECEIPT_003` object; no receipt yet gives 404.
   Medium polls every 300 ms until its configured timeout, then checks schema,
   requested ID, actor, digest/revision and operation/effect before fetching
   fresh coherent State. Ahead-of-State receipts or same-revision commitment
   mismatch fail. Older receipts receive no historical ancestry proof and
   need not correspond to a submission retained by this process.
   `RECEIPT_OBSERVED` is journaled before return. Timeout means unaccepted
   **or unavailable**, not rejected; journal failure cannot reverse
   acceptance. A receipt grants no present capability. Evidence:
   `V:api.go`, `V:state.go`, `M:src/engine.js`, `M:src/vm-client.js`,
   `M:src/journal.js`, `M:test/engine.test.js`,
   `M:test/vm-client.test.js`, `M:test/journal.test.js`.

7. **Departure/carried-state material → renewed entry.** Medium continuity
   helpers/CLI can produce ordered commitment passages; VM consumes them in
   checkpoint/reentry payloads through the same signed-transition path.
   Steps bind participant, ordinal (1–128), predecessor, delta digest and
   derived successor. Empty passage is valid only with unchanged result.
   VM requires checkpoint-before-exit, sealing, fresh presentation/admission,
   an ended previous entry, same actor, correct residue where applicable,
   fresh reservation and single-use checkpoint consumption. The helpers do
   not inspect private delta bytes and the Medium engine does not itself
   invoke them to validate `REENTER`. Private-State semantics, lawful
   resumption and real signer-to-VM passage remain outside the proof.
   Evidence: `M:src/continuity.js`, `M:bin/locality-medium.js`,
   `M:test/continuity.test.js`, `V:operations.go`,
   `V:transition.go`, `V:continuity_reentry_test.go`.

**UNSUPPORTED as a demonstrated coupling:** a complete real encounter →
Medium guard → protected signer → VM consensus acceptance → matching Medium
receipt, including a real controlled handler and renewed entry. Medium tests
stop at mocked pending submission; their fabricated receipt test is separate.
The VM's network rehearsal uses its own client, not the Medium. The actual
historical bridge is Medium inspection of VM-produced evidence, not a live
joint run. Missing coupling is preserved, not filled in.

## 5. Concrete State and data forms

- **Persisted application State:** `LOCALITY_RUNTIME_STATE_003`
  (`V:state.go`) includes revision, state commitment, last transition/time,
  host locality/profile, active locus, body/capacity, authorities/formation
  authority, pulses, successor proposals/boundary, actor keys/next nonces,
  presentation/entry histories, departure checkpoints, residue history,
  loci/events/corrections and incorporated residues. Its commitment hashes
  Go-serialized State with its own commitment field empty. Historical entry
  and consequence records are not erased by ending or correction.
- **Execution/consensus metadata:** `V:vm.go` stores exact genesis bytes/ID,
  last accepted tip, current State, block and per-block snapshot, height
  index, transition and receipt. Keys use `locality/meta/`,
  `locality/state/current`, `locality/block/`, `locality/state/block/`,
  `locality/height/`, `locality/transition/`, `locality/receipt/`; height
  suffixes are big-endian uint64, block suffixes raw IDs, transition/receipt
  suffixes hex IDs. Pending work is not persisted.
- **Restart:** initialization requires identical stored genesis. For a
  non-genesis accepted tip it loads that block's State snapshot and checks
  its commitment and invariants; at a genesis tip it instead uses freshly
  reconstructed genesis State, without loading that persisted snapshot.
  This is **not replay of all application events from genesis**, even where
  qualification gate wording says replay. `locality/state/current` is
  written, but non-genesis restart reads the per-block snapshot. Lost storage,
  incompatible genesis or corrupted snapshots are not repaired by custody
  of these archives.
- **Transport/signature objects:** unsigned/signed transitions, admission
  requests, status/state views and pending replies are not canonical CSC/DCR
  representations. Go signing/canonical transition bytes use `json.Marshal`;
  Medium uses ordered-object `JSON.stringify`. Medium draft commitments
  instead use recursively key-sorted JSON. Transition IDs hash the signed
  envelope. These different hashes must not be interchanged. JavaScript safe
  integers are narrower than Go uint64/int64. There is a **source-observed
  conditional incompatibility**: Go escapes `<`, `>`, `&`, U+2028 and U+2029,
  while JavaScript leaves them literal. Valid presentation strings can
  contain them; Medium reserialization then fails VM's canonical-byte check,
  and a Go-byte signature can fail Medium verification. Genuine VM drafts
  otherwise preserve typed payload field order; this is not universal
  incompatibility or a claim about all non-ASCII text. No cross-runtime
  failure was executed here (`V:canonical.go`, `V:transition.go`,
  `M:src/signature.js`, `M:src/vm-client.js`).
- **Journal and receipts:** `M:src/journal.js` persists a synced NDJSON chain
  with schema, sequence, previous hash, timestamp, type, public fields and
  event hash. It verifies its chain on opening, checks restrictive file
  permissions, serializes appends within one instance and rejects reserved
  or content-bearing field names. This is neither multi-process locking nor
  externally anchored chronology; a permitted field's value is not thereby
  proven public. It contains decisions/draft/submission/receipt metadata,
  not the VM State or private payload. A receipt is an accepted-operation
  representation, not a signature, identity or current-capacity oracle.
- **Checkpoint/dormancy/reentry:** departure checkpoints carry commitment
  relations and change OPEN → SEALED → CONSUMED under the appropriate
  operations. `DORMANT_P0` does not fabricate active participants.
  Deterministic private-state succession hashes preserve edges, not private
  state contents, truth of deltas or constitutional resumption.
- **Evidence and configuration:** genesis/profile/configuration objects
  select one deployment's concrete parameters; templates are not accepted
  State. Qualification snapshots, test output, manifests and build metadata
  are historical evidence carriers. The rehearsal's sorted-complete-State
  digest is distinct from the VM commitment with its self-field set to empty.
  **Canonical State representations remain the unchanged CSC/DCR/LINEAGE
  contracts**, not any of these predecessor wire or storage forms.

## 6. Domain-to-core map

Rule IDs refer to the unchanged
[CSC model](../../CONTINUITY-STATE-CAPABILITY/CONTINUITY-STATE-CAPABILITY.json),
[DCR model](../../DYNAMIC-CAPACITY-REGULATION/DYNAMIC-CAPACITY-REGULATION.json)
and [LINEAGE model](../LINEAGE.json). Relations below are bounded aspects,
limits or gaps, not assertions that predecessor objects satisfy those rules'
complete schemas or predicates. A similarity is not implementation; a test
is not present capability; a mapped rule is not constitutional conformance.

Labels are documentation metadata only: `ARCHIVE_DECLARED` describes a carrier's
claim; `SOURCE_OBSERVED` inspected logic; `TEST_DEMONSTRATED` an archived test
plus historical report at its tested conditions, never a fresh run;
`HISTORICAL_QUALIFICATION` archived output; `CONFLICTED` incompatible accounts;
`UNSUPPORTED` absent support for the stated scope; `UNRESOLVED` an open question.

| Archive/component | Concrete mechanism | Data/form | Canonical relation | Evidence | Observed status | Failure/effect ceiling |
|---|---|---|---|---|---|---|
| VM transition/application | Next revision, previous commitment and signed actor nonce | `LOCALITY_TRANSITION_003`, state events | `csc.transition-links`, bounded endpoint-link aspect | `V:transition.go`, `V:state.go`, `V:state_test.go` | `SOURCE_OBSERVED` | No canonical transition/basis proof; a signature is not Authority |
| VM correction | Same-actor targeted append | Corrections and retained events | `csc.correction-links`, `csc.correction-history` | `V:operations.go`, `V:state_test.go` | `SOURCE_OBSERVED` | Does not establish repair of every consequence or relation |
| VM persistence | Per-block snapshot and exact-genesis restart | Accepted tip, State, receipts | `csc.change-history`, `csc.state-discontinuity`, preservation aspects only | `V:vm.go`, `V:vm_test.go`, `VE:LOCALITY_REHEARSAL_003_POST_RESTART.json` | `HISTORICAL_QUALIFICATION` | Snapshot equality only; universal commit-failure atomicity unsupported |
| VM + Medium continuity | Checkpoint, deterministic passage and fresh reentry | Commitment edges, SEALED/CONSUMED checkpoint | `csc.resumption-binding`, `csc.replacement-ceiling`, `LINEAGE_REENTRY_EFFECTS` | `V:continuity_reentry_test.go`, `V:operations.go`, `M:src/continuity.js`, `M:test/continuity.test.js` | `SOURCE_OBSERVED` | No private delta semantics, present Agency or inherited identity |
| VM capacity + Medium phase | Capacity reservation and deficit-based allowlists | VM capacity/body fields; profile phase | `DCR-C01`, `DCR-C05`, scoped-support aspects | `V:state.go`, `V:operations.go`, `M:src/state.js`, `M:profiles/LOCALITY_MEDIUM_001.json` | `SOURCE_OBSERVED` | Signed local units/posture are not full structural/situated evidence |
| Medium fresh gate | Re-observe accepted revision/commitment at invocation | Public decision and optional expected commitment | `DCR-C04`, `DCR-D01` | `M:src/engine.js`, `M:src/broker.js`, `M:test/engine.test.js` | `TEST_DEMONSTRATED` | Mocked exit/pending checks; no atomic external effect or continuous guard |
| Medium HOLD profile | Narrow correction/egress remain routable | `HOLD_CAPACITY_DEFICIT` phase | `DCR-D03`, `DCR-H01`, `DCR-H02`, `DCR-H03`, `DCR-H04`, `DCR-H05`: lifecycle gap, not implementation of these rules | `M:src/state.js`, `M:src/engine.js`, `M:test/state.test.js` | `UNSUPPORTED` | No portable attributable HOLD review/correction/release record lifecycle |
| VM authority checks + Medium signed submission | Capability/nonce law separate from key-byte check | Actor binding, authority maps, signed envelope | `DCR-A01`, `DCR-M02`, separation aspect only | `V:state.go`, `V:operations.go`, `M:src/signature.js` | `SOURCE_OBSERVED` | Medium does not establish authority roles; local signature does not establish rightful effect |
| Medium receipt + journal | Accepted-view observation and retained metadata | Receipt, hash-chained NDJSON | `LINEAGE_OCCURRENCE_DISTINCTIONS`, `LINEAGE_OCCURRENCE_NON_CONVERSION`, `DCR-E05` | `M:src/engine.js`, `M:src/journal.js`, `V:api.go`, `M:test/journal.test.js` | `SOURCE_OBSERVED` | No original-submission/older-receipt ancestry proof; append failure cannot undo VM effect |
| Paired qualification bridge | Medium reads exact VM revision-18 evidence | VM evidence digest and compatibility report | `csc.historical-ceiling`, `DCR-E01`, `DCR-E05`, `LINEAGE_IMPLEMENTATION_PROOF_CEILING` | `VE:LOCALITY_REHEARSAL_003_EVIDENCE.json`, `MQ:LOCALITY_MEDIUM_001-v2.0.0-vm003-evidence-compatibility.json`, `M:scripts/verify-vm003-evidence.mjs` | `HISTORICAL_QUALIFICATION` | Public-state compatibility only, expressly not live coupling |
| Paired configuration/identity | Frozen target but stale operator instructions/template | Plugin ID; required VM version | `LINEAGE_IMPLEMENTATION_EXACT_BINDING`, claim limitation | `V:SECURITY.md`, `V:docs/DEPLOYMENT_HANDOFF.md`, `V:RELEASE_MANIFEST.json`, `M:examples/config.template.json`, `M:src/config.js` | `CONFLICTED` | Neither side silently selected as deployable guidance |
| Paired release claims | Reference/production-intent declarations | Release manifests and archived Source binding | `LINEAGE_IMPLEMENTATION_NO_SELF_VALIDATION`, `LINEAGE_IMPLEMENTATION_PRESENTNESS` | `V:RELEASE_MANIFEST.json`, `M:RELEASE_MANIFEST.json`, both `source/FORM_BINDING.json` files | `ARCHIVE_DECLARED` | Releases, custody and historical Source copies confer no constitutional effect |

## 7. Dependencies and compatibility conditions

These are **declared historical dependencies**, not installation instructions
or present compatibility/safety recommendations.

- **Go/node process:** `V:go.mod`, `V:go.sum`, `V:rehearsal/go.mod`,
  `V:RELEASE_MANIFEST.json`, `VQ:VM_BUILD_INFO.txt` and
  `VQ:AVALANCHEGO_BUILD_INFO.txt` fix Go **1.25.13**, AvalancheGo **v1.15.0**
  at peeled commit `70bd6d063b7343fd2cd8217200aaf77b57f19f68`,
  RPCChainVM **46**, and profile **v1.15.0+LOCALITY_SECURITY_OVERLAY_001**.
  `V:compat/AVALANCHEGO_SECURITY_OVERLAY_001.json` pins authenticated base/
  graft module sums and before/after module hashes; it applies gRPC
  **v1.83.2** and resolved dependencies including `x/crypto` **v0.55.0**,
  `x/net` **v0.58.0**, `x/sys` **v0.47.0**, `x/text` **v0.41.0**.
  Full transitive versions/sums remain in those exact manifests and
  `VB:DEPENDENCIES.txt`, not a new copied dependency inventory.
- **Platform/build:** canonical qualification target is Ubuntu **24.04 amd64**;
  output reports Linux **x86_64**. Native build uses **CGO_ENABLED=1**,
  `-buildvcs=false`, `-trimpath`, stripped linker flags and recorded
  **GOAMD64=v1**. Go-toolchain archive SHA-256 is
  `39042a078ea9ceebe3ecda4a7188f0f5b96e14a071d27923ba7f40b456e85ae3`;
  declared scanner is `golang.org/x/vuln/cmd/govulncheck@v1.7.0`.
  Shell/Python/Go tooling prepares the node overlay, verifies source,
  materializes genesis and packages qualification
  (`V:scripts/build.sh`, `V:scripts/verify.sh`,
  `V:scripts/prepare-avalanchego.sh`, `V:scripts/package-release.sh`,
  `V:rehearsal/run_lifecycle.py`). Nothing here was invoked.
- **Medium:** Node **>=22**, built-in crypto/HTTP/filesystem/fetch, ESM and
  no declared third-party dependencies (`M:package.json`); CI specifies
  Node **22**, historical qualification reports **v24.19.0**, Linux
  **6.18.44 x86_64**. `M:scripts/verify.mjs`, `M:scripts/qualify.mjs`,
  `M:scripts/seal.mjs` and package scripts describe source/profile checks,
  tests/coverage, predecessor locks and packaging. Runtime pack retains
  package scripts and a source manifest but omits their tests/tooling:
  37 files, 36 matching manifest entries, 18 listed entries absent.
- **Transport/time:** RPCChainVM is the plugin/process protocol. Application
  routes are ordinary HTTP JSON, **not JSON-RPC methods**.
  VM's named routes additionally include `/genesis`, `/capacity`,
  `/currentness`, `/transitions/{id}` and `/blocks/{id}`. Historical node
  routing supports `/ext/bc/<blockchainID>/…` and direct routing with
  `Avalanche-Api-Route`; node `/ext/info` tooling is separately JSON-RPC.
  Medium uses the configured path route, no redirect following, bounded
  fetch/receipt timeouts and response sizes. HTTPS is required except for
  explicitly allowed IP-loopback HTTP; credentials/query/fragment in the
  endpoint are rejected, and a path segment must match the blockchain ID.
  This is not full chain authentication. Medium serves only IPv4/IPv6
  loopback (default port **8787**); request timeout range **250–60000 ms**,
  receipt timeout **1000–600000 ms**, response limit **1 KiB–16 MiB**.
  Local clock observations govern offer/timeout checks, not timeless truth
  (`M:src/config.js`, `M:src/vm-client.js`, `M:src/server.js`, `V:api.go`).
- **Storage/process availability:** AvalancheGo, consensus peers and its
  database are external process/storage requirements; the VM chooses no
  independent backend. The Medium additionally needs a writable restricted
  journal and in-memory draft state. Database loss breaks accepted-state
  recovery; process loss drops pending/draft material; journal loss removes
  this local evidence trail. No rollback of external consequence follows.
- **Keys/profiles:** Ed25519 public actor bindings, an external protected
  signer, separately governed formation/continuity roles and node credentials
  are concrete dependencies, not rightful identity. Medium loads validated/
  frozen operational profiles; release-profile digest verification belongs
  to qualification tooling, not runtime profile loading. A key, signer,
  profile, handler or authority change requires a new exact basis; no
  automatic custody/Authority/identity inheritance is supplied.

Loss of node, transport, signer, matching State, storage or supported profile
can block or invalidate this path; failure does not establish the absence of
external reality. Replacement needs fresh version/wire/behavior qualification,
deployment identity and independent constitutional basis. No dependency is
made sovereign, universally required or privileged by this record.

## 8. Demonstrations and proof ceilings

**This crossing:** read-only source/test/report inspection and byte/schema/
reference checks only. No predecessor tests or binaries were executed.
`SOURCE_OBSERVED` is not an observation of a running implementation.

**VM unit-test source:** 18 named Go tests and three parser fuzz targets cover
canonical envelopes, lifecycle distinctions, capacity/egress, offers,
checkpoint reuse/reentry, authority exhaustion/successor freeze, HTTP behavior,
restart/tampering, rejection requeue and drafting. Relevant files are
`V:state_test.go`, `V:vm_test.go`, `V:api_test.go`,
`V:authority_test.go`, `V:metabolism_test.go`,
`V:continuity_reentry_test.go`, `V:main_test.go`, `V:fuzz_test.go`.
Restart unit tests reuse a memory database; they do not simulate disk-commit
faults. Normal test/race gates are not a sustained fuzz campaign.

**VM historical qualification:** `VQ:QUALIFICATION_REPORT.json` reports PASS
at **2026-09-21T15:34:29Z**, including unit/race/vet/build/node-profile gates.
`VE:LOCALITY_REHEARSAL_003_EVIDENCE.json` is 91648 bytes, SHA-256
`36202aafef90a5409350ed5ceb9a46447ad3a568676163b2e58f09a1b08a4fc0`.
It reports a disposable three-node network, 18 accepted transitions and
height/revision 18; final VM State commitment is
`be615433b6e6eb3a9c6c2284ddb565aae2718f0f45c80a3c1ed5647fe3203a64`.
`VE:LOCALITY_REHEARSAL_003_LIFECYCLE.json`,
`VE:LOCALITY_REHEARSAL_003_POST_RESTART.json` and
`VE:LOCALITY_REHEARSAL_003_POST_REJOIN.json` retain lifecycle/full-network
restart/single-node rejoin snapshots. The latter observations are timestamped
**15:34:23.765988Z** and **15:34:28.559205Z** on that date. Their complete
sorted-State digest
`cb8b8cee6d8fac7a49e3e67aa51356d46d663e88f20d3f53d714ed0deebd49b7`
matches across those snapshots; this is not the VM commitment algorithm.
The network trace's final state still has one authority, no successor
proposals and null successor boundary: it does not demonstrate threshold
succession. State-test gates must not be recast as network observations.

**Medium tests and reports:** `M:test/engine.test.js` uses a mocked VM;
its draft/sign/submission test stops at pending consensus. Its receipt test
uses a separate fabricated receipt. `M:test/server.test.js` uses a stub
engine; `M:test/vm-client.test.js` uses an HTTP fixture. Other test files
cover CLI/config/profiles, continuity edges, phase and journal behavior.
`MQ:LOCALITY_MEDIUM_001-v2.0.0-tests.txt` reports **20 passed**;
`MQ:LOCALITY_MEDIUM_001-v2.0.0-coverage.txt` reports **86.65% lines,
58.63% branches, 84.78% functions**. None was rerun here.
The qualification JSON has **no exact qualification timestamp**. The outer
`local_medium/LOCALITY_MEDIUM_001/LOCALITY_MEDIUM_001-v2.0.0-CROSSING_REPORT.md`
reports sealing at **2026-09-21T15:56:24Z**; this is not silently converted
into a test time. Archive timestamp metadata supplies no missing precision.

**The verified cross-carrier evidence bridge:**
`MQ:LOCALITY_MEDIUM_001-v2.0.0-vm003-evidence-compatibility.json` (801 bytes,
SHA-256 `cac8fc2ea2995f4a6b75f73a6a6b44dedd0179ab1c813876b9da16a33007824c`)
references the exact VM evidence digest above, derives renewed-entry
`PRESENT` with a `CONSUMED` checkpoint at revision 18 and records restart/
rejoin preservation. The referenced VM evidence is actually in the other
carrier and was verified there. `M:scripts/verify-vm003-evidence.mjs` reads
that evidence, not a live VM/Medium/signer. Its restart value is reported;
the script explicitly requires rejoin preservation, not a separate successful
live restart experiment. `M:scripts/qualify.mjs`'s “wire schema exact” gate
checks source substrings after digest locking, not exhaustive Go/JS wire
equivalence. Both the report and qualification expressly say
**`NOT_LIVE_SUCCESSOR_L1_COUPLING` / `NOT_RUN_DEPLOYMENT_SPECIFIC`**.

**Binary metadata, not execution:** `VB:locality-vm` is a 15791056-byte
Linux x86-64 ELF, SHA-256
`b7cddc8619010d8aa542d6cc37b1be2e17a497cecf832b4a798b03206726f364`,
matching the historical report. The reported AvalancheGo executable digest
is not independently checked against an enclosed node binary: none is
bundled. `VQ:VM_VULNERABILITY_SCAN.txt` and
`VQ:AVALANCHEGO_VULNERABILITY_SCAN.txt` report zero reachable findings but
three findings in required modules not apparently called. They establish
neither zero vulnerable dependencies nor current safety.

## 9. Preserved conflicts, omissions and unsupported claims

1. **CONFLICTED — VM identity/version family in operator documentation.**
   `V:SECURITY.md` and `V:docs/DEPLOYMENT_HANDOFF.md` still specify plugin ID
   `2JnfZqeNUW34DmiSznMp1oJSpX5Fpqv9Ms6BYEyGF3LYVx7jQ5` (VM002), while
   VM003 runtime, README, manifests and qualification specify
   `25tZjky6SecZA1Gwc6VwD2ouy64xAUdgNLo1C8dkXfbsFNxaTk` and version
   `locality-vm/3.0.0`. The earlier broad identity/version-conflict finding
   is retained at this exact verified scope: references expressly describing
   VM002 v2.1.0 as ancestry are not themselves stale runtime declarations.
   Numeric genesis/schema version 4 is another axis, not an inferred fix.
   Installation/preflight do not automatically resolve the mismatched ID.
2. **CONFLICTED — Medium template versus runtime.**
   `MR:examples/config.template.json` requires `locality-vm/2.1.0`;
   `M:src/constants.js` and `M:src/config.js` require `locality-vm/3.0.0`.
   Filling placeholders alone is insufficient. Tests construct configurations
   from current constants rather than validating that template. Neither
   predecessor side was repaired.
3. **UNSUPPORTED — universal persistence atomicity.** Database batching does
   not make pending-map mutation plus storage acceptance universally atomic.
   `V:vm.go` removes selected pending entries before commit and has no
   restoration on commit failure. No injected commit-fault test establishes
   the stronger claim. Medium journaling likewise cannot roll back a VM
   submission made before an append failure.
4. **UNSUPPORTED — portable HOLD lifecycle.** Profile gating and phase changes
   provide no complete held-request identity, attributed review, corrected
   basis, preserved lifecycle and current-grounded release/resumption chain.
   Offer expiry, receipt timeout and an allowlist for correction are not
   substitutes for DCR's HOLD lifecycle.
5. **CONFLICTED as a proposed current binding — historical Source.** The older
   pair is internally bound but differs from current Source. Qualification
   of its quotations cannot qualify the newer Source or canonical contracts.
6. **UNSUPPORTED — present capability/deployment.** Archived PASS reports,
   tests, restart/rejoin, source ancestry and manifests establish no live L1,
   present CSC/DCR, current compatibility, adoption, succession or Presence.
   The Medium handoff expressly leaves real signer/handlers and complete live
   passage outstanding (`M:docs/DEPLOYMENT_HANDOFF.md`).
7. **CONFLICTED — conditional wire compatibility.** Go/JavaScript escaping
   differs for `<`, `>`, `&`, U+2028 and U+2029 in otherwise permitted strings;
   canonical-byte/signature checks can therefore reject this coupled path.
   This is static source evidence, not a reproduced runtime result.
   Compatible restricted payloads are not thereby disproved. Full-range
   uint64/int64 interchange is also unsupported by Medium's safe integers.
8. **UNRESOLVED — wider failure coverage.** Lost storage/keys, partitions,
   conflicting blocks and resource exhaustion are not settled by these
   reports. Runtime profile-digest enforcement and older-receipt ancestry
   verification are absent from the inspected Medium path. Hashes, signature
   checks and positive labels do not fill missing evidence or attest
   independent constitutional grounds.

## 10. Lawful derivation boundary

Another implementation may learn concrete patterns for exact predecessor
binding, accepted-versus-pending distinctions, reserved correction/egress,
scoped fresh checks, explicit commitment passages and attributable records.
It may also learn from the failures and omissions. It must independently
establish its own domain, dependencies, present conditions and proof limits.
These mechanisms do not become universal constitutional requirements.

Copying code is not adoption. Historical qualification is not present
qualification. Platform compatibility requires fresh verification.
Implementation does not create capability; capability does not create
capacity; capacity does not create Authority. A signature is not identity or
rightful effect. Record possession is not continuity; code ancestry is not
succession; deployment is not Presence. Possession or repository custody of
the archives creates no Source, identity or succession. The unchanged
canonical core, not this concrete stack, states its platform-neutral contract.

## 11. Inspection boundary

The two original ZIPs were inspected in memory, read-only; no source,
binaries, manifests, tests or evidence were unpacked into this repository.
Secret-material screening found no actual embedded private key or credential
material; that is a bounded screening result, not a universal safety claim.
No public chain or predecessor service was contacted, no AvalancheGo,
validator, predecessor binary or disposable network was run, no credentials
were used, no keys generated, and no release, deployment, tag or merge made.
Schema/vector checks evaluate representation only. Source and both archive
carriers retain their exact paths/bytes; all existing CSC/DCR files and both
first-pass inventories remain byte-identical.

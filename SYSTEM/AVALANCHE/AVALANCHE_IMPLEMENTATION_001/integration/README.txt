Coupled qualification
=====================

From AVALANCHE_IMPLEMENTATION_001:

  ./integration/run.sh

Optional public result capture (the path is resolved in vm/):

  COUPLED_EVIDENCE_PATH=../integration/qualification-run.json ./integration/run.sh

The runner uses only existing Go and de standard test/runtime tools. It starts
the real Go VM API on an ephemeral loopback HTTP listener, then executes the
real FIELD MediumEngine, VMClient, profiles, signature encoding and journal in
a separate Node process. No fake VM client supplies accepted state or receipts.
The VM test controller explicitly calls BuildBlock, Verify and Accept. Control
routes exist only in the integration-tagged test, not in the production API.

Every accepted operation is checked through:
  FIELD authorization -> draft -> external Ed25519 signature -> submission
  -> pending/no receipt/no state mutation -> real block verification/acceptance
  -> FIELD receipt observation -> renewed accepted-state observation.

The critical first pair is BOUND followed by DECLARE_CAPACITY with an empty
locus_id while the bounded locus remains active. Further coverage includes:
  - Wrong global scope and invalid signatures rejected by FIELD and the VM.
  - A stale prepared draft refused after another transition is accepted.
  - Presentation, gate admission, entry and controlled broker capability.
  - Crossing, local matter admission and local emergence remain distinct.
  - Signed payloads containing <, >, &, U+2028 and U+2029 traverse the real
    Go/JavaScript signing, transition-ID and canonical-submission boundary.
  - Capacity deficit suspends substantive work while correction/egress work.
  - Checkpoint, exit, closure and addressed residue without silent uptake.
  - Historical correction uses the target's old locus, not a new active locus.
  - Fresh presentation/admission and REENTER work in the same still-open locus
    despite its retained EXITED entry, then checkpoint and exit again.
  - After closure, a later fresh presentation/admission and REENTER consume
    that checkpoint without reopening the old locus; this lifecycle exits
    and closes in a new locus.
  - Dormant P0 pulses do not invent participants.
  - Two VM reinitializations from the same in-memory database preserve accepted
    state, transitions and receipts; old receipts cannot grant current presence.

Private signer keys are generated randomly in memory and supplied over stdin,
never persisted. Public journals are created under this directory and removed.
Go work files use a project-relative scratch directory removed by run.sh.

Read qualification-boundary.json before interpreting any PASS result. The
extended driver above tests explicit VM acceptance, not a validator quorum.
The separate network.sh/network-driver.mjs test is retained for three actual
local AvalancheGo validators and the custom plugin. It requires a node binary
whose Go build metadata matches PRESENCE_AVALANCHE_SECURITY_OVERLAY_001. No
historical run is transferred to the current contract digest. The bounded PR
#17 observation is preserved only in
records/HISTORICAL_PR17_NETWORK_OBSERVATION_001.json.

These results are implementation evidence, not constitutional effect,
Authority, formation, adoption, succession, external truth or Presence.

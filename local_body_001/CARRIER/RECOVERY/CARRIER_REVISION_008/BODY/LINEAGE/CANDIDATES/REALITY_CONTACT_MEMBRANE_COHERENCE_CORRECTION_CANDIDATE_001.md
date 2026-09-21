# Reality Contact Membrane Coherence Correction Candidate 001

Status: **CANDIDATE · CORRECTION ONLY**  
Candidate ID: `RM-REALITY-CONTACT-MEMBRANE-COHERENCE-CORRECTION-CANDIDATE-001`  
Body ID: `RM-LOCAL-REALITY-BODY-001`  
Expected currentness revision: `5`  
Proposed successor revision: `6`  
Applicable function: `CORRECTION_AUTHORITY`

## 1. Exact starting condition

Revision `5` admitted and activated `RM-REALITY-CONTACT-MEMBRANE-001`. The mechanism then performed real append-only observation. That operation did not amend the Body, admit Matter, create a relation, create a Contact Event, create a consequence, or mutate Body currentness.

The exact revision-`5` static carrier is preserved at `CARRIER/RECOVERY/CARRIER_REVISION_005/`, with manifest SHA-256 `25c7ad1b8e4c53cb3ccdff83abc9ba3f0242a5cfe527d2d1addaf1b8f5849014`. Its exact current state is also preserved at `LINEAGE/STATE_REVISIONS/CURRENT_REVISION_005.json`, SHA-256 `4d79749a3404f4705df3a9c37aec9f6088f23e5e3760bda4d13c584bd12e7543`.

The living state is not converted into static recovery. Its journal and raw carriers remain in the fixed mechanism-owned root and must survive the correction byte-for-byte.

## 2. Pressure that revealed representational failure

The first actual crossing exposed three exact defects in revision `5`:

1. `EMBODIMENT/STATE.json` recorded the Body entrypoint hash as `f32660448a8fb3eea1520313a62b5cd18ce5b320faa401ead8361080e8eb3dc8`, while the actual admitted entrypoint and static manifest carried `c3a3405a0dfb05395220e7248095e1864e99f6359228809ddfbc2c1d982bc5cb`.
2. The whole-Body verifier checked the entrypoint carrier but did not cross-check that represented hash, and its living-path allowlist named files the admitted mechanism never creates.
3. `CURRENT/STATE.json` and `EMBODIMENT/STATE.json` statically mirrored mutable live status and count as `AVAILABLE_EMPTY` and `0`. A valid crossing therefore made those static representations false even though the mechanism correctly refused to mutate Body currentness.

These are correction defects. They do not justify a new mechanism, constitutional change, adaptation, or retroactive semantic admission.

## 3. Selected correction

The smallest coherent correction is:

- correct the represented Body-entrypoint hash to the admitted bytes;
- require the whole-Body verifier to cross-check every represented membrane carrier and hash;
- make the mechanism-local append-only journal the sole owner of live entry count, journal tip, and empty/non-empty operational status;
- remove mutable live count and current live status from static Body-currentness projections;
- represent only the fixed route, mechanism identity, activation ceiling, verification requirement, and dynamic-owner relation in static currentness;
- accept only `local_journal.jsonl` and `raw_crossings/*.bin` as living-state paths, with every carrier exactly referenced by one valid `CROSSING_OBSERVED` journal entry and no unreferenced carriers;
- pressure-test an absent/zero state, the preserved current living state, and a disposable future append without requiring a static currentness rewrite; and
- preserve exact revision `5` and every actual crossing before replacing the carrier.

## 4. Living-state custody

`EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/` remains excluded from `CARRIER/MANIFEST.sha256` because it is living operational state. Exclusion from the static manifest is not exclusion from integrity verification.

The mechanism-local journal owns:

- the exact number of observed crossings;
- the current journal tip;
- whether the surface is empty or contains observations; and
- the raw-carrier references and byte hashes.

`CURRENT/STATE.json` and `EMBODIMENT/STATE.json` own only the durable fact that this named living surface exists at a fixed route, is written only by the admitted mechanism, is independently verified, and carries no semantic ownership beyond recipient-local observation.

No currentness revision is required when the journal later grows. A future correction or adaptation is required only if the static mechanism identity, route, ceiling, or ownership boundary changes.

## 5. Preserved living evidence

At the time this candidate was prepared, the Body-local journal verified with two entries. The correction must preserve at least these exact anchors:

- sequence `1`, entry SHA-256 `a7e16911aba51f1099ba9e746a38f4e433b6f28966d59f6b6f0c684677ebb88b`, raw carrier SHA-256 `765df8a6f7f898f52cf17c924c756afd71b5a281fd3c995f0145a346536fbbc6`;
- sequence `2`, entry SHA-256 `7d81684301e2790686edfbc894463299b99d0bdc4b217a9ae48efe5a04c169b5`, raw carrier SHA-256 `91ae6dfe68cce0764232618d618161ccaa2675b18a3f6043cc246c955dbfc36d`; and
- journal file SHA-256 after sequence `2`: `5c672dbab257f2e8630fb720fe60be7e4af9b830d23d11a93ff9a2c3f52e6035`.

If the live journal advances before substitution, the later valid state must also be preserved. These anchors are a minimum preservation floor, not a frozen maximum or a semantic classification of the crossings.

## 6. Verification and failure

Preflight must prove the staged correction while leaving the live Body untouched. Immediately before substitution, the executor must reread the live mechanism state. If it differs from the staged living state, the executor must either copy the newer valid state into the stage and reverify or HOLD. It must never overwrite a newer live journal.

Final verification must prove exact revision-`5` static recovery, exact preserved living bytes, correct dynamic ownership, current and future living-path validity, schema/currentness agreement, complete static manifest coverage, and all non-effects.

Any failure leaves or restores revision `5` plus its complete then-live mechanism-owned state. Failure creates no success occurrence and no currentness advance.

## 7. Explicit non-effects

This candidate does not authorize constitutional change; Body identity, Locus, formation, Authority, or relation change; mechanism redesign or new anatomy; Matter or epistemic admission; consequence creation; Contact Event, witness, participant, second Body, inter-Body relation, Relational Field, shared currentness, or settlement; outbound crossing; autonomous invocation; external publication; Recovery-corpus mutation; or relocation of any locality.

The second observed crossing carried an invitation. This candidate neither accepts nor refuses that invitation and creates no outbound capability. It corrects the Body's account of what its already-admitted inbound mechanism actually owns.

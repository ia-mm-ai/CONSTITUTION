# SYSTEM

## Standing

SYSTEM contains derivative machinery around [SOURCE](../SOURCE/CONSTITUTION_0%28%291.md)
and [STATE](../STATE/STATE.md) and is not another core module.
[SYSTEM.json](SYSTEM.json) records only the repository-integrity contract for
verification, export, and recovery. Separately bounded executable forms retain
their own identities, records, qualification ceilings, and non-effects.
Neither this file nor SYSTEM.json is an implementation claim under
[LINEAGE's existing schema](../STATE/LINEAGE/SCHEMAS/IMPLEMENTATION-CLAIM.schema.json).
The human Constitution governs meaning; machine references depend on its
exact-byte binding. Verification, export, recovery, publication, execution,
testing, schema acceptance, and parsing do not establish adoption, identity,
Agency, Authority, formation, Locality, settlement, or present capability.

## Current boundary

The repository contains the Source pair, STATE and its CSC/DCR/LINEAGE models,
23 schemas, 14 vector suites, the ORIGIN locator, two derivation records, and
the paired human/machine VM003–Medium001 historical implementation account.
The separate predecessor Site occurrence account retains a user-attributed
report, not a recovered service or a preserved screenshot.
Surface selects these files and both unchanged ZIP carriers. Invalid vectors
remain normative examples rather than reports of events; archive contents
remain historical evidence, not canonical or executable publication machinery.

[AVALANCHE_IMPLEMENTATION_001](AVALANCHE/AVALANCHE_IMPLEMENTATION_001/)
is a separately bounded source implementation with functional test evidence.
Its recorded status remains `SOURCE_IMPLEMENTED`, `FUNCTIONALLY_TESTED`,
`QUALIFICATION_INCOMPLETE`, `NOT_RELEASED`, `NOT_DEPLOYED`, and
`NOT_READY_FOR_MAINNET`. Its presence records no L1, Locality, settlement
occurrence, release, deployment, or public-network act.

[conformance.py](conformance.py) executes the formal evaluation contract in
`STATE/STATE.md`: offline Draft 2020-12 validation with format checking, exact
decimal comparisons, and ordered semantic rule evaluation against each vector's
expected triple. It also validates the models, rules, origin, derivation maps,
occurrence records, and implementation claims. Occurrences in
`STATE/LINEAGE/OCCURRENCES` receive both schema and LINEAGE semantic checks;
passing does not authenticate their evidence. Its expression engine and regression cases derive
from the former evaluator, now maintained solely in SYSTEM; no program in
canonical STATE is required or executed. Passing means representation
consistency at the declared scope, not present capability or constitutional effect.

The documents are operational contracts, not filled-in claims about a deployment.
No public address, release revision, generated digest, or conformance result
is invented as a placeholder.

## Maintenance operations

- **verify.py** reads one full, locally available Git commit, ignoring dirty
  working-tree files. It checks the Source pair, LINEAGE's Source bindings,
  declared resources, stable Source references, schema definitions, offline
  schema dependencies, and selection of model and implementation file relations.
  Both ZIPs must match their derivation digests and lengths and the account's
  exact carrier locators, and be recognizable ZIP files. No carrier contents
  are extracted or executed; nested `!/` locators retain their historical
  meaning, but their members are not independently re-inspected by this check.
  It runs the selected commit's SYSTEM evaluator
  in an isolated temporary tree, with a timeout, and records its byte binding.
  Select only trusted local Git
  revisions: this executes that revision's checker, not a sandboxed program.
  The report names the revision, checked resources, checks performed, and
  separate conformance result. Exit 0 means the requested check passed,
  1 means failure, and 2 means the selected revision lacks SYSTEM's evaluator.
- **export.py** constructs the edition from those exact Git blobs, never from
  a mixture of a named commit and current files. It writes only into a new
  destination whose parent exists, then checks the resulting edition.
- **recover.py** constructs the same publication from a selected revision, or
  copies a verified export into a new destination and verifies it again.
  Recovery of an export never executes code from the supplied directory.
  It recovers the selected exact carriers, not Git history, unselected evidence,
  or living State.
- **requirements.txt** pins the direct Python dependency actually imported by
  these operations and the interface. Python 3.10+ and Git must be installed
  separately; no web framework or remote schema fetch is required.

Use the entry points' `--help` for exact options. From the repository root:

```sh
python -m pip install -r SYSTEM/requirements.txt
COMMIT=$(git rev-parse --verify HEAD)
python SYSTEM/verify.py --revision "$COMMIT"
python SYSTEM/export.py --revision "$COMMIT" --destination ../presence-edition
# Use the exporter's printed digest, obtained through an independently trusted channel.
python SYSTEM/verify.py --export ../presence-edition --manifest-sha256 "$MANIFEST_SHA256"
python SYSTEM/recover.py --export ../presence-edition --manifest-sha256 "$MANIFEST_SHA256" --destination ../presence-recovered
```

`COMMIT` must be a full commit ID already available in the local repository and
must contain the declaration, entrance, and SYSTEM evaluator for a complete
edition. No default branch is silently selected or fetched. `--allow-incomplete`
is retained only for explicitly labelled previews of revisions lacking the
SYSTEM evaluator; it cannot bypass failed conformance, carrier bindings, or
missing selected relations.

For local evaluator development, `python SYSTEM/conformance.py` checks the
working tree, not a revision-bound edition. Run its regression tests with
`python -m unittest discover -s SYSTEM -p 'test_*.py'`. Revision verification
always ignores working-tree changes, including changes to the evaluator.

## Portable edition and trust

The publication contains `index.html`, generated `index.json`, generated
`manifest.json`, and the explicitly selected core files, implementation account,
and two historical ZIP carriers at their original
repository-relative paths. It contains no maintenance scripts, Git metadata,
unlisted working-tree files, or newly inferred evidence. The ZIPs are served
as `application/zip` exact bytes, never as executable interface code.
Relative model and account file references therefore retain their original
meaning and must resolve to selected material before export.

`index.json` binds resource paths, stable references, media types, operation
contracts, and the verification report to one revision. `manifest.json` binds
every other exported file by exact stored bytes (SHA-256 and byte length),
including the HTML and index. It does not attempt to hash itself. The exporter
prints its digest; distribute that digest through an independently trusted
channel. A changed file, extra file, missing file, symlink, unsafe path,
inconsistent index, or wrong supplied manifest digest fails export verification.

`verify.py --export ... --manifest-sha256 ...` checks those bytes and reports
`INTEGRITY_VERIFIED`, **not** a new core conformance run. Recorded conformance
remains explicitly labelled as recorded. A digest supplied by the same
untrusted party as an export does not authenticate its revision or provenance.
For independent correspondence to Git, reconstruct from the trusted local
commit and compare the resulting manifest digest using the same tool version.

The ORIGIN locator's `inspected_commit` describes its historical input snapshot;
it is never substituted for the commit selected for a new edition. Provenance
remains in [STATE/LINEAGE](../STATE/LINEAGE/LINEAGE.md), not a SYSTEM ledger.

## Publication boundary

The edition operations can produce and verify a portable edition. This
repository establishes no publication provider, deployment workflow, or public
location; domain control remains unverified. Publication is a later, separately
authorized operation and establishes no succession or constitutional coupling.

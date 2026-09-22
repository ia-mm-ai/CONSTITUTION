# SYSTEM

## Standing

SYSTEM is derivative maintenance around [SOURCE](../SOURCE/CONSTITUTION_0%28%291.md)
and [STATE](../STATE/STATE.md), not another core module or a runtime for CSC,
DCR, or LINEAGE. [SYSTEM.json](SYSTEM.json) records this initial operational
contract. Neither file is an implementation claim under
[LINEAGE's existing schema](../STATE/LINEAGE/SCHEMAS/IMPLEMENTATION-CLAIM.schema.json).
The human Constitution governs meaning; machine references depend on its
exact-byte binding. Publication, recovery, schema acceptance, and parsing do
not establish adoption, identity, Agency, Authority, or present capability.

## Current boundary

The repository contains the Source pair, STATE and its CSC/DCR/LINEAGE models,
23 schemas, 14 vector suites, the ORIGIN locator, and two derivation records.
The first surface selects these actual files, including invalid vectors as
normative examples rather than reports of events.

`STATE/STATE.md` describes `STATE/LINEAGE/CONFORMANCE.py`, but that executable
and its tests are absent at this initial externalisation. SYSTEM does **not**
replace its rule evaluator or treat schema well-formedness as vector success.
Revision verification reports `INCOMPLETE` when it cannot run that checker.
Ordinary export is blocked; `--allow-incomplete` deliberately makes a labelled
preview with the gap retained in both generated JSON files. An available
checker that fails cannot be bypassed by that option.

The documents are initial contracts, not filled-in claims about a deployment.
No public address, release revision, generated digest, or conformance result
is invented as a placeholder.

## Maintenance operations

- **verify.py** reads one full, locally available Git commit, ignoring dirty
  working-tree files. It checks the Source pair, LINEAGE's Source bindings,
  declared resources, stable Source references, schema definitions, and offline
  schema dependencies. When present, it runs the selected commit's core checker
  in an isolated temporary tree, with a timeout. Select only trusted local Git
  revisions: this executes that revision's checker, not a sandboxed program.
  The report names the revision, checked resources, checks performed, and
  separate conformance result. Exit 0 means the requested check passed,
  1 means failure, and 2 means revision conformance remains incomplete.
- **export.py** constructs the edition from those exact Git blobs, never from
  a mixture of a named commit and current files. It writes only into a new
  destination whose parent exists, then checks the resulting edition.
- **recover.py** constructs the same publication from a selected revision, or
  copies a verified export into a new destination and verifies it again.
  Recovery of an export never executes code from the supplied directory.
  It is not recovery of Git history, missing evidence archives, or living State.
- **requirements.txt** pins the direct Python dependency actually imported by
  these operations and the interface. Python 3.10+ and Git must be installed
  separately; no web framework or remote schema fetch is required.

Use the entry points' `--help` for exact options. From any directory:

```sh
python -m pip install -r /home/runner/work/PRESENCE/PRESENCE/SYSTEM/requirements.txt
python /home/runner/work/PRESENCE/PRESENCE/SYSTEM/verify.py --revision "$COMMIT"
python /home/runner/work/PRESENCE/PRESENCE/SYSTEM/export.py --revision "$COMMIT" --destination /tmp/presence-edition
python /home/runner/work/PRESENCE/PRESENCE/SYSTEM/recover.py --export /tmp/presence-edition --manifest-sha256 "$MANIFEST_SHA256" --destination /tmp/presence-recovered
```

`COMMIT` must be a full commit ID already available in the local repository and
must contain the declaration and entrance. No default branch is silently
selected or fetched. For the present checker gap, append `--allow-incomplete`
to an export or revision recovery only when an incomplete preview is intended.

## Portable edition and trust

The publication contains `index.html`, generated `index.json`, generated
`manifest.json`, and the explicitly selected core files at their original
repository-relative paths. It contains no maintenance scripts, Git metadata,
predecessor archives, unlisted working-tree files, or newly inferred evidence.
Relative core references therefore retain their original meaning.

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

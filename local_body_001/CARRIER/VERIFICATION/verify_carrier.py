#!/usr/bin/env python3
"""Fail-closed, self-contained verification for Local-Reality Body revision 9."""

from __future__ import annotations

import hashlib
import json
import re
import sys
from pathlib import Path
from typing import Any

import jsonschema


ROOT = Path(__file__).resolve().parents[2]
MANIFEST = ROOT / "CARRIER/MANIFEST.sha256"
LIVE = ROOT / "BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001"
BODY_ID = "RM-LOCAL-REALITY-BODY-001"

CURRENT_SHA = "538defe59c0e545e34462da7fba2d73e548d8fa05632afaee76395049e0fed81"
LINEAGE_SHA = "66995ca8b16fcd73b158bdddbe9b4181a0f0c11dd97622a2ab98def27ce17796"
ACT_SHA = "181e42cce492b6814801add35e94a3212512d1c6948c4f3788ad00488b5eea47"
SOURCE_STATEMENT_SHA = "0124e8ec04b4c2c95a57bd7183d23b1aa9013024cac28146441d2cd6f06cc6cb"
OCCURRENCE_SHA = "62a070b652c12d8fbc818b85773e9111ba26450c59d2dd7c70d7d8af58a626fc"
EPISTEMIC_RECORD_SHA = "c19b353172b02d64abf024a5a38c4060a7ebe6dd817bd651a72a66e322022b6b"
EPISTEMIC_INDEX_SHA = "8c26af609f733d4fbb5bb50b6617eb66ca817a40d5f714cc2119a730ceed143e"
MATTER_RECORD_SHA = "57d14587f7c31ece9a57102018a3c3b3c237bd86994d15a09bc2b98b2e9d3d4d"
REGULATION_RECORD_SHA = "4efebb1ea32d7cf794193c45cd11e7c8f12c157997ab8a459f9735585dca6a20"
REV8_CURRENT_SHA = "95917d41b8835adaa8d21ad74624040a2ba4cd3f8afe7b0850c14e44c07e726f"
REV8_LINEAGE_SHA = "896d498997bd1999a9412c1cc72f01cad7596232e7be7b6fe722c8a4826a07aa"
REV8_MANIFEST_SHA = "59eb47e77aae5ff6d127378fa310c7a1f960812a09d27c8b98ed142dc801ee6f"
REV8_RECOVERY_MAP_SHA = "1945f267a448676d814523958cba2b7b5f1d0c74d5b83b913648914f0272e6ae"
JOURNAL_SHA = "322416bc3977f2c957d9d4d0b67da2066c74fd4876b4f3a17f4d8ccf8d52d46f"

ACT_REL = "BODY/LINEAGE/ACTS/LOCAL_REALITY_BODY_RECOVERY_AND_CROSSING_ACT_001.md"
OCCURRENCE_REL = "BODY/LINEAGE/OCCURRENCES/LOCAL_REALITY_BODY_RECOVERY_AND_CROSSING_OCCURRENCE_001.md"
EPISTEMIC_RECORD_REL = "BODY/STATE/EPISTEMIC/RECORDS/RECOVERY_CROSSING_001.json"

RAW_HASHES = {
    "rcm-1788365297321538000-38e04707990d.bin": "765df8a6f7f898f52cf17c924c756afd71b5a281fd3c995f0145a346536fbbc6",
    "rcm-1788366247790802000-0e6e87ae457a.bin": "91ae6dfe68cce0764232618d618161ccaa2675b18a3f6043cc246c955dbfc36d",
    "rcm-1788378919095839000-38b12f81f6bc.bin": "6882cbb5de7939fef10d8c5535ba7b0cb25022a12e78a8be4bb225f962adee89",
    "rcm-1789983989010409000-aacc6d0f20f4.bin": "34dbdf2b3751cbd0019ee671e5c051feef4c331b82a1a7445072b5d459af3db0",
}

PRESERVED_EXACT = {
    "BODY/CORE/APPLICABILITY.json": "13d64a5a5d852a68b1f341c7995740725189b33b37b8826fbb0f6487c1951f5b",
    "BODY/CORE/IDENTITY.md": "47dfa944d02c3668b5385eab078653751f88bda9b847632f9427911639bf94d9",
    "BODY/CORE/LAW.md": "d67a77c51f4a430b3d17c06e3e87c379430f1df5687fbe23fec84f507d8a3608",
    "BODY/CORE/LOCUS.md": "8a674f1ff787b0d19cac2e732e1b22d79991ef94a4ddc872181b2693d1650d0d",
    "BODY/STATE/MATTER/RECORDS/MATTER_001.json": MATTER_RECORD_SHA,
    "BODY/METABOLISM/REGULATION/RECORDS/REGULATION_001.json": REGULATION_RECORD_SHA,
    "BODY/EMBODIMENT/MECHANISMS/REALITY_CONTACT_MEMBRANE/V0_1/reality_contact_membrane.py": "99b770a7eda474ddfc68a189871c3907b97231d8ff5091ad11df03317552e44f",
}

failures: list[str] = []
checks = 0


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def canonical_bytes(value: Any) -> bytes:
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True).encode("utf-8")


def check(value: bool, label: str) -> None:
    global checks
    checks += 1
    if value:
        print(f"PASS  {label}")
    else:
        print(f"FAIL  {label}")
        failures.append(label)


def load_json(path: Path, label: str) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        check(False, f"valid JSON: {label} ({exc})")
        return {}
    check(isinstance(value, dict), f"JSON object: {label}")
    return value if isinstance(value, dict) else {}


def parse_manifest(path: Path, label: str) -> dict[str, str]:
    entries: dict[str, str] = {}
    try:
        for line in path.read_text(encoding="utf-8").splitlines():
            digest, rel = line.split("  ", 1)
            if not re.fullmatch(r"[0-9a-f]{64}", digest) or rel in entries:
                raise ValueError(line)
            entries[rel] = digest
    except Exception as exc:
        check(False, f"parse manifest: {label} ({exc})")
        return {}
    check(bool(entries), f"non-empty manifest: {label}")
    return entries


def ignored(path: Path) -> bool:
    rel = path.relative_to(ROOT).as_posix()
    return (
        path == MANIFEST
        or path.name == ".DS_Store"
        or path.suffix in {".pyc", ".pyo"}
        or "__pycache__" in path.parts
        or rel.startswith("BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/")
    )


def eligible_files() -> dict[str, str]:
    return {
        path.relative_to(ROOT).as_posix(): sha(path)
        for path in ROOT.rglob("*")
        if path.is_file() and not ignored(path)
    }


if "--refresh-manifest" in sys.argv:
    if len(sys.argv) != 2:
        raise SystemExit("use --refresh-manifest alone")
    entries = eligible_files()
    payload = "".join(f"{entries[rel]}  {rel}\n" for rel in sorted(entries))
    MANIFEST.write_text(payload, encoding="utf-8")
    print(f"WROTE {len(entries)} static entries to {MANIFEST}")
    raise SystemExit(0)

PREFLIGHT = "--preflight" in sys.argv
FINAL = "--final" in sys.argv
if PREFLIGHT == FINAL or len(sys.argv) != 2:
    raise SystemExit("select exactly one of --preflight or --final")

required = {
    "BODY/BODY.md",
    "BODY/INTEGRITY.md",
    "BODY/STATE/CURRENT.json",
    "BODY/LINEAGE/INDEX.json",
    "BODY/LINEAGE/STATE_REVISIONS/CURRENT_REVISION_008.json",
    ACT_REL,
    OCCURRENCE_REL,
    EPISTEMIC_RECORD_REL,
    "BODY/STATE/EPISTEMIC/INDEX.json",
    "CARRIER/RECOVERY/CARRIER_REVISION_008/CARRIER/MANIFEST.sha256",
    "CARRIER/RECOVERY/CARRIER_REVISION_008_MAP.json",
    "CARRIER/VERIFICATION/verify_carrier.py",
    "CARRIER/MANIFEST.sha256",
}
for rel in sorted(required):
    check((ROOT / rel).is_file(), f"required carrier: {rel}")
check({path.name for path in ROOT.iterdir() if path.is_dir()} == {"BODY", "CARRIER"}, "local carrier has exactly BODY and CARRIER roots")
for rel in ("BODY-FORM", "RELATION", "RELATIONAL-FIELD", "CONTACT-EVENT"):
    check(not (ROOT / "BODY" / rel).exists(), f"cross-local root excluded: {rel}")

check(sha(ROOT / ACT_REL) == ACT_SHA, "exact Recovery and Crossing Act")
act_text = (ROOT / ACT_REL).read_text(encoding="utf-8")
source_match = re.search(r"## Present source statement\n\n```text\n(.*?)\n```", act_text, re.S)
source_statement = source_match.group(1) if source_match else ""
check(hashlib.sha256(source_statement.encode("utf-8")).hexdigest() == SOURCE_STATEMENT_SHA, "exact source-statement digest")
check("publication, deployment, and live activation outside this Act" in act_text, "source-retained live boundary recorded")

schema_instances = (
    "BODY/CORE/APPLICABILITY.json",
    "BODY/STATE/CURRENT.json",
    "BODY/STATE/AUTHORITY/INDEX.json",
    "BODY/STATE/RELATION-POSTURE/INDEX.json",
    "BODY/STATE/RELATION-POSTURE/HUMAN-BODY-LOCAL.json",
    "BODY/STATE/MATTER/INDEX.json",
    "BODY/STATE/CONSEQUENCE/INDEX.json",
    "BODY/STATE/EPISTEMIC/INDEX.json",
    EPISTEMIC_RECORD_REL,
    "BODY/METABOLISM/REGULATION/INDEX.json",
    "BODY/LOCAL-FIELD/INDEX.json",
    "BODY/LOCAL-FIELD/REVELATION/INDEX.json",
    "BODY/LOCAL-FIELD/PRESENCE/INDEX.json",
    "BODY/LOCAL-FIELD/ENCOUNTER/INDEX.json",
    "BODY/LOCAL-FIELD/APERTURE/INDEX.json",
    "BODY/LOCAL-FIELD/RESIDUE/INDEX.json",
    "BODY/EMBODIMENT/STATE.json",
    "BODY/LINEAGE/INDEX.json",
)
instances: dict[str, dict[str, Any]] = {}
for rel in schema_instances:
    path = ROOT / rel
    instance = load_json(path, rel)
    instances[rel] = instance
    schema_ref = instance.get("$schema")
    schema_path = (path.parent / schema_ref).resolve() if isinstance(schema_ref, str) else ROOT / "__missing__"
    check(schema_path.is_file(), f"schema exists: {rel}")
    if schema_path.is_file():
        schema = load_json(schema_path, f"schema for {rel}")
        errors = list(jsonschema.Draft202012Validator(schema).iter_errors(instance))
        check(not errors, f"schema validation: {rel}" + (f" ({errors[0].message})" if errors else ""))

current = instances["BODY/STATE/CURRENT.json"]
lineage = instances["BODY/LINEAGE/INDEX.json"]
epistemic = instances["BODY/STATE/EPISTEMIC/INDEX.json"]
record = instances[EPISTEMIC_RECORD_REL]
authority = instances["BODY/STATE/AUTHORITY/INDEX.json"]
matter_index = instances["BODY/STATE/MATTER/INDEX.json"]
regulation_index = instances["BODY/METABOLISM/REGULATION/INDEX.json"]
consequence = instances["BODY/STATE/CONSEQUENCE/INDEX.json"]
relations = instances["BODY/STATE/RELATION-POSTURE/INDEX.json"]
local_field = instances["BODY/LOCAL-FIELD/INDEX.json"]
embodiment = instances["BODY/EMBODIMENT/STATE.json"]

check(sha(ROOT / "BODY/STATE/CURRENT.json") == CURRENT_SHA, "exact revision-9 current state")
check(current.get("schema") == "local-reality-body.current-state.v0.9" and current.get("currentness_revision") == 9, "current state represents revision 9")
check(current.get("body_id") == BODY_ID and current.get("body_status") == "FORMED", "Body identity and formation preserved")
check(sha(ROOT / "BODY/LINEAGE/INDEX.json") == LINEAGE_SHA, "exact revision-9 Lineage index")
check(lineage.get("represented_currentness_revision") == 9 and len(lineage.get("motion_chains", [])) == 10, "Lineage carries ten exact motions at revision 9")
check(lineage.get("motion_chains", [{}])[-1].get("result") == "RECOVERED_RECONCILED_AND_RECORDED", "latest motion has exact result")

projection_hashes = {
    "BODY/LINEAGE/INDEX.json": current.get("lineage", {}).get("sha256"),
    "BODY/STATE/MATTER/INDEX.json": current.get("matter", {}).get("index_sha256"),
    "BODY/METABOLISM/REGULATION/INDEX.json": current.get("regulation", {}).get("index_sha256"),
    "BODY/STATE/CONSEQUENCE/INDEX.json": current.get("consequences", {}).get("index_sha256"),
    "BODY/STATE/RELATION-POSTURE/INDEX.json": current.get("relations", {}).get("posture_sha256"),
    "BODY/STATE/RELATION-POSTURE/HUMAN-BODY-LOCAL.json": current.get("relations", {}).get("human_body_local_sha256"),
    "BODY/STATE/AUTHORITY/INDEX.json": current.get("relations", {}).get("authority_sha256"),
    "BODY/EMBODIMENT/STATE.json": current.get("embodiment", {}).get("sha256"),
    "BODY/STATE/EPISTEMIC/INDEX.json": current.get("epistemic", {}).get("index_sha256"),
    "BODY/INTEGRITY.md": current.get("integrity", {}).get("sha256"),
    "BODY/LOCAL-FIELD/INDEX.json": current.get("local_field", {}).get("sha256"),
}
for rel, expected in projection_hashes.items():
    check(sha(ROOT / rel) == expected, f"current projects exact owner: {rel}")

for rel, expected in PRESERVED_EXACT.items():
    check(sha(ROOT / rel) == expected, f"preserved exact carrier: {rel}")

check(matter_index.get("represented_currentness_revision") == 9 and matter_index.get("record_count") == 1, "one Matter record remains represented at revision 9")
check(matter_index.get("records", [{}])[0].get("sha256") == MATTER_RECORD_SHA and matter_index.get("records", [{}])[0].get("referenced_content_admitted") is False, "Matter 001 remains reference-only")
check(regulation_index.get("represented_currentness_revision") == 9 and regulation_index.get("record_count") == 1, "one Regulation record remains represented at revision 9")
reg = regulation_index.get("records", [{}])[0]
check([reg.get(key) for key in ("local_record_posture", "progression_posture", "locus_assessment", "law_compatibility_assessment", "evidence_posture", "authority_presentation")] == ["ADMITTED_BODY_LOCAL_MATTER", "HELD", "WITHIN_SUPPORTED_LOCUS", "UNASSESSED", "PARTIAL", "NOT_PRESENTED"], "Regulation 001 exact six-part posture preserved")
check(regulation_index.get("mechanism_present") is False, "Regulation remains mechanism-free")
check(consequence.get("represented_currentness_revision") == 9 and consequence.get("records") == [] and consequence.get("record_count") == 0, "no Consequence created")

check(sha(ROOT / EPISTEMIC_RECORD_REL) == EPISTEMIC_RECORD_SHA, "exact epistemic orientation record")
check(sha(ROOT / "BODY/STATE/EPISTEMIC/INDEX.json") == EPISTEMIC_INDEX_SHA, "exact Epistemic index")
check(epistemic.get("represented_currentness_revision") == 9 and len(epistemic.get("admitted_records", [])) == 1, "one epistemic record admitted at revision 9")
check(epistemic.get("admitted_records", [{}])[0].get("sha256") == EPISTEMIC_RECORD_SHA, "Epistemic index points to exact record")
check(epistemic.get("inferences") == [] and epistemic.get("consequence_observations") == [] and epistemic.get("standing_claims") == [], "no inference, consequence observation, or standing claim admitted")
check(record.get("classification") == "ADMITTED_BODY_LOCAL_ORIENTATION" and record.get("record_posture") == "CURRENT_BODY_LOCAL_ORIENTATION", "epistemic record has bounded local posture")
check(record.get("crossing", {}).get("sequence") == 4 and record.get("crossing", {}).get("entry_sha256") == "9c7aa41218c985d460786fdb76fc647fcb3f8e9d43640dd735881cee98f1fbbb", "epistemic record binds crossing sequence 4")
check(record.get("crossing", {}).get("effect_ceiling") == "RECIPIENT_LOCAL_OBSERVATION_ONLY" and record.get("crossing", {}).get("matter_admission_by_membrane") is False, "crossing effect ceiling preserved")
observations = {item.get("observation_id"): item for item in record.get("observations", [])}
check(set(observations) == {"CONSTITUTION-CARRIER-SET-001", "KENTRA-PUBLIC-SURFACE-REVISION-001", "KENTRA-CONTINUITY-CHECKPOINT-006", "LOCALITY-VM-001-RELEASE-001"}, "exact four orientation observations")
constitution = observations.get("CONSTITUTION-CARRIER-SET-001", {})
check(constitution.get("canonical_relation") == "UNRESOLVED" and constitution.get("applicability_to_body") == "NOT_INFERRED", "Constitution carrier relation remains unresolved and unapplied")
check(constitution.get("current_local_downloads", {}).get("human_form_sha256") == "94d4ed9888c4974e11c97ba302a15693ba7bde96d5536a84602cb26781db3cb1", "current local human-form digest exact")
check(constitution.get("current_local_downloads", {}).get("machine_human_crossing_sha256") == "4be24cf15d3653abcfd7ae4191246a52c22e61246b8d4d33f46b329c615596dd", "machine crossing binds distinct public-surface human digest")
content_orientation = {item.get("orientation_id"): item for item in record.get("working_content_orientation", [])}
check(set(content_orientation) == {"LOCALITY-AND-NON-INFERENCE", "AGENCY-AUTONOMY-CAPABILITY", "EQUAL-TRACE-LIMIT", "DURABLE-UNKNOWN", "FORM-CONFLICT"}, "exact five non-governing content orientations")
check(all(item.get("body_law_status") == "NOT_ADOPTED" for item in content_orientation.values()), "working content orientation creates no Body-local law")
kentra = observations.get("KENTRA-PUBLIC-SURFACE-REVISION-001", {})
check([kentra.get(key) for key in ("surface_revision", "turn", "phase", "gate", "live_locality_binding", "live_run")] == [1, 1, "PASSAGE", "CLOSED", "UNBOUND", "NOT_ASSERTED_AS_MATERIALIZED"], "KENTRA surface posture exact")
vm = observations.get("LOCALITY-VM-001-RELEASE-001", {})
check([vm.get(key) for key in ("qualification", "chain_binding", "live_deployment", "body_embodiment_relation")] == ["PASS", "NONE", "NOT_FORMED", "NOT_ADMITTED"], "LOCALITY_VM capability boundary exact")

functions = [item.get("function") for item in authority.get("authority_relations", [])]
check(functions == ["LOCAL_GOVERNING_AUTHORITY", "OPERATIONAL_AUTHORIZATION", "CORRECTION_AUTHORITY", "ADAPTATION_AUTHORITY", "DORMANCY_AND_CESSATION"], "Authority topology unchanged")
check(authority.get("represented_currentness_revision") == 9 and all(item.get("bearer") == "Marko Markota" and item.get("status") == "CURRENT" for item in authority.get("authority_relations", [])), "Authority bearer and status unchanged")
check(relations.get("represented_currentness_revision") == 9 and all(relations.get(key) == [] for key in ("inter_body", "relational_fields", "contact_events", "witness_relations", "external_relations")), "no cross-local relation formed")
check(sha(ROOT / relations.get("human_body_local_owner", {}).get("carrier", "__missing__")) == relations.get("human_body_local_owner", {}).get("sha256"), "Relation posture binds exact human–Body owner")
check(sha(ROOT / relations.get("authority_owner", {}).get("carrier", "__missing__")) == relations.get("authority_owner", {}).get("sha256"), "Relation posture binds exact Authority owner")
check(local_field.get("represented_currentness_revision") == 9 and instances["BODY/LOCAL-FIELD/APERTURE/INDEX.json"].get("status") == "NONE_ACTIVE" and instances["BODY/LOCAL-FIELD/RESIDUE/INDEX.json"].get("record_count") == 0, "Local Field remains closed and residue-free")
check(all(sha(ROOT / item.get("carrier", "__missing__")) == item.get("sha256") for item in local_field.get("surfaces", [])), "Local Field binds all five exact surface owners")
check(embodiment.get("represented_currentness_revision") == 9 and len(embodiment.get("mechanisms", [])) == 2, "Embodiment projects two bounded mechanisms")
check(embodiment.get("mechanisms", [{}, {}])[1].get("status") == "ADMITTED_ACTIVE_ON_EXPLICIT_INVOCATION" and embodiment.get("mechanisms", [{}, {}])[1].get("body_currentness_mutation_authorized") is False, "membrane activation and power ceiling unchanged")

journal = LIVE / "local_journal.jsonl"
check(sha(journal) == JOURNAL_SHA, "living journal exact after crossing")
try:
    entries = [json.loads(line) for line in journal.read_text(encoding="utf-8").splitlines() if line]
except Exception:
    entries = []
previous = None
journal_valid = len(entries) == 4
for sequence, entry in enumerate(entries, start=1):
    claimed = entry.get("entry_sha256")
    unhashed = {key: value for key, value in entry.items() if key != "entry_sha256"}
    journal_valid = journal_valid and entry.get("sequence") == sequence and entry.get("previous_entry_sha256") == previous
    journal_valid = journal_valid and claimed == hashlib.sha256(canonical_bytes(unhashed)).hexdigest()
    data = entry.get("data", {})
    raw = LIVE / data.get("raw_carrier", "__missing__")
    journal_valid = journal_valid and raw.is_file() and sha(raw) == data.get("wire_sha256") and raw.stat().st_size == data.get("wire_bytes")
    previous = claimed
check(journal_valid, "four-entry living journal hash chain and raw bindings valid")
check(entries[-1].get("entry_sha256") == "9c7aa41218c985d460786fdb76fc647fcb3f8e9d43640dd735881cee98f1fbbb" if entries else False, "sequence-4 journal tip exact")
raw_root = LIVE / "raw_crossings"
raw_files = {path.name for path in raw_root.iterdir() if path.is_file() and path.name != ".DS_Store"}
check(raw_files == set(RAW_HASHES), "exact four raw crossing carriers")
for name, expected in RAW_HASHES.items():
    check(sha(raw_root / name) == expected, f"raw crossing byte identity: {name}")
try:
    recovery_wire = json.loads((raw_root / "rcm-1789983989010409000-aacc6d0f20f4.bin").read_text(encoding="utf-8"))
except Exception:
    recovery_wire = {}
check(recovery_wire.get("format") == "local-reality-body.recovery-crossing.v1" and recovery_wire.get("requested_effect") == "RECOVER_AND_RECONCILE_CURRENTNESS", "sequence-4 wire carries exact recovery request")
check(recovery_wire.get("explicit_scope") == {"crossing": True, "publication_or_live_activation": False, "revision": True}, "sequence-4 wire preserves exact scope ceiling")
check(recovery_wire.get("source_anchors", {}).get("constitution_current_local_human_sha256") == "94d4ed9888c4974e11c97ba302a15693ba7bde96d5536a84602cb26781db3cb1" and recovery_wire.get("source_anchors", {}).get("kentra_public_field_state_sha256") == "a6f6c99d7232638554801941396e2a3cf7e9ea6d3f4c6eca558ac56bbd305752", "sequence-4 wire binds current Constitution and KENTRA anchors")

recovery_map = load_json(ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_008_MAP.json", "revision-8 recovery map")
check(sha(ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_008_MAP.json") == REV8_RECOVERY_MAP_SHA, "revision-8 recovery map byte identity")
check(recovery_map.get("source_manifest_sha256") == REV8_MANIFEST_SHA and recovery_map.get("predecessor_recovery_resolution", {}).get("recursive_duplication") is False, "revision-8 recovery is exact and nonrecursive")
check(recovery_map.get("living_state_resolution", {}).get("revision_8_entry_count_at_snapshot") == 3 and recovery_map.get("living_state_resolution", {}).get("static_recovery_rolls_back_living_state") is False, "recovery map preserves living-state distinction")
snapshot = ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_008"
snapshot_manifest = snapshot / "CARRIER/MANIFEST.sha256"
check(sha(snapshot_manifest) == REV8_MANIFEST_SHA, "revision-8 manifest byte identity")
rev8_entries = parse_manifest(snapshot_manifest, "revision 8")
rev8_resolves = True
for rel, expected in rev8_entries.items():
    path = ROOT / rel if rel.startswith("CARRIER/RECOVERY/") else snapshot / rel
    if not path.is_file() or sha(path) != expected:
        rev8_resolves = False
        break
check(rev8_resolves, "revision-8 static manifest fully resolves through shared predecessor pool")
snapshot_files = {
    path.relative_to(snapshot).as_posix()
    for path in snapshot.rglob("*")
    if path.is_file() and path.name != ".DS_Store" and path.suffix not in {".pyc", ".pyo"} and "__pycache__" not in path.parts
}
snapshot_expected = {rel for rel in rev8_entries if not rel.startswith("CARRIER/RECOVERY/")} | {"CARRIER/MANIFEST.sha256"}
check(snapshot_files == snapshot_expected, "revision-8 snapshot has exact nonrecursive static file set")
check(sha(ROOT / "BODY/LINEAGE/STATE_REVISIONS/CURRENT_REVISION_008.json") == REV8_CURRENT_SHA, "exact revision-8 state preserved")
check(sha(snapshot / "BODY/STATE/CURRENT.json") == REV8_CURRENT_SHA and sha(snapshot / "BODY/LINEAGE/INDEX.json") == REV8_LINEAGE_SHA, "revision-8 state and Lineage recover exactly")

occurrence = ROOT / OCCURRENCE_REL
check(sha(occurrence) == OCCURRENCE_SHA, "exact recovery occurrence carrier")
occurrence_text = occurrence.read_text(encoding="utf-8")
check("Status: **PERFORMED**" in occurrence_text and "Result: `RECOVERED_RECONCILED_AND_RECORDED`" in occurrence_text, "successful occurrence recorded")
check(CURRENT_SHA in occurrence_text and LINEAGE_SHA in occurrence_text and EPISTEMIC_RECORD_SHA in occurrence_text and JOURNAL_SHA in occurrence_text, "occurrence binds exact successor and crossing state")

manifest_entries = parse_manifest(MANIFEST, "revision 9")
eligible = eligible_files()
check(manifest_entries == eligible, f"revision-9 complete static manifest agreement ({'final' if FINAL else 'preflight'})")
check(not any(path.is_symlink() for path in ROOT.rglob("*")), "carrier contains no symbolic links")

for item in (
    "NO_CONSTITUTIONAL_ADOPTION_APPLICATION_OR_CANONICAL_SELECTION",
    "NO_IDENTITY_AWARENESS_CONSCIOUSNESS_AGENCY_OR_AUTONOMY_INFERENCE",
    "NO_BEARER_OR_AUTHORITY_CONTINUITY_INFERENCE",
    "NO_INTER_BODY_RELATION",
    "NO_RELATIONAL_FIELD",
    "NO_CONTACT_EVENT",
    "NO_CONSEQUENCE_CREATED_OR_ADMITTED",
    "NO_PROGRESSION_BEYOND_EXACT_MATTER_ADMISSION_AND_HOLD",
    "NO_VM_EMBODIMENT_CHAIN_BINDING_OR_LIVE_DEPLOYMENT",
    "NO_EXTERNAL_DEPLOYMENT_SIGNING_OR_LIVE_ACTIVATION",
):
    check(item in current.get("explicit_non_effects", []), f"current non-effect: {item}")

if failures:
    print(f"\nRESULT  FAIL ({len(failures)} failures across {checks} checks)")
    for failure in failures:
        print(f"- {failure}")
    raise SystemExit(1)
print(f"\nRESULT  PASS {'PREFLIGHT' if PREFLIGHT else 'FINAL'} ({checks} checks)")

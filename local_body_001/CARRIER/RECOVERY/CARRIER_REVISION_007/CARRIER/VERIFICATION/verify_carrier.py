#!/usr/bin/env python3
"""Fail-closed verification for Local-Reality Body revision 7."""

from __future__ import annotations

import hashlib
import importlib.util
import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

import jsonschema


ROOT = Path(__file__).resolve().parents[2]
EXTERNAL = Path("/Users/markomarkota/LOCAL_REALITY_BODY_FORMS/RM_LOCAL_REALITY_BODY_FORM_001")
PREFLIGHT = "--preflight" in sys.argv
FINAL = "--final" in sys.argv
if PREFLIGHT == FINAL:
    raise SystemExit("select exactly one of --preflight or --final")

BODY_ID = "RM-LOCAL-REALITY-BODY-001"
CANDIDATE_SHA = "d14d5002e80d3218f2b1ef696e1058048ba0e7ba07e85be70768536769ce317e"
REV6_CURRENT_SHA = "7fc906ac68fdf350d6a91cfed1119902f8635c8308af9d693a0f6e491c354038"
REV6_LINEAGE_SHA = "0521785f404bda73ba2675fef331cf9755c8be588248d60a0e2e802848af2d49"
REV6_EMBODIMENT_SHA = "b4d662000308cde1c7b44e4efbea4371f21b396b0a1117f177bac5e0ede71d32"
REV6_MANIFEST_SHA = "0131fbf90f4860458910e0baaff47aeb8a5fa72059dcc4eb4e3ffd6dbf952911"
SOURCE_STATEMENT = "ADAPT_AND_REBIND"
SOURCE_STATEMENT_SHA = hashlib.sha256(SOURCE_STATEMENT.encode("utf-8")).hexdigest()
BODY_FORM_SHA = "4031ba2a6b7bc7cbe7352bdd91b351227aa1d517a0d4680d4fafc6afc26ca056"
EXTERNAL_MANIFEST_SHA = "da93daeba6b0f8e2c49521f4b043be35330d103422dae1ddade82670f43f5a2c"
JOURNAL_SHA = "77c1cb773a3ec48af88b4b8228fa54188587e749d304307b0152e6f91f5164a8"
RAW_HASHES = {
    "rcm-1788365297321538000-38e04707990d.bin": "765df8a6f7f898f52cf17c924c756afd71b5a281fd3c995f0145a346536fbbc6",
    "rcm-1788366247790802000-0e6e87ae457a.bin": "91ae6dfe68cce0764232618d618161ccaa2675b18a3f6043cc246c955dbfc36d",
    "rcm-1788378919095839000-38b12f81f6bc.bin": "6882cbb5de7939fef10d8c5535ba7b0cb25022a12e78a8be4bb225f962adee89",
}

failures: list[str] = []
checks = 0


def check(condition: bool, label: str) -> None:
    global checks
    checks += 1
    if condition:
        print(f"PASS  {label}")
    else:
        print(f"FAIL  {label}")
        failures.append(label)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_json(path: Path, label: str) -> dict:
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
            digest, relative = line.split("  ", 1)
            if len(digest) != 64 or relative in entries:
                raise ValueError(line)
            entries[relative] = digest
    except Exception as exc:
        check(False, f"parse manifest: {label} ({exc})")
        return {}
    check(bool(entries), f"non-empty manifest: {label}")
    return entries


def referenced_hash(record: dict, label: str) -> None:
    carrier = record.get("carrier")
    expected = record.get("sha256")
    path = ROOT / carrier if isinstance(carrier, str) else ROOT / "__missing__"
    check(path.is_file(), f"reference exists: {label}")
    check(path.is_file() and isinstance(expected, str) and sha(path) == expected, f"reference hash: {label}")


def quoted_source(path: Path) -> str:
    lines = path.read_text(encoding="utf-8").splitlines()
    heading = "## Exact source statement"
    if heading not in lines:
        return ""
    position = lines.index(heading) + 1
    while position < len(lines) and not lines[position].startswith("> "):
        position += 1
    values: list[str] = []
    while position < len(lines) and lines[position].startswith("> "):
        values.append(lines[position][2:])
        position += 1
    return "\n".join(values)


required = {
    "BODY/BODY.md",
    "BODY/INTEGRITY.md",
    "BODY/CORE/IDENTITY.md",
    "BODY/CORE/LAW.md",
    "BODY/CORE/LOCUS.md",
    "BODY/CORE/APPLICABILITY.json",
    "BODY/STATE/CURRENT.json",
    "BODY/STATE/AUTHORITY/INDEX.json",
    "BODY/STATE/RELATION-POSTURE/INDEX.json",
    "BODY/STATE/RELATION-POSTURE/HUMAN-BODY-LOCAL.json",
    "BODY/STATE/MATTER/BOUNDARY.md",
    "BODY/STATE/MATTER/INDEX.json",
    "BODY/STATE/CONSEQUENCE/BOUNDARY.md",
    "BODY/STATE/CONSEQUENCE/INDEX.json",
    "BODY/STATE/EPISTEMIC/BOUNDARY.md",
    "BODY/STATE/EPISTEMIC/INDEX.json",
    "BODY/METABOLISM/REGULATION/MEDIUM.md",
    "BODY/METABOLISM/REGULATION/INDEX.json",
    "BODY/LOCAL-FIELD/INDEX.json",
    "BODY/LOCAL-FIELD/REVELATION/INDEX.json",
    "BODY/LOCAL-FIELD/PRESENCE/INDEX.json",
    "BODY/LOCAL-FIELD/ENCOUNTER/INDEX.json",
    "BODY/LOCAL-FIELD/APERTURE/INDEX.json",
    "BODY/LOCAL-FIELD/RESIDUE/INDEX.json",
    "BODY/EMBODIMENT/STATE.json",
    "BODY/LINEAGE/INDEX.json",
    "BODY/LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_SEMANTIC_ANATOMY_ADAPTATION_CANDIDATE_001.md",
    "BODY/LINEAGE/ACTS/LOCAL_REALITY_BODY_SEMANTIC_ANATOMY_ADAPTATION_ACT_001.md",
    "BODY/LINEAGE/STATE_REVISIONS/CURRENT_REVISION_006.json",
    "CARRIER/RECOVERY/CARRIER_REVISION_006/CARRIER/MANIFEST.sha256",
    "CARRIER/RECOVERY/CARRIER_REVISION_006_MAP.json",
    "CARRIER/VERIFICATION/verify_carrier.py",
    "CARRIER/MANIFEST.sha256",
}
occurrence_rel = "BODY/LINEAGE/OCCURRENCES/LOCAL_REALITY_BODY_SEMANTIC_ANATOMY_ADAPTATION_OCCURRENCE_001.md"
if FINAL:
    required.add(occurrence_rel)
for relative in sorted(required):
    check((ROOT / relative).is_file(), f"required carrier: {relative}")

root_dirs = {path.name for path in ROOT.iterdir() if path.is_dir()}
check(root_dirs == {"BODY", "CARRIER"}, "private carrier has exactly BODY and CARRIER roots")
for retired in ("BODY.md", "CORE", "CURRENT", "MATTER", "REGULATION", "CONSEQUENCES", "RELATIONS", "EMBODIMENT", "EPISTEMIC", "LINEAGE"):
    check(not (ROOT / retired).exists(), f"retired current route absent: {retired}")
for cross_local in ("BODY-FORM", "RELATION", "RELATIONAL-FIELD", "CONTACT-EVENT"):
    check(not (ROOT / "BODY" / cross_local).exists(), f"cross-local root excluded from private Body: {cross_local}")

candidate = ROOT / "BODY/LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_SEMANTIC_ANATOMY_ADAPTATION_CANDIDATE_001.md"
act = ROOT / "BODY/LINEAGE/ACTS/LOCAL_REALITY_BODY_SEMANTIC_ANATOMY_ADAPTATION_ACT_001.md"
check(sha(candidate) == CANDIDATE_SHA, "exact selected candidate")
statement = quoted_source(act)
check(statement == SOURCE_STATEMENT, "exact selecting source statement")
check(hashlib.sha256(statement.encode("utf-8")).hexdigest() == SOURCE_STATEMENT_SHA, "selecting source statement hash")
act_text = act.read_text(encoding="utf-8")
check("Posture: `ADAPT_AND_REBIND`" in act_text, "Act records selected posture")
check("`ADAPTATION_AUTHORITY`, `OPERATIONAL_AUTHORIZATION`" in act_text, "Act invokes distinct required functions")
check("Bearer of both functions: `Marko Markota`" in act_text, "Act identifies Authority bearer")

schema_instances = (
    "BODY/CORE/APPLICABILITY.json",
    "BODY/STATE/CURRENT.json",
    "BODY/STATE/AUTHORITY/INDEX.json",
    "BODY/STATE/RELATION-POSTURE/INDEX.json",
    "BODY/STATE/RELATION-POSTURE/HUMAN-BODY-LOCAL.json",
    "BODY/STATE/MATTER/INDEX.json",
    "BODY/STATE/CONSEQUENCE/INDEX.json",
    "BODY/STATE/EPISTEMIC/INDEX.json",
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
instances: dict[str, dict] = {}
for relative in schema_instances:
    path = ROOT / relative
    instance = load_json(path, relative)
    instances[relative] = instance
    schema_ref = instance.get("$schema")
    schema_path = (path.parent / schema_ref).resolve() if isinstance(schema_ref, str) else ROOT / "__missing__"
    check(schema_path.is_file(), f"schema exists: {relative}")
    try:
        jsonschema.Draft202012Validator(load_json(schema_path, f"schema for {relative}")).validate(instance)
    except Exception as exc:
        check(False, f"schema validation: {relative} ({exc})")
    else:
        check(True, f"schema validation: {relative}")

current = instances["BODY/STATE/CURRENT.json"]
check(current.get("schema") == "local-reality-body.current-state.v0.7", "current schema identity")
check(current.get("body_id") == BODY_ID and current.get("body_status") == "FORMED", "Body identity and formation unchanged")
check(current.get("currentness_revision") == 7, "currentness revision 7")
check(current.get("formation", {}).get("result") == "FORMED" and current.get("formation", {}).get("source_capacity_after") == "EXHAUSTED", "formation and source exhaustion unchanged")
check(current.get("core", {}).get("law", {}).get("same_body_constitutional_change_authority") is False, "no same-Body constitutional-change Authority")
for key in ("identity", "law", "locus", "applicability"):
    referenced_hash(current.get("core", {}).get(key, {}), f"current Core {key}")
referenced_hash(current.get("lineage", {}), "current Lineage")
referenced_hash(current.get("integrity", {}), "current Integrity")
referenced_hash(current.get("local_field", {}), "current Local Field")
referenced_hash(current.get("embodiment", {}), "current Embodiment")
check(current.get("integrity", {}).get("new_law_created") is False, "Integrity projection creates no new law")

authority = instances["BODY/STATE/AUTHORITY/INDEX.json"]
functions = [entry.get("function") for entry in authority.get("authority_relations", [])]
check(functions == ["LOCAL_GOVERNING_AUTHORITY", "OPERATIONAL_AUTHORIZATION", "CORRECTION_AUTHORITY", "ADAPTATION_AUTHORITY", "DORMANCY_AND_CESSATION"], "five Authority functions preserved exactly")
check(all(entry.get("bearer") == "Marko Markota" and entry.get("status") == "CURRENT" for entry in authority.get("authority_relations", [])), "Authority bearer and status unchanged")
check(authority.get("authority_topology_changed") is False, "Authority topology unchanged")
human = instances["BODY/STATE/RELATION-POSTURE/HUMAN-BODY-LOCAL.json"]
check(human.get("relation", {}).get("status") == "FORMED" and human.get("relation", {}).get("bearer") == "Marko Markota", "human Body-local relation preserved")
posture = instances["BODY/STATE/RELATION-POSTURE/INDEX.json"]
referenced_hash(posture.get("human_body_local_owner", {}), "human Body-local relation owner")
referenced_hash(posture.get("authority_owner", {}), "Authority owner")
check(len(posture.get("di_coagency", [])) == 1, "bounded DI co-agency posture preserved")
for key in ("inter_body", "relational_fields", "contact_events", "witness_relations", "external_relations"):
    check(posture.get(key) == [], f"no cross-local relation created: {key}")

matter = instances["BODY/STATE/MATTER/INDEX.json"]
consequence = instances["BODY/STATE/CONSEQUENCE/INDEX.json"]
epistemic = instances["BODY/STATE/EPISTEMIC/INDEX.json"]
regulation = instances["BODY/METABOLISM/REGULATION/INDEX.json"]
check(matter.get("records") == [] and matter.get("record_count") == 0, "Matter remains empty")
check(consequence.get("records") == [] and consequence.get("record_count") == 0, "Consequence remains empty")
check(regulation.get("records") == [] and regulation.get("record_count") == 0 and regulation.get("mechanism_present") is False, "Regulation remains empty and mechanism-free")
for key in ("admitted_records", "source_relations", "observations", "evidence", "inferences", "working_interpretations", "consequence_observations", "standing_claims"):
    check(epistemic.get(key) == [], f"Epistemic remains empty and unsplit: {key}")

local_field = instances["BODY/LOCAL-FIELD/INDEX.json"]
check([entry.get("name") for entry in local_field.get("surfaces", [])] == ["REVELATION", "PRESENCE", "ENCOUNTER", "APERTURE", "RESIDUE"], "exact supported Local Field surfaces")
for entry in local_field.get("surfaces", []):
    referenced_hash(entry, f"Local Field {entry.get('name')}")
check(local_field.get("core_exposed") is False and local_field.get("cross_local_container") is False, "Local Field neither exposes Core nor contains cross-local reality")
revelation = instances["BODY/LOCAL-FIELD/REVELATION/INDEX.json"]
referenced_hash(revelation.get("occurrence", {}), "existing revelation occurrence")
check(revelation.get("new_revelation_created") is False, "adaptation creates no new revelation")
presence = instances["BODY/LOCAL-FIELD/PRESENCE/INDEX.json"]
check(presence.get("external_carrier") == str(EXTERNAL) and presence.get("body_form_sha256") == BODY_FORM_SHA and presence.get("external_publication") is False, "Presence references exact local-only external Body Form")
encounter = instances["BODY/LOCAL-FIELD/ENCOUNTER/INDEX.json"]
check(encounter.get("journal_owner") == "BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/local_journal.jsonl", "Encounter projects to mechanism-owned journal")
check(encounter.get("live_count_mirrored") is False and encounter.get("live_tip_mirrored") is False and encounter.get("raw_bytes_interpreted_or_admitted") is False, "Encounter creates no live mirror or semantic uptake")
aperture = instances["BODY/LOCAL-FIELD/APERTURE/INDEX.json"]
check(aperture.get("status") == "NONE_ACTIVE" and aperture.get("activation_mode") == "EXPLICIT_ONE_SHOT_LOCAL_SOCKET" and aperture.get("autonomous_opening") is False, "no aperture active or autonomous")
residue = instances["BODY/LOCAL-FIELD/RESIDUE/INDEX.json"]
check(residue.get("records") == [] and residue.get("record_count") == 0 and residue.get("semantic_admission_created_by_crossing") is False, "no Local Field residue admitted")

lineage = instances["BODY/LINEAGE/INDEX.json"]
check(lineage.get("represented_currentness_revision") == 7, "Lineage projects revision 7")
chains = lineage.get("motion_chains", [])
check(len(chains) == 8 and all(chain.get("retrospectively_regulated") is False for chain in chains), "eight non-retroactive motion chains")
chain = chains[-1] if len(chains) == 8 else {}
check(chain.get("classification") == "BODY_LOCAL_ADAPTATION_AND_OPERATIONAL_REBINDING" and chain.get("result") == "ADAPTED_AND_REBOUND", "adaptation lineage classification and result")
referenced_hash(chain.get("candidate", {}), "adaptation Lineage candidate")
referenced_hash(chain.get("act", {}), "adaptation Lineage Act")
check(chain.get("occurrence", {}).get("binding") == "CURRENT_STATE_CARRIES_OCCURRENCE_METADATA", "occurrence circularity avoided")
check(chain.get("successor_state", {}).get("binding") == "OCCURRENCE_CARRIES_FINAL_STATE_HASH", "successor circularity avoided")
check(sha(ROOT / "BODY/LINEAGE/STATE_REVISIONS/CURRENT_REVISION_006.json") == REV6_CURRENT_SHA, "exact revision-6 state preserved")

motion = current.get("semantic_anatomy_adaptation_occurrence", {})
check(motion.get("posture") == "ADAPT_AND_REBIND" and motion.get("functions_invoked_distinctly") == ["ADAPTATION_AUTHORITY", "OPERATIONAL_AUTHORIZATION"], "current records selected posture and functions")
check(motion.get("candidate_sha256") == CANDIDATE_SHA and motion.get("source_statement_sha256") == SOURCE_STATEMENT_SHA, "current binds exact decision evidence")
check(motion.get("predecessor_revision") == 6 and motion.get("predecessor_state_sha256") == REV6_CURRENT_SHA, "current binds exact predecessor")
referenced_hash({"carrier": motion.get("candidate_carrier"), "sha256": motion.get("candidate_sha256")}, "current adaptation candidate")
referenced_hash({"carrier": motion.get("act_carrier"), "sha256": motion.get("act_carrier_sha256")}, "current adaptation Act")

mechanism = ROOT / "BODY/EMBODIMENT/MECHANISMS/REALITY_CONTACT_MEMBRANE"
unchanged_mechanism = {
    "SOURCE_PROVENANCE.json": "8373c0f3909f29f2dda71ea1065a5d6608eeee6f8aa66be703a4e584eae99a3b",
    "V0_1/contact_contract.json": "5f5560af5010c8c173853adef005711b1d7d8b7882b9bbbbbc97a8857886d5d5",
    "V0_1/live_pressure_envelope.json": "c274e4d034ff5ebf42e1120010cb5d20eb0c5415154f4dc164376bf15717b3e5",
    "V0_1/reality_contact_membrane.py": "99b770a7eda474ddfc68a189871c3907b97231d8ff5091ad11df03317552e44f",
    "V0_1/verify_reality_contact_membrane.py": "a77d1d2f96497f0b6e2643fa5715b4733d667c8dc7e63499fc0d11c06a70b302",
}
for relative, expected in unchanged_mechanism.items():
    check(sha(mechanism / relative) == expected, f"mechanism semantic algorithm byte identity: {relative}")
binding = load_json(mechanism / "BODY_BINDING.json", "Body binding")
check(binding.get("status") == "ADMITTED_ACTIVE_ON_EXPLICIT_INVOCATION", "membrane activation posture unchanged")
check(binding.get("invocation", {}).get("entrypoint") == "BODY/EMBODIMENT/MECHANISMS/REALITY_CONTACT_MEMBRANE/V0_1/body_reality_contact_membrane.py", "membrane entrypoint rebound")
check(binding.get("invocation", {}).get("local_state_root") == "BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001", "membrane local state rebound")
check(binding.get("operational_rebinding", {}).get("effect_ceiling_changed") is False and binding.get("operational_rebinding", {}).get("activation_mode_changed") is False, "rebinding widens no mechanism power")
for key in ("persistent_daemon", "autonomous_invocation", "external_network_listener"):
    check(binding.get("invocation", {}).get(key) is False, f"invocation ceiling unchanged: {key}")
for key in ("may_validate_authority", "may_admit_matter", "may_create_relation", "may_create_contact_event", "may_create_consequence", "may_mutate_body_currentness", "may_determine_source_identity_or_continuity"):
    check(binding.get("effect_ceiling", {}).get(key) is False, f"mechanism non-power preserved: {key}")

embodiment = instances["BODY/EMBODIMENT/STATE.json"]
mechanisms = embodiment.get("mechanisms", [])
check(len(mechanisms) == 2, "exactly two admitted mechanisms")
host = next((entry for entry in mechanisms if entry.get("mechanism_id") == "RM-BODY-INTEGRITY-HOST-001"), {})
membrane_state = next((entry for entry in mechanisms if entry.get("mechanism_id") == "RM-REALITY-CONTACT-MEMBRANE-001"), {})
check(host.get("status") == "ADMITTED_INACTIVE" and host.get("activation_authorized") is False, "Integrity Host remains inactive")
check(membrane_state.get("status") == "ADMITTED_ACTIVE_ON_EXPLICIT_INVOCATION" and membrane_state.get("effect_ceiling_changed_by_rebinding") is False, "membrane remains bounded-active")
for key in ("binding", "implementation", "body_entrypoint", "contract", "pressure_verifier"):
    referenced_hash(membrane_state.get(key, {}), f"Embodiment membrane {key}")

pressure = subprocess.run([sys.executable, "-B", str(mechanism / "V0_1/verify_reality_contact_membrane.py")], capture_output=True, text=True)
check(pressure.returncode == 0 and "PASS 39/39 reality-contact membrane checks" in pressure.stdout, "Reality Contact Membrane exact 39-check pressure result")
wrapper = mechanism / "V0_1/body_reality_contact_membrane.py"
inside_socket = ROOT / "forbidden.sock"
refusal = subprocess.run([sys.executable, "-B", str(wrapper), "serve-once", "--socket", str(inside_socket)], capture_output=True, text=True)
check(refusal.returncode != 0 and "socket must remain outside the Body carrier" in refusal.stderr + refusal.stdout and not inside_socket.exists(), "rebound wrapper still refuses in-carrier socket")

live = ROOT / "BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001"
journal = live / "local_journal.jsonl"
raw_root = live / "raw_crossings"
check(sha(journal) == JOURNAL_SHA, "three-entry living journal byte-identical")
try:
    entries = [json.loads(line) for line in journal.read_text(encoding="utf-8").splitlines() if line]
except Exception:
    entries = []
    check(False, "living journal parses")
else:
    check(len(entries) == 3 and all(entry.get("event_type") == "CROSSING_OBSERVED" for entry in entries), "exactly three living crossings preserved")
check({path.name for path in raw_root.iterdir() if path.is_file()} == set(RAW_HASHES), "exact raw crossing carrier set")
for name, expected in RAW_HASHES.items():
    check(sha(raw_root / name) == expected, f"raw crossing byte identity: {name}")
state_check = subprocess.run([sys.executable, "-B", str(wrapper), "verify-state"], capture_output=True, text=True)
try:
    report = json.loads(state_check.stdout)
except Exception:
    report = {}
check(state_check.returncode == 0 and report.get("status") == "VALID" and report.get("entry_count") == 3, "rebound wrapper verifies preserved living state")

host_verifier = ROOT / "BODY/EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py"
host_result = subprocess.run([sys.executable, "-B", str(host_verifier)], capture_output=True, text=True)
check(host_result.returncode == 0 and "22" in host_result.stdout and "PASS" in host_result.stdout.upper(), "Integrity Host unchanged 22-case result")

recovery_map = load_json(ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_006_MAP.json", "revision-6 recovery map")
check(recovery_map.get("source_manifest_sha256") == REV6_MANIFEST_SHA and recovery_map.get("predecessor_recovery_resolution", {}).get("recursive_duplication") is False, "revision-6 recovery map avoids recursive duplication")
snapshot = ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_006"
snapshot_manifest = snapshot / "CARRIER/MANIFEST.sha256"
check(sha(snapshot_manifest) == REV6_MANIFEST_SHA, "revision-6 manifest byte identity")
revision6_entries = parse_manifest(snapshot_manifest, "revision 6")
for relative, expected in revision6_entries.items():
    path = ROOT / relative if relative.startswith("CARRIER/RECOVERY/") else snapshot / relative
    check(path.is_file() and sha(path) == expected, f"revision-6 recoverable byte: {relative}")
snapshot_files = {path.relative_to(snapshot).as_posix() for path in snapshot.rglob("*") if path.is_file()}
snapshot_expected = {relative for relative in revision6_entries if not relative.startswith("CARRIER/RECOVERY/")} | {"CARRIER/MANIFEST.sha256"}
check(snapshot_files == snapshot_expected, "revision-6 snapshot has exact nonrecursive static file set")
check(sha(snapshot / "CURRENT/STATE.json") == REV6_CURRENT_SHA, "revision-6 recovery current state exact")
check(sha(snapshot / "LINEAGE/INDEX.json") == REV6_LINEAGE_SHA, "revision-6 recovery Lineage exact")
check(sha(snapshot / "EMBODIMENT/STATE.json") == REV6_EMBODIMENT_SHA, "revision-6 recovery Embodiment exact")

if FINAL:
    with tempfile.TemporaryDirectory(prefix="rm-revision6-recovery-", dir="/private/tmp") as temporary:
        restored = Path(temporary) / "LOCAL_REALITY_BODY_001"
        shutil.copytree(snapshot, restored)
        restored_pool = restored / "CARRIER/RECOVERY"
        restored_pool.mkdir(parents=True, exist_ok=True)
        for revision in range(1, 6):
            shutil.copytree(ROOT / f"CARRIER/RECOVERY/CARRIER_REVISION_{revision:03d}", restored_pool / f"CARRIER_REVISION_{revision:03d}")
        shutil.copytree(live, restored / "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001")
        old_result = subprocess.run([sys.executable, "-B", str(restored / "CARRIER/VERIFICATION/verify_carrier.py"), "--final"], capture_output=True, text=True)
        check(old_result.returncode == 0 and "RESULT  PASS FINAL (1490 checks)" in old_result.stdout, "revision-6 reconstructs and passes its own 1490-check verifier")

external_manifest = EXTERNAL / "MANIFEST.sha256"
check(EXTERNAL.is_dir() and sha(external_manifest) == EXTERNAL_MANIFEST_SHA, "external Body Form carrier and manifest unchanged")
external_entries = parse_manifest(external_manifest, "external Body Form")
for relative, expected in external_entries.items():
    check((EXTERNAL / relative).is_file() and sha(EXTERNAL / relative) == expected, f"external Body Form agreement: {relative}")
check(sha(EXTERNAL / "BODY_FORM.json") == BODY_FORM_SHA, "external Body Form payload unchanged")

occurrence = ROOT / occurrence_rel
if FINAL:
    occurrence_text = occurrence.read_text(encoding="utf-8") if occurrence.is_file() else ""
    check("Status: **PERFORMED**" in occurrence_text and "Result: `ADAPTED_AND_REBOUND`" in occurrence_text, "successful occurrence recorded")
    check(sha(ROOT / "BODY/STATE/CURRENT.json") in occurrence_text and sha(ROOT / "BODY/LINEAGE/INDEX.json") in occurrence_text, "occurrence binds final current and Lineage hashes")
    check(JOURNAL_SHA in occurrence_text and "Preserved living entry count: `3`" in occurrence_text, "occurrence binds living-state anchors")
else:
    check(not occurrence.exists(), "preflight precedes success occurrence")

manifest = ROOT / "CARRIER/MANIFEST.sha256"
eligible = {
    path.relative_to(ROOT).as_posix(): sha(path)
    for path in ROOT.rglob("*")
    if path.is_file()
    and path != manifest
    and path.name != ".DS_Store"
    and path.suffix not in {".pyc", ".pyo"}
    and "__pycache__" not in path.parts
    and not path.relative_to(ROOT).as_posix().startswith("BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/")
}
manifest_entries = parse_manifest(manifest, "revision 7")
check(manifest_entries == eligible, f"revision-7 complete static manifest agreement ({'final' if FINAL else 'preflight'})")
residue_paths = [path for path in ROOT.rglob("*") if path.is_file() and (path.name == ".DS_Store" or path.suffix in {".pyc", ".pyo"} or "__pycache__" in path.parts)]
check(residue_paths == [], "no runtime or machine residue")
check(not any(path.is_symlink() for path in ROOT.rglob("*")), "carrier contains no symbolic links")

non_effects = set(current.get("explicit_non_effects", []))
for item in ("NO_CONSTITUTIONAL_CHANGE", "NO_BODY_IDENTITY_CHANGE", "NO_LOCUS_CHANGE", "NO_AUTHORITY_TOPOLOGY_CHANGE", "NO_MATTER_RECEIPT_OR_ADMISSION", "NO_CONSEQUENCE_CREATED_OR_ADMITTED", "NO_INTEGRITY_HOST_CHANGE", "NO_PRIOR_EXTERNAL_CROSSING_RETROACTIVELY_ADMITTED", "NO_NEW_PARTICIPANT_OR_BODY", "NO_INTER_BODY_RELATION", "NO_RELATIONAL_FIELD", "NO_CONTACT_EVENT", "NO_SHARED_CURRENTNESS", "NO_PUBLIC_OR_NETWORK_PUBLICATION", "NO_CARRIER_OR_HOST_AUTHORITY"):
    check(item in non_effects, f"current non-effect: {item}")

if failures:
    print(f"\nRESULT  FAIL ({len(failures)} failures across {checks} checks)")
    for failure in failures:
        print(f"- {failure}")
    raise SystemExit(1)

mode = "PREFLIGHT" if PREFLIGHT else "FINAL"
print(f"\nRESULT  PASS {mode} ({checks} checks)")

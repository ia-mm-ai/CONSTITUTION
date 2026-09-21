#!/usr/bin/env python3
"""Fail-closed carrier verification for Local-Reality Body revision 6.

This verifier proves revision-5 recoverability, exact revision-6 static
carriers, the corrected membrane representation, dynamic living-state
ownership, living-state integrity, and the declared non-effects. It does not
originate Authority, an Act, an occurrence, an encounter, Matter, relation,
consequence, or currentness.
"""

from __future__ import annotations

import hashlib
import importlib.util
import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
EXTERNAL = Path("/Users/markomarkota/LOCAL_REALITY_BODY_FORMS/RM_LOCAL_REALITY_BODY_FORM_001")
PREFLIGHT = "--preflight" in sys.argv
FINAL = "--final" in sys.argv
if PREFLIGHT == FINAL:
    raise SystemExit("select exactly one of --preflight or --final")

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
                raise ValueError(f"invalid or duplicate entry: {line}")
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
    for heading in ("## Exact source statement", "## Exact statement"):
        if heading in lines:
            position = lines.index(heading) + 1
            while position < len(lines) and not lines[position].startswith("> "):
                position += 1
            values: list[str] = []
            while position < len(lines) and lines[position].startswith("> "):
                values.append(lines[position][2:])
                position += 1
            return "\n".join(values)
    return ""


REV5_STATE_SHA = "4d79749a3404f4705df3a9c37aec9f6088f23e5e3760bda4d13c584bd12e7543"
REV5_LINEAGE_SHA = "cb735a0125faad1fb22518eec8e3ddafb4e17f615704289bf9378ac24639344a"
REV5_EMBODIMENT_SHA = "86f8d46f051d24a42f0777ddd2e64ae9f1d124985a730bc14d90b211fc9c8198"
REV5_MANIFEST_SHA = "25c7ad1b8e4c53cb3ccdff83abc9ba3f0242a5cfe527d2d1addaf1b8f5849014"
CURRENT_SHA = "7fc906ac68fdf350d6a91cfed1119902f8635c8308af9d693a0f6e491c354038"
LINEAGE_SHA = "0521785f404bda73ba2675fef331cf9755c8be588248d60a0e2e802848af2d49"
CANDIDATE_SHA = "0ebab0ccaf19ae3168b3119f8966e9e2f2cb6ac8909c27fa43a43cf73e6875db"
ACT_SHA = "53e146347a13a9f9523b9534e768e7fb5a63545dbef7af0215a6f86eea45d576"
SOURCE_STATEMENT_SHA = "80c07be038073b7456009902e017f790b8fee9d3c6cbafc0bb77565b362ab602"
FIRST_ENTRY_SHA = "a7e16911aba51f1099ba9e746a38f4e433b6f28966d59f6b6f0c684677ebb88b"
FIRST_RAW_SHA = "765df8a6f7f898f52cf17c924c756afd71b5a281fd3c995f0145a346536fbbc6"
SECOND_ENTRY_SHA = "7d81684301e2790686edfbc894463299b99d0bdc4b217a9ae48efe5a04c169b5"
SECOND_RAW_SHA = "91ae6dfe68cce0764232618d618161ccaa2675b18a3f6043cc246c955dbfc36d"
BODY_FORM_SHA = "4031ba2a6b7bc7cbe7352bdd91b351227aa1d517a0d4680d4fafc6afc26ca056"
EXTERNAL_MANIFEST_SHA = "da93daeba6b0f8e2c49521f4b043be35330d103422dae1ddade82670f43f5a2c"

MECHANISM = "EMBODIMENT/MECHANISMS/REALITY_CONTACT_MEMBRANE"
MECHANISM_HASHES = {
    f"{MECHANISM}/BODY_BINDING.json": "2a184a88c1b03a601938996d3329b597dadff2729ebfe775decf0bcc09c16a68",
    f"{MECHANISM}/README.md": "c3718a36a439a6e56936e011e89c6f808c8dcc7edeffd1e83bbc5da580ab6ca4",
    f"{MECHANISM}/SOURCE_PROVENANCE.json": "8373c0f3909f29f2dda71ea1065a5d6608eeee6f8aa66be703a4e584eae99a3b",
    f"{MECHANISM}/V0_1/body_reality_contact_membrane.py": "c3a3405a0dfb05395220e7248095e1864e99f6359228809ddfbc2c1d982bc5cb",
    f"{MECHANISM}/V0_1/contact_contract.json": "5f5560af5010c8c173853adef005711b1d7d8b7882b9bbbbbc97a8857886d5d5",
    f"{MECHANISM}/V0_1/live_pressure_envelope.json": "c274e4d034ff5ebf42e1120010cb5d20eb0c5415154f4dc164376bf15717b3e5",
    f"{MECHANISM}/V0_1/reality_contact_membrane.py": "99b770a7eda474ddfc68a189871c3907b97231d8ff5091ad11df03317552e44f",
    f"{MECHANISM}/V0_1/verify_reality_contact_membrane.py": "a77d1d2f96497f0b6e2643fa5715b4733d667c8dc7e63499fc0d11c06a70b302",
}
OWNER_HASHES = {
    "MATTER/INDEX.json": "51b64d3a98397b1c256509a3da419cc422bf1fb9e554021e0015fe7aec538c60",
    "REGULATION/INDEX.json": "9fa9759a19bb06aec71ce6e0967d3f102a0ff968288037b36b2725389c586916",
    "CONSEQUENCES/INDEX.json": "a06426b447a1896a41045e43d44742d037f63b422055ad45885907408d8cf669",
    "RELATIONS/INDEX.json": "8d3244652091017a5e457660da64967039db42588b7f86c90cbd017109bb3f13",
    "EMBODIMENT/STATE.json": "b4d662000308cde1c7b44e4efbea4371f21b396b0a1117f177bac5e0ede71d32",
    "EPISTEMIC/INDEX.json": "ea2eec8bc0faa54083241d3198ea0343e07428ab002edfcabbb5682c6609237c",
}

required = {
    "BODY.md", "CORE/IDENTITY.md", "CORE/LAW.md", "CORE/LOCUS.md", "CORE/APPLICABILITY.json",
    "CURRENT/STATE.json", "LINEAGE/INDEX.json",
    "LINEAGE/CANDIDATES/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_CANDIDATE_001.md",
    "LINEAGE/ACTS/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_ACT_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_005.json",
    "CARRIER/RECOVERY/CARRIER_REVISION_005/CARRIER/MANIFEST.sha256",
    "CARRIER/VERIFICATION/verify_carrier.py", "CARRIER/MANIFEST.sha256",
} | set(MECHANISM_HASHES) | set(OWNER_HASHES)
if FINAL:
    required.add("LINEAGE/OCCURRENCES/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_OCCURRENCE_001.md")
for relative in sorted(required):
    check((ROOT / relative).is_file(), f"required carrier: {relative}")


# Exact revision-5 static predecessor recovery, including its own successful verifier.
recovery = ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_005"
recovery_manifest = recovery / "CARRIER/MANIFEST.sha256"
check(recovery_manifest.is_file() and sha(recovery_manifest) == REV5_MANIFEST_SHA, "revision-5 manifest byte identity")
recovery_entries = parse_manifest(recovery_manifest, "revision 5")
for relative, expected in sorted(recovery_entries.items()):
    path = recovery / relative
    check(path.is_file(), f"revision-5 recovery exists: {relative}")
    check(path.is_file() and sha(path) == expected, f"revision-5 recovery hash: {relative}")
recovered_files = {
    path.relative_to(recovery).as_posix() for path in recovery.rglob("*")
    if path.is_file() and path.name != ".DS_Store"
}
check(recovered_files == set(recovery_entries) | {"CARRIER/MANIFEST.sha256"}, "complete revision-5 static recovery file set")
recovered_verifier = recovery / "CARRIER/VERIFICATION/verify_carrier.py"
recovered = subprocess.run([sys.executable, "-B", str(recovered_verifier), "--final"], capture_output=True, text=True)
check(recovered.returncode == 0, "recovered revision-5 verifier passes")
check("RESULT  PASS FINAL (814 checks)" in recovered.stdout, "recovered revision-5 exact 814-check result")
check(sha(recovery / "CURRENT/STATE.json") == REV5_STATE_SHA, "recovered revision-5 current state")
check(sha(recovery / "LINEAGE/INDEX.json") == REV5_LINEAGE_SHA, "recovered revision-5 lineage")
check(sha(recovery / "EMBODIMENT/STATE.json") == REV5_EMBODIMENT_SHA, "recovered revision-5 embodiment")
check(not (recovery / "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001").exists(), "revision-5 static recovery does not duplicate living state")
check(sha(ROOT / "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_005.json") == REV5_STATE_SHA, "exact revision-5 state revision")


# Every predecessor file not named as a revision-6 projection or verifier remains exact.
changed = {
    "CURRENT/STATE.json", "LINEAGE/INDEX.json", "MATTER/INDEX.json", "REGULATION/INDEX.json",
    "CONSEQUENCES/INDEX.json", "RELATIONS/INDEX.json", "EMBODIMENT/STATE.json", "EPISTEMIC/INDEX.json",
    "CARRIER/SCHEMAS/current-state.schema.json", "CARRIER/SCHEMAS/lineage-index.schema.json",
    "CARRIER/SCHEMAS/matter-index.schema.json", "CARRIER/SCHEMAS/regulation-index.schema.json",
    "CARRIER/SCHEMAS/consequence-index.schema.json", "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json", "CARRIER/SCHEMAS/epistemic-index.schema.json",
    "CARRIER/VERIFICATION/verify_carrier.py",
}
for relative, expected in sorted(recovery_entries.items()):
    if relative not in changed:
        check((ROOT / relative).is_file() and sha(ROOT / relative) == expected, f"unchanged revision-5 byte identity: {relative}")


# Exact selection evidence and source statement.
candidate = ROOT / "LINEAGE/CANDIDATES/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_CANDIDATE_001.md"
act = ROOT / "LINEAGE/ACTS/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_ACT_001.md"
check(sha(candidate) == CANDIDATE_SHA, "exact membrane candidate")
check(sha(act) == ACT_SHA, "exact membrane Act")
statement = quoted_source(act)
check(statement == "ok bitch, than proceed with coherence update (self-maintance, self-reflection) 💅", "exact selecting source statement")
check(hashlib.sha256(statement.encode("utf-8")).hexdigest() == SOURCE_STATEMENT_SHA, "selecting source statement hash")
act_text = act.read_text(encoding="utf-8")
check("Function invoked: `CORRECTION_AUTHORITY`" in act_text, "Act invokes Correction Authority")
check("does not invoke `ADAPTATION_AUTHORITY`" in act_text, "Act excludes Adaptation Authority")
check("locality relocation" in act_text, "Act excludes locality relocation")


# Currentness and exact owner references.
check(sha(ROOT / "CURRENT/STATE.json") == CURRENT_SHA, "exact revision-6 current state")
check(sha(ROOT / "LINEAGE/INDEX.json") == LINEAGE_SHA, "exact revision-6 lineage")
current = load_json(ROOT / "CURRENT/STATE.json", "current state")
check(current.get("schema") == "local-reality-body.current-state.v0.6", "current schema identity")
check(current.get("body_id") == "RM-LOCAL-REALITY-BODY-001" and current.get("body_status") == "FORMED", "Body identity and formation unchanged")
check(current.get("currentness_revision") == 6, "currentness revision 6")
check(current.get("initial_form_posture_ceiling", {}).get("operational_activation_claimed") is False, "initial-form claim remains historically non-operational")
check(current.get("current_authority_relations", {}).get("bearer") == "Marko Markota", "Authority bearer unchanged")
check(current.get("current_authority_relations", {}).get("functions") == [
    "LOCAL_GOVERNING_AUTHORITY", "OPERATIONAL_AUTHORIZATION", "CORRECTION_AUTHORITY",
    "ADAPTATION_AUTHORITY", "DORMANCY_AND_CESSATION",
], "Authority topology unchanged")
for key in ("identity", "law", "locus", "applicability"):
    referenced_hash(current.get("core", {}).get(key, {}), f"current core {key}")
referenced_hash(current.get("lineage", {}), "current lineage")
referenced_hash(current.get("relations", {}), "current relations")
referenced_hash(current.get("embodiment", {}), "current embodiment")
for relative, expected in OWNER_HASHES.items():
    check(sha(ROOT / relative) == expected, f"exact revision-6 owner: {relative}")
    owner = load_json(ROOT / relative, relative)
    check(owner.get("represented_currentness_revision") == 6, f"owner projects revision 6: {relative}")

motion = current.get("reality_contact_membrane_coherence_correction_occurrence", {})
check(motion.get("result") == "COHERENCE_CORRECTED_LIVING_STATE_PRESERVED", "current correction result")
check(motion.get("posture") == "CORRECT" and motion.get("function_invoked") == "CORRECTION_AUTHORITY", "current correction posture and function")
check(motion.get("predecessor_revision") == 5 and motion.get("predecessor_state_sha256") == REV5_STATE_SHA, "current exact predecessor binding")
check(motion.get("candidate_sha256") == CANDIDATE_SHA and motion.get("act_carrier_sha256") == ACT_SHA, "current exact decision evidence")
check(motion.get("source_statement_sha256") == SOURCE_STATEMENT_SHA, "current source statement binding")
check(motion.get("recovery_carrier") == "CARRIER/RECOVERY/CARRIER_REVISION_005", "current recovery route")

capability = current.get("operational_capability", {})
check(capability.get("status") == "ADMITTED_ACTIVE_ON_EXPLICIT_INVOCATION", "bounded capability active")
check(capability.get("entrypoint") == f"{MECHANISM}/V0_1/body_reality_contact_membrane.py", "fixed Body entrypoint")
check(capability.get("entrypoint_sha256") == MECHANISM_HASHES[f"{MECHANISM}/V0_1/body_reality_contact_membrane.py"], "current cross-checks Body entrypoint hash")
check(capability.get("local_state_root") == "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001", "fixed Body local state")
check(capability.get("living_state_owner") == "MECHANISM_LOCAL_JOURNAL", "mechanism-local journal owns living state")
check(capability.get("live_record_count_source") == "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/local_journal.jsonl", "current routes live count to journal")
check(capability.get("static_currentness_owns_live_record_count") is False, "static currentness does not own live count")
check(capability.get("static_currentness_owns_live_journal_tip") is False, "static currentness does not own journal tip")
check("body_local_encounter_record_count" not in capability, "mutable live count absent from static capability")
for key in ("autonomous_invocation", "external_network_listener", "may_admit_matter", "may_create_relation_or_contact_event", "may_create_consequence", "may_mutate_currentness"):
    check(capability.get(key) is False, f"capability ceiling: {key}")
exposure = current.get("exposure", {})
check(exposure.get("body_form_exposure_does_not_claim_receiver_encounter") is True, "Body Form exposure does not claim receiver encounter")
check(exposure.get("live_encounter_state_owner") == "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/local_journal.jsonl", "exposure routes living state to journal")
check(exposure.get("static_currentness_owns_live_encounter_status") is False and exposure.get("static_currentness_owns_live_encounter_record_count") is False, "exposure refuses mutable live-state ownership")
check("body_local_encounter_record_count" not in exposure and "body_local_encounter_claimed" not in exposure, "stale live mirrors removed from exposure")
check(exposure.get("external_encounter_claims_admitted") is False and exposure.get("receiver_formation_claimed") is False, "no external encounter or receiver admitted")
check(exposure.get("body_form_sha256") == BODY_FORM_SHA, "exposed Body Form unchanged")


# Empty owners and unchanged relation topology.
matter = load_json(ROOT / "MATTER/INDEX.json", "Matter index")
regulation = load_json(ROOT / "REGULATION/INDEX.json", "Regulation index")
consequences = load_json(ROOT / "CONSEQUENCES/INDEX.json", "Consequence index")
relations = load_json(ROOT / "RELATIONS/INDEX.json", "Relation index")
epistemic = load_json(ROOT / "EPISTEMIC/INDEX.json", "Epistemic index")
check(matter.get("records") == [] and matter.get("record_count") == 0, "Matter remains empty")
check(regulation.get("records") == [] and regulation.get("record_count") == 0 and regulation.get("mechanism_present") is False, "Regulation remains empty and mechanism-free")
check(consequences.get("records") == [] and consequences.get("record_count") == 0, "Consequences remain empty")
for key in ("admitted_records", "source_relations", "observations", "evidence", "inferences", "working_interpretations", "consequence_observations", "standing_claims"):
    check(epistemic.get(key) == [], f"Epistemic remains empty: {key}")
check(len(relations.get("human_body_local", [])) == 1 and len(relations.get("authority_relations", [])) == 5, "existing local and Authority relations unchanged")
for key in ("inter_body", "relational_fields", "contact_events", "witness_relations", "external_relations"):
    check(relations.get(key) == [], f"no relation created: {key}")


# Exact mechanism, binding, and living-state boundaries.
for relative, expected in sorted(MECHANISM_HASHES.items()):
    check(sha(ROOT / relative) == expected, f"exact mechanism byte identity: {relative}")
binding = load_json(ROOT / MECHANISM / "BODY_BINDING.json", "Body binding")
check(binding.get("status") == "ADMITTED_ACTIVE_ON_EXPLICIT_INVOCATION", "binding activation status")
authority = binding.get("authority", {})
check(authority.get("adaptation_function") == "ADAPTATION_AUTHORITY" and authority.get("operational_function") == "OPERATIONAL_AUTHORIZATION", "binding separates Authority functions")
check(authority.get("common_bearership_merges_functions") is False, "binding does not merge functions")
invocation = binding.get("invocation", {})
check(invocation.get("mode") == "EXPLICIT_ONE_SHOT_LOCAL_SOCKET", "one-shot invocation only")
check(invocation.get("socket_must_be_outside_carrier") is True, "socket outside carrier")
for key in ("persistent_daemon", "autonomous_invocation", "external_network_listener"):
    check(invocation.get(key) is False, f"invocation ceiling: {key}")
effect = binding.get("effect_ceiling", {})
for key in ("may_validate_authority", "may_admit_matter", "may_create_relation", "may_create_contact_event", "may_create_consequence", "may_mutate_body_currentness", "may_determine_source_identity_or_continuity"):
    check(effect.get(key) is False, f"mechanism non-power: {key}")

embodiment = load_json(ROOT / "EMBODIMENT/STATE.json", "Embodiment state")
mechanisms = embodiment.get("mechanisms", [])
check(len(mechanisms) == 2, "exactly two admitted mechanisms")
host = next((entry for entry in mechanisms if entry.get("mechanism_id") == "RM-BODY-INTEGRITY-HOST-001"), {})
membrane = next((entry for entry in mechanisms if entry.get("mechanism_id") == "RM-REALITY-CONTACT-MEMBRANE-001"), {})
check(host.get("status") == "ADMITTED_INACTIVE" and host.get("activation_authorized") is False, "Integrity Host unchanged and inactive")
check(membrane.get("status") == "ADMITTED_ACTIVE_ON_EXPLICIT_INVOCATION", "membrane admitted and bounded-active")
check(membrane.get("pressure_verifier", {}).get("activation_result") == "PASS_39_OF_39", "embodiment binds pressure result")
for key in ("binding", "implementation", "body_entrypoint", "contract", "pressure_verifier"):
    referenced_hash(membrane.get(key, {}), f"embodiment membrane {key}")
surfaces = embodiment.get("surfaces", [])
surface = surfaces[0] if len(surfaces) == 1 else {}
check(len(surfaces) == 1 and surface.get("activation_baseline_record_count") == 0, "living surface preserves activation baseline")
check(surface.get("live_status_owner") == "MECHANISM_LOCAL_JOURNAL", "journal owns live surface status")
check(surface.get("live_record_count_owner") == "MECHANISM_LOCAL_JOURNAL", "journal owns live surface count")
check(surface.get("live_journal_tip_owner") == "MECHANISM_LOCAL_JOURNAL", "journal owns live surface tip")
check(surface.get("static_projection_claims_current_live_status") is False and surface.get("static_projection_claims_current_live_record_count") is False, "Embodiment refuses mutable live mirrors")
check("status" not in surface and "record_count" not in surface, "stale mutable surface fields removed")
check(surface.get("records_created_only_on_actual_crossing") is True, "living records require actual crossing")
check(surface.get("static_manifest_coverage") is False and surface.get("independent_hash_chain_verification_required") is True, "living state has independent integrity owner")

pressure_verifier = ROOT / MECHANISM / "V0_1/verify_reality_contact_membrane.py"
pressure = subprocess.run([sys.executable, "-B", str(pressure_verifier)], capture_output=True, text=True)
check(pressure.returncode == 0, "Reality Contact Membrane pressure verifier returns success")
check("PASS 39/39 reality-contact membrane checks" in pressure.stdout, "Reality Contact Membrane exact 39-check result")

wrapper = ROOT / MECHANISM / "V0_1/body_reality_contact_membrane.py"
inside_socket = ROOT / "forbidden.sock"
refusal = subprocess.run([sys.executable, "-B", str(wrapper), "serve-once", "--socket", str(inside_socket)], capture_output=True, text=True)
check(refusal.returncode != 0 and "socket must remain outside the Body carrier" in refusal.stderr + refusal.stdout, "Body wrapper refuses in-carrier socket")
check(not inside_socket.exists(), "wrapper refusal creates no socket")

# A disposable Body-shaped root proves the fixed receiver path without touching live state.
with tempfile.TemporaryDirectory(prefix="rm-body-wrapper-") as temporary:
    disposable = Path(temporary) / "BODY"
    disposable_package = disposable / MECHANISM / "V0_1"
    disposable_package.mkdir(parents=True)
    for name in ("body_reality_contact_membrane.py", "reality_contact_membrane.py"):
        shutil.copy2(ROOT / MECHANISM / "V0_1" / name, disposable_package / name)
    result = subprocess.run([sys.executable, "-B", str(disposable_package / "body_reality_contact_membrane.py"), "verify-state"], capture_output=True, text=True)
    check(result.returncode == 0, "disposable Body wrapper executes")
    try:
        reported = json.loads(result.stdout)
    except Exception:
        reported = {}
    check(reported == {"entry_count": 0, "journal_tip": None, "status": "VALID"}, "disposable Body wrapper begins with zero encounters")
    expected_root = disposable / "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001"
    check(not expected_root.exists(), "read-only verification does not instantiate recipient-local state")
    check(not any(path.is_file() for path in disposable.rglob("*") if path.parent != disposable_package), "wrapper check invents no crossing")

live_state = ROOT / "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001"
check(live_state.is_dir(), "preserved living state root exists")
check(not any(path.is_symlink() for path in live_state.rglob("*")), "living state contains no symbolic links")
root_names = {path.name for path in live_state.iterdir()} if live_state.is_dir() else set()
check(root_names == {"local_journal.jsonl", "raw_crossings"}, "living root contains only the mechanism-owned journal and carrier directory")
raw_root = live_state / "raw_crossings"
check(raw_root.is_dir(), "raw crossing directory exists")
check(not any(path.is_dir() for path in raw_root.rglob("*")), "raw crossing directory has no nested directories")
check(all(path.is_file() and path.suffix == ".bin" for path in raw_root.iterdir()) if raw_root.is_dir() else False, "raw crossing directory contains only binary carriers")

journal_path = live_state / "local_journal.jsonl"
try:
    living_entries = [json.loads(line) for line in journal_path.read_text(encoding="utf-8").splitlines() if line.strip()]
except Exception:
    living_entries = []
    check(False, "living journal parses")
else:
    check(all(isinstance(entry, dict) for entry in living_entries), "living journal contains objects")

state_check = subprocess.run([sys.executable, "-B", str(wrapper), "verify-state"], capture_output=True, text=True)
check(state_check.returncode == 0, "living state hash chain verifies")
try:
    state_report = json.loads(state_check.stdout)
except Exception:
    state_report = {}
check(state_report.get("status") == "VALID", "living wrapper reports valid")
check(state_report.get("entry_count") == len(living_entries), "dynamic entry count comes from mechanism-local journal")
check(state_report.get("journal_tip") == (living_entries[-1].get("entry_sha256") if living_entries else None), "dynamic journal tip comes from mechanism-local journal")

crossings = [entry for entry in living_entries if entry.get("event_type") == "CROSSING_OBSERVED"]
referenced_raw = {entry.get("data", {}).get("raw_carrier") for entry in crossings}
actual_raw = {path.relative_to(live_state).as_posix() for path in raw_root.iterdir() if path.is_file()} if raw_root.is_dir() else set()
check(None not in referenced_raw and actual_raw == referenced_raw, "every raw carrier is referenced exactly by the living journal")
check(len(living_entries) >= 2 and living_entries[0].get("entry_sha256") == FIRST_ENTRY_SHA, "first living crossing entry preserved")
check(len(living_entries) >= 2 and living_entries[0].get("data", {}).get("wire_sha256") == FIRST_RAW_SHA, "first living crossing bytes preserved")
check(len(living_entries) >= 2 and living_entries[1].get("entry_sha256") == SECOND_ENTRY_SHA, "second living crossing entry preserved")
check(len(living_entries) >= 2 and living_entries[1].get("data", {}).get("wire_sha256") == SECOND_RAW_SHA, "second living crossing bytes preserved")
check((live_state / "raw_crossings/rcm-1788365297321538000-38e04707990d.bin").is_file() and sha(live_state / "raw_crossings/rcm-1788365297321538000-38e04707990d.bin") == FIRST_RAW_SHA, "first raw carrier exact")
check((live_state / "raw_crossings/rcm-1788366247790802000-0e6e87ae457a.bin").is_file() and sha(live_state / "raw_crossings/rcm-1788366247790802000-0e6e87ae457a.bin") == SECOND_RAW_SHA, "second raw carrier exact")

# A disposable clone proves later crossings advance only the mechanism-local
# journal and carriers; static Body currentness stays byte-identical.
with tempfile.TemporaryDirectory(prefix="rm-body-living-growth-") as temporary:
    disposable_state = Path(temporary) / "REALITY_CONTACT_MEMBRANE_001"
    shutil.copytree(live_state, disposable_state)
    implementation_path = ROOT / MECHANISM / "V0_1/reality_contact_membrane.py"
    spec = importlib.util.spec_from_file_location("rm_reality_contact_membrane_revision6_test", implementation_path)
    module = importlib.util.module_from_spec(spec) if spec and spec.loader else None
    if module is not None and spec is not None and spec.loader is not None:
        spec.loader.exec_module(module)
        before_count = len(module.RealityContactMembrane(disposable_state).verify())
        before_current = sha(ROOT / "CURRENT/STATE.json")
        module.RealityContactMembrane(disposable_state).observe_wire_crossing(b'{"test":"future-living-growth"}')
        after_entries = module.RealityContactMembrane(disposable_state).verify()
        check(len(after_entries) == before_count + 1, "future disposable crossing advances living count")
        check(sha(ROOT / "CURRENT/STATE.json") == before_current, "future living growth requires no static currentness mutation")
    else:
        check(False, "load membrane implementation for future-growth test")
        check(False, "future living growth leaves static currentness unchanged")

host_verifier = ROOT / "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py"
host_result = subprocess.run([sys.executable, "-B", str(host_verifier)], capture_output=True, text=True)
check(host_result.returncode == 0 and "22" in host_result.stdout and "PASS" in host_result.stdout.upper(), "Integrity Host unchanged 22-case pressure result")


# Lineage and schemas describe the revision without substituting for exact evidence.
lineage = load_json(ROOT / "LINEAGE/INDEX.json", "Lineage index")
check(lineage.get("represented_currentness_revision") == 6, "lineage projects revision 6")
chains = lineage.get("motion_chains", [])
check(len(chains) == 7 and all(chain.get("retrospectively_regulated") is False for chain in chains), "seven non-retroactive motion chains")
chain = chains[-1] if len(chains) == 7 else {}
check(chain.get("classification") == "BODY_LOCAL_CORRECTION", "correction lineage classification")
check(chain.get("result") == "COHERENCE_CORRECTED_LIVING_STATE_PRESERVED", "correction lineage result")
referenced_hash(chain.get("candidate", {}), "correction lineage candidate")
referenced_hash(chain.get("act", {}), "correction lineage Act")
check(chain.get("occurrence", {}).get("binding") == "CURRENT_STATE_CARRIES_OCCURRENCE_METADATA", "occurrence circularity avoided")
check(chain.get("successor_state", {}).get("binding") == "OCCURRENCE_CARRIES_FINAL_STATE_HASH", "successor circularity avoided")

for relative in (
    "CARRIER/SCHEMAS/matter-index.schema.json", "CARRIER/SCHEMAS/regulation-index.schema.json",
    "CARRIER/SCHEMAS/consequence-index.schema.json", "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json", "CARRIER/SCHEMAS/epistemic-index.schema.json",
    "CARRIER/SCHEMAS/lineage-index.schema.json",
):
    schema = load_json(ROOT / relative, relative)
    check(schema.get("properties", {}).get("represented_currentness_revision", {}).get("const") == 6, f"schema projects revision 6: {relative}")
current_schema = load_json(ROOT / "CARRIER/SCHEMAS/current-state.schema.json", "current-state schema")
check(current_schema.get("$id") == "local-reality-body.current-state.v0.6", "current-state schema version")
check(current_schema.get("properties", {}).get("currentness_revision", {}).get("const") == 6, "current-state schema revision")
check({"reality_contact_membrane_occurrence", "reality_contact_membrane_coherence_correction_occurrence", "operational_capability"}.issubset(current_schema.get("required", [])), "current-state schema requires membrane motion, correction, and capability")


# The exposed Form remains exact and independent.
external_manifest = EXTERNAL / "MANIFEST.sha256"
check(EXTERNAL.is_dir(), "external Body Form carrier present")
check(external_manifest.is_file() and sha(external_manifest) == EXTERNAL_MANIFEST_SHA, "external Body Form manifest unchanged")
external_entries = parse_manifest(external_manifest, "external Body Form") if external_manifest.is_file() else {}
for relative, expected in external_entries.items():
    check((EXTERNAL / relative).is_file() and sha(EXTERNAL / relative) == expected, f"external Body Form agreement: {relative}")
check((EXTERNAL / "BODY_FORM.json").is_file() and sha(EXTERNAL / "BODY_FORM.json") == BODY_FORM_SHA, "external Body Form payload unchanged")


# Final occurrence binds the already-stable current and lineage hashes.
occurrence_relative = "LINEAGE/OCCURRENCES/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_OCCURRENCE_001.md"
occurrence_path = ROOT / occurrence_relative
if FINAL:
    occurrence_text = occurrence_path.read_text(encoding="utf-8") if occurrence_path.is_file() else ""
    check("Status: **PERFORMED**" in occurrence_text, "successful occurrence status")
    check(CURRENT_SHA in occurrence_text and LINEAGE_SHA in occurrence_text, "occurrence binds final current and lineage hashes")
    check("living entry count" in occurrence_text and FIRST_ENTRY_SHA in occurrence_text and SECOND_ENTRY_SHA in occurrence_text, "occurrence binds preserved living anchors")
    check("No crossing was admitted as Matter" in occurrence_text, "occurrence refuses semantic capture")
else:
    check(not occurrence_path.exists(), "preflight precedes success occurrence")


# Static manifest excludes exactly the living operational root and nothing else.
new_paths = set(MECHANISM_HASHES) | {
    "LINEAGE/CANDIDATES/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_CANDIDATE_001.md",
    "LINEAGE/ACTS/REALITY_CONTACT_MEMBRANE_COHERENCE_CORRECTION_ACT_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_005.json",
}
if FINAL:
    new_paths.add(occurrence_relative)
recovery_prefix = "CARRIER/RECOVERY/CARRIER_REVISION_005"
expected_static = set(recovery_entries) | new_paths | {
    f"{recovery_prefix}/{relative}" for relative in recovery_entries
} | {f"{recovery_prefix}/CARRIER/MANIFEST.sha256"}
manifest_path = ROOT / "CARRIER/MANIFEST.sha256"
eligible = {
    path.relative_to(ROOT).as_posix(): sha(path)
    for path in ROOT.rglob("*")
    if path.is_file()
    and path != manifest_path
    and path.name != ".DS_Store"
    and not path.relative_to(ROOT).as_posix().startswith("EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/")
}
check(set(eligible) == expected_static, f"revision-6 exact authorized static file set ({'final' if FINAL else 'preflight'})")
if FINAL:
    manifest_entries = parse_manifest(manifest_path, "revision 6") if manifest_path.is_file() else {}
    check(set(manifest_entries) == set(eligible), "revision-6 manifest complete static coverage")
    check(manifest_entries == eligible, "revision-6 manifest byte agreement")
else:
    check(sha(manifest_path) == REV5_MANIFEST_SHA, "preflight leaves predecessor manifest unclaimed")

residue = [
    path.relative_to(ROOT).as_posix() for path in ROOT.rglob("*")
    if path.is_file() and (path.name == ".DS_Store" or path.suffix in {".pyc", ".pyo"} or "__pycache__" in path.parts)
]
check(residue == [], "no runtime or machine residue in Body carrier")

non_effects = set(current.get("explicit_non_effects", []))
for item in (
    "NO_CONSTITUTIONAL_CHANGE", "NO_BODY_IDENTITY_CHANGE", "NO_LOCUS_CHANGE", "NO_AUTHORITY_TOPOLOGY_CHANGE",
    "NO_MATTER_RECEIPT_OR_ADMISSION", "NO_CONSEQUENCE_CREATED_OR_ADMITTED", "NO_INTEGRITY_HOST_CHANGE",
    "NO_BODY_ACT_PROCESSING_BY_ACTIVE_MECHANISM", "NO_BODY_CURRENTNESS_MUTATION_BY_ACTIVE_MECHANISM",
    "NO_PRIOR_EXTERNAL_CROSSING_RETROACTIVELY_ADMITTED", "NO_PERSISTENT_DAEMON_OR_AUTONOMOUS_INVOCATION",
    "NO_NEW_PARTICIPANT_OR_BODY", "NO_INTER_BODY_RELATION", "NO_RELATIONAL_FIELD", "NO_CONTACT_EVENT",
    "NO_SHARED_CURRENTNESS", "NO_PUBLIC_OR_NETWORK_PUBLICATION", "NO_CARRIER_OR_HOST_AUTHORITY",
):
    check(item in non_effects, f"current non-effect: {item}")

if failures:
    print(f"\nRESULT  FAIL ({len(failures)} failures across {checks} checks)")
    for failure in failures:
        print(f"- {failure}")
    raise SystemExit(1)

mode = "PREFLIGHT" if PREFLIGHT else "FINAL"
print(f"\nRESULT  PASS {mode} ({checks} checks)")

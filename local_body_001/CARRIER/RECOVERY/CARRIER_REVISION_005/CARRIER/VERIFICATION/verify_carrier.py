#!/usr/bin/env python3
"""Fail-closed carrier verification for Local-Reality Body revision 5.

This verifier proves predecessor recoverability, exact revision-5 static
carriers, the bounded contact-membrane implementation, living-state integrity,
and the declared non-effects. It does not originate Authority, an Act, an
occurrence, an encounter, Matter, relation, consequence, or currentness.
"""

from __future__ import annotations

import hashlib
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


REV4_STATE_SHA = "451afca839f93a99b2bac1e6c7226391df66aeffe1ee2caa92b14d93d381ff7c"
REV4_LINEAGE_SHA = "423b4ab2bafa8dff8441c62b062f1773ed3753a9abfce5faf76a0a9c2adda4b1"
REV4_EMBODIMENT_SHA = "bdd86dee604f1a725a9952e89887d95a5d56921b21319a40366d007fe7bd6a38"
REV4_MANIFEST_SHA = "237c2d713ec97e06c373e6e1f637836a0bc653057a98f58f559785c6ab13a63e"
CURRENT_SHA = "4d79749a3404f4705df3a9c37aec9f6088f23e5e3760bda4d13c584bd12e7543"
LINEAGE_SHA = "cb735a0125faad1fb22518eec8e3ddafb4e17f615704289bf9378ac24639344a"
CANDIDATE_SHA = "e79d4640f08bf291fbf7d221215c6cef2ddfb476a91ba0f31b7734d542f96292"
ACT_SHA = "14fdd9ab64b689aa0813395e5436a46ce2bd51e33fa8708a739867b1cac5fff0"
SOURCE_STATEMENT_SHA = "e5cc66c2f20d5e161621b1f2adfcaf591e0ed06a9a53defdff74092e6538d69b"
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
    "MATTER/INDEX.json": "07c2e4276530ff788dbdee67b048e6ae86ace3ff99e2c8ef3be639afd866b98c",
    "REGULATION/INDEX.json": "efefb2ece5ed0cdfd1ca99ab3dc0d39cee8626b8f9461baac1d7352d6891a1af",
    "CONSEQUENCES/INDEX.json": "07d6e328ca36641a2d03c7939212ae583bd7882747329558a93daf128e7a2f39",
    "RELATIONS/INDEX.json": "1524d2042f036c1c90b7ff3256055e75f0afd667858f571f44646308b2388472",
    "EMBODIMENT/STATE.json": "86f8d46f051d24a42f0777ddd2e64ae9f1d124985a730bc14d90b211fc9c8198",
    "EPISTEMIC/INDEX.json": "02b94730ba566114f0c240609ecec0964b70a148da7331ad7dc945ca386510c8",
}

required = {
    "BODY.md", "CORE/IDENTITY.md", "CORE/LAW.md", "CORE/LOCUS.md", "CORE/APPLICABILITY.json",
    "CURRENT/STATE.json", "LINEAGE/INDEX.json",
    "LINEAGE/CANDIDATES/REALITY_CONTACT_MEMBRANE_ACTIVATION_CANDIDATE_001.md",
    "LINEAGE/ACTS/REALITY_CONTACT_MEMBRANE_ACTIVATION_ACT_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_004.json",
    "CARRIER/RECOVERY/CARRIER_REVISION_004/CARRIER/MANIFEST.sha256",
    "CARRIER/VERIFICATION/verify_carrier.py", "CARRIER/MANIFEST.sha256",
} | set(MECHANISM_HASHES) | set(OWNER_HASHES)
if FINAL:
    required.add("LINEAGE/OCCURRENCES/REALITY_CONTACT_MEMBRANE_ACTIVATION_OCCURRENCE_001.md")
for relative in sorted(required):
    check((ROOT / relative).is_file(), f"required carrier: {relative}")


# Exact revision-4 predecessor recovery, including its own successful verifier.
recovery = ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_004"
recovery_manifest = recovery / "CARRIER/MANIFEST.sha256"
check(recovery_manifest.is_file() and sha(recovery_manifest) == REV4_MANIFEST_SHA, "revision-4 manifest byte identity")
recovery_entries = parse_manifest(recovery_manifest, "revision 4")
for relative, expected in sorted(recovery_entries.items()):
    path = recovery / relative
    check(path.is_file(), f"revision-4 recovery exists: {relative}")
    check(path.is_file() and sha(path) == expected, f"revision-4 recovery hash: {relative}")
recovered_files = {
    path.relative_to(recovery).as_posix() for path in recovery.rglob("*")
    if path.is_file() and path.name != ".DS_Store"
}
check(recovered_files == set(recovery_entries) | {"CARRIER/MANIFEST.sha256"}, "complete revision-4 recovery file set")
recovered_verifier = recovery / "CARRIER/VERIFICATION/verify_carrier.py"
recovered = subprocess.run([sys.executable, "-B", str(recovered_verifier), "--final"], capture_output=True, text=True)
check(recovered.returncode == 0, "recovered revision-4 verifier passes")
check("RESULT  PASS FINAL (532 checks)" in recovered.stdout, "recovered revision-4 exact 532-check result")
check(sha(recovery / "CURRENT/STATE.json") == REV4_STATE_SHA, "recovered revision-4 current state")
check(sha(recovery / "LINEAGE/INDEX.json") == REV4_LINEAGE_SHA, "recovered revision-4 lineage")
check(sha(recovery / "EMBODIMENT/STATE.json") == REV4_EMBODIMENT_SHA, "recovered revision-4 embodiment")
check(sha(ROOT / "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_004.json") == REV4_STATE_SHA, "exact revision-4 state revision")


# Every predecessor file not named as a revision-5 projection or verifier remains exact.
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
        check((ROOT / relative).is_file() and sha(ROOT / relative) == expected, f"unchanged revision-4 byte identity: {relative}")


# Exact selection evidence and source statement.
candidate = ROOT / "LINEAGE/CANDIDATES/REALITY_CONTACT_MEMBRANE_ACTIVATION_CANDIDATE_001.md"
act = ROOT / "LINEAGE/ACTS/REALITY_CONTACT_MEMBRANE_ACTIVATION_ACT_001.md"
check(sha(candidate) == CANDIDATE_SHA, "exact membrane candidate")
check(sha(act) == ACT_SHA, "exact membrane Act")
statement = quoted_source(act)
check(statement == "ok bitch, please proceed 💅", "exact selecting source statement")
check(hashlib.sha256(statement.encode("utf-8")).hexdigest() == SOURCE_STATEMENT_SHA, "selecting source statement hash")
act_text = act.read_text(encoding="utf-8")
for function in ("ADAPTATION_AUTHORITY", "OPERATIONAL_AUTHORIZATION"):
    check(function in act_text, f"Act invokes distinct function: {function}")
check("Common bearership does not merge the functions" in act_text, "Act preserves function separation")
check("cannot be generalized" in act_text and "relocation" in act_text, "Act excludes locality relocation")


# Currentness and exact owner references.
check(sha(ROOT / "CURRENT/STATE.json") == CURRENT_SHA, "exact revision-5 current state")
check(sha(ROOT / "LINEAGE/INDEX.json") == LINEAGE_SHA, "exact revision-5 lineage")
current = load_json(ROOT / "CURRENT/STATE.json", "current state")
check(current.get("schema") == "local-reality-body.current-state.v0.5", "current schema identity")
check(current.get("body_id") == "RM-LOCAL-REALITY-BODY-001" and current.get("body_status") == "FORMED", "Body identity and formation unchanged")
check(current.get("currentness_revision") == 5, "currentness revision 5")
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
    check(sha(ROOT / relative) == expected, f"exact revision-5 owner: {relative}")
    owner = load_json(ROOT / relative, relative)
    check(owner.get("represented_currentness_revision") == 5, f"owner projects revision 5: {relative}")

motion = current.get("reality_contact_membrane_occurrence", {})
check(motion.get("result") == "MECHANISM_ADMITTED_AND_ACTIVATED_BOUNDED", "current membrane result")
check(motion.get("posture") == "ADMIT_AND_ACTIVATE_BOUNDED", "current membrane posture")
check(motion.get("functions_invoked_distinctly") == ["ADAPTATION_AUTHORITY", "OPERATIONAL_AUTHORIZATION"], "current distinct function invocation")
check(motion.get("predecessor_revision") == 4 and motion.get("predecessor_state_sha256") == REV4_STATE_SHA, "current exact predecessor binding")
check(motion.get("candidate_sha256") == CANDIDATE_SHA and motion.get("act_carrier_sha256") == ACT_SHA, "current exact decision evidence")
check(motion.get("source_statement_sha256") == SOURCE_STATEMENT_SHA, "current source statement binding")
check(motion.get("recovery_carrier") == "CARRIER/RECOVERY/CARRIER_REVISION_004", "current recovery route")

capability = current.get("operational_capability", {})
check(capability.get("status") == "ADMITTED_ACTIVE_ON_EXPLICIT_INVOCATION", "bounded capability active")
check(capability.get("entrypoint") == f"{MECHANISM}/V0_1/body_reality_contact_membrane.py", "fixed Body entrypoint")
check(capability.get("local_state_root") == "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001", "fixed Body local state")
check(capability.get("body_local_encounter_record_count") == 0, "no fictional Body-local encounter")
for key in ("autonomous_invocation", "external_network_listener", "may_admit_matter", "may_create_relation_or_contact_event", "may_create_consequence", "may_mutate_currentness"):
    check(capability.get(key) is False, f"capability ceiling: {key}")
exposure = current.get("exposure", {})
check(exposure.get("body_local_encounter_claimed") is False and exposure.get("body_local_encounter_record_count") == 0, "exposure claims no Body-local encounter")
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
surfaces = embodiment.get("surfaces", [])
check(len(surfaces) == 1 and surfaces[0].get("record_count") == 0, "living surface declared empty")
check(surfaces[0].get("records_created_only_on_actual_crossing") is True, "living records require actual crossing")
check(surfaces[0].get("static_manifest_coverage") is False and surfaces[0].get("independent_hash_chain_verification_required") is True, "living state has independent integrity owner")

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
if live_state.exists():
    allowed_names = {"JOURNAL.jsonl", "CURRENT.json", "LOCK", "CARRIERS"}
    check(all(part.name in allowed_names or "CARRIERS" in part.parts for part in live_state.rglob("*")), "living state contains only mechanism-owned paths")
    state_check = subprocess.run([sys.executable, "-B", str(wrapper), "verify-state"], capture_output=True, text=True)
    check(state_check.returncode == 0, "living state hash chain verifies")
else:
    check(True, "living state absent before first Body-local crossing")
    check(True, "absent living state needs no chain verification")

host_verifier = ROOT / "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py"
host_result = subprocess.run([sys.executable, "-B", str(host_verifier)], capture_output=True, text=True)
check(host_result.returncode == 0 and "22" in host_result.stdout and "PASS" in host_result.stdout.upper(), "Integrity Host unchanged 22-case pressure result")


# Lineage and schemas describe the revision without substituting for exact evidence.
lineage = load_json(ROOT / "LINEAGE/INDEX.json", "Lineage index")
check(lineage.get("represented_currentness_revision") == 5, "lineage projects revision 5")
chains = lineage.get("motion_chains", [])
check(len(chains) == 6 and all(chain.get("retrospectively_regulated") is False for chain in chains), "six non-retroactive motion chains")
chain = chains[-1] if len(chains) == 6 else {}
check(chain.get("classification") == "BODY_LOCAL_ADAPTATION_AND_BOUNDED_OPERATIONAL_ACTIVATION", "membrane lineage classification")
check(chain.get("result") == "MECHANISM_ADMITTED_AND_ACTIVATED_BOUNDED", "membrane lineage result")
referenced_hash(chain.get("candidate", {}), "membrane lineage candidate")
referenced_hash(chain.get("act", {}), "membrane lineage Act")
check(chain.get("occurrence", {}).get("binding") == "CURRENT_STATE_CARRIES_OCCURRENCE_METADATA", "occurrence circularity avoided")
check(chain.get("successor_state", {}).get("binding") == "OCCURRENCE_CARRIES_FINAL_STATE_HASH", "successor circularity avoided")

for relative in (
    "CARRIER/SCHEMAS/matter-index.schema.json", "CARRIER/SCHEMAS/regulation-index.schema.json",
    "CARRIER/SCHEMAS/consequence-index.schema.json", "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json", "CARRIER/SCHEMAS/epistemic-index.schema.json",
    "CARRIER/SCHEMAS/lineage-index.schema.json",
):
    schema = load_json(ROOT / relative, relative)
    check(schema.get("properties", {}).get("represented_currentness_revision", {}).get("const") == 5, f"schema projects revision 5: {relative}")
current_schema = load_json(ROOT / "CARRIER/SCHEMAS/current-state.schema.json", "current-state schema")
check(current_schema.get("$id") == "local-reality-body.current-state.v0.5", "current-state schema version")
check(current_schema.get("properties", {}).get("currentness_revision", {}).get("const") == 5, "current-state schema revision")
check({"reality_contact_membrane_occurrence", "operational_capability"}.issubset(current_schema.get("required", [])), "current-state schema requires membrane motion and capability")


# The exposed Form remains exact and independent.
external_manifest = EXTERNAL / "MANIFEST.sha256"
check(EXTERNAL.is_dir(), "external Body Form carrier present")
check(external_manifest.is_file() and sha(external_manifest) == EXTERNAL_MANIFEST_SHA, "external Body Form manifest unchanged")
external_entries = parse_manifest(external_manifest, "external Body Form") if external_manifest.is_file() else {}
for relative, expected in external_entries.items():
    check((EXTERNAL / relative).is_file() and sha(EXTERNAL / relative) == expected, f"external Body Form agreement: {relative}")
check((EXTERNAL / "BODY_FORM.json").is_file() and sha(EXTERNAL / "BODY_FORM.json") == BODY_FORM_SHA, "external Body Form payload unchanged")


# Final occurrence binds the already-stable current and lineage hashes.
occurrence_relative = "LINEAGE/OCCURRENCES/REALITY_CONTACT_MEMBRANE_ACTIVATION_OCCURRENCE_001.md"
occurrence_path = ROOT / occurrence_relative
if FINAL:
    occurrence_text = occurrence_path.read_text(encoding="utf-8") if occurrence_path.is_file() else ""
    check("Status: **PERFORMED**" in occurrence_text, "successful occurrence status")
    check(CURRENT_SHA in occurrence_text and LINEAGE_SHA in occurrence_text, "occurrence binds final current and lineage hashes")
    check("PASS_39_OF_39" in occurrence_text and "record_count: `0`" in occurrence_text, "occurrence binds pressure pass and zero encounters")
    check("No prior external crossing was retroactively admitted" in occurrence_text, "occurrence refuses retroactive capture")
else:
    check(not occurrence_path.exists(), "preflight precedes success occurrence")


# Static manifest excludes exactly the living operational root and nothing else.
new_paths = set(MECHANISM_HASHES) | {
    "LINEAGE/CANDIDATES/REALITY_CONTACT_MEMBRANE_ACTIVATION_CANDIDATE_001.md",
    "LINEAGE/ACTS/REALITY_CONTACT_MEMBRANE_ACTIVATION_ACT_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_004.json",
}
if FINAL:
    new_paths.add(occurrence_relative)
recovery_prefix = "CARRIER/RECOVERY/CARRIER_REVISION_004"
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
check(set(eligible) == expected_static, f"revision-5 exact authorized static file set ({'final' if FINAL else 'preflight'})")
if FINAL:
    manifest_entries = parse_manifest(manifest_path, "revision 5") if manifest_path.is_file() else {}
    check(set(manifest_entries) == set(eligible), "revision-5 manifest complete static coverage")
    check(manifest_entries == eligible, "revision-5 manifest byte agreement")
else:
    check(sha(manifest_path) == REV4_MANIFEST_SHA, "preflight leaves live revision-4 manifest unclaimed")

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

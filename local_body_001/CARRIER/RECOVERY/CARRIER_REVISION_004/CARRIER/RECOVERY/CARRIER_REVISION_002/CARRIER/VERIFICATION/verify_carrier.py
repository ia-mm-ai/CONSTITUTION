#!/usr/bin/env python3
"""Bounded mechanical verification for Local-Reality Body carrier revision 2.

This verifier tests representation, byte identity, references, effect ceilings,
recovery, and manifest agreement. It creates no law, Authority, relation,
occurrence, currentness, or standing.
"""

from __future__ import annotations

import hashlib
import json
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
FINAL = "--final" in sys.argv
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


def load_json(relative: str) -> dict:
    path = ROOT / relative
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:  # bounded reporting, not recovery
        failures.append(f"valid JSON: {relative} ({exc})")
        return {}
    check(isinstance(value, dict), f"JSON object: {relative}")
    return value


def referenced_hash(record: dict, label: str) -> None:
    carrier = record.get("carrier")
    expected = record.get("sha256")
    path = ROOT / carrier if isinstance(carrier, str) else ROOT / "__missing__"
    check(path.is_file(), f"reference exists: {label}")
    check(path.is_file() and sha(path) == expected, f"reference hash: {label}")


def quoted_statement(path: Path) -> str:
    lines = path.read_text(encoding="utf-8").splitlines()
    try:
        start = lines.index("## Exact statement") + 1
    except ValueError:
        return ""
    while start < len(lines) and not lines[start].startswith("> "):
        start += 1
    quoted: list[str] = []
    while start < len(lines) and lines[start].startswith("> "):
        quoted.append(lines[start][2:])
        start += 1
    return "\n".join(quoted)


required_base = [
    "BODY.md",
    "CORE/IDENTITY.md",
    "CORE/LAW.md",
    "CORE/LOCUS.md",
    "CORE/APPLICABILITY.json",
    "RELATIONS/INDEX.json",
    "EMBODIMENT/STATE.json",
    "EPISTEMIC/BOUNDARY.md",
    "EPISTEMIC/INDEX.json",
    "LINEAGE/CANDIDATES/FIRST_LOCAL_REALITY_BODY_CANDIDATE_001.md",
    "LINEAGE/CANDIDATES/BODY_INTEGRITY_HOST_ADMISSION_CANDIDATE_001.md",
    "LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_CARRIER_TOPOLOGY_CANDIDATE_001.md",
    "LINEAGE/ACTS/FORMATION_ACT_001.md",
    "LINEAGE/ACTS/BODY_INTEGRITY_HOST_ADMISSION_ACT_001.md",
    "LINEAGE/ACTS/CARRIER_TOPOLOGY_ACT_001.md",
    "LINEAGE/OCCURRENCES/FORMATION_OCCURRENCE_001.md",
    "LINEAGE/OCCURRENCES/BODY_INTEGRITY_HOST_ADMISSION_OCCURRENCE_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_000.json",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_001.json",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/body_integrity_host.py",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py",
    "CARRIER/SCHEMAS/applicability.schema.json",
    "CARRIER/SCHEMAS/current-state.schema.json",
    "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json",
    "CARRIER/SCHEMAS/epistemic-index.schema.json",
    "CARRIER/VERIFICATION/STATE_REVISION_002_CANDIDATE.json",
    "CARRIER/VERIFICATION/verify_carrier.py",
    "CARRIER/RECOVERY/CARRIER_REVISION_001/SYSTEM/MANIFEST.sha256",
]
for relative in required_base:
    check((ROOT / relative).is_file(), f"required carrier: {relative}")


expected_hashes = {
    "LINEAGE/CANDIDATES/FIRST_LOCAL_REALITY_BODY_CANDIDATE_001.md": "f61959c964b768110708d7ca98b28f87b16ea1e496ffd926cfa177b9dff3310d",
    "LINEAGE/CANDIDATES/BODY_INTEGRITY_HOST_ADMISSION_CANDIDATE_001.md": "3751edb1019a7164f1c563a7a3b95df30fd3b2be3e65de6770d7aaad4a8c68d4",
    "LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_CARRIER_TOPOLOGY_CANDIDATE_001.md": "393e00f2292a793fdd64f22379c3d69d22a126f9b17a3db7c3562fa683d498da",
    "LINEAGE/OCCURRENCES/FORMATION_OCCURRENCE_001.md": "eb419df3af209d1684c8935747b9e9f5f77c33babc978bfec6c49f84b9fb1da6",
    "LINEAGE/OCCURRENCES/BODY_INTEGRITY_HOST_ADMISSION_OCCURRENCE_001.md": "c61b3a710b2cc5be89443602bfa913855c28c0fae2a0ce7b1534a1348b74eaf9",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_000.json": "2a7e7d51aea65ba37d89a46e9b0d78d737d3e4143b68609cb7171e3763ef58f7",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_001.json": "5e7c1f17430a626e5c49db8cfafc345882bd96c58800b2944747b526d4042356",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/body_integrity_host.py": "4f0ae922a9bd8a593b9be05b1e9b37fee1c41092f4c893781ea37096c97f7490",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py": "902bae93d0fa2711f4cf626003f5c0dbdff7539ac12644b64d1046e23dd17d64",
}
for relative, expected in expected_hashes.items():
    check(sha(ROOT / relative) == expected, f"frozen/current lineage hash: {relative}")


recovery = ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_001"
recovery_manifest = recovery / "SYSTEM/MANIFEST.sha256"
check(sha(recovery_manifest) == "f6fd722f82452b26dd8f12d4d4a7bc5420ee204d1cebe62273df94fdf80f547c", "revision-1 manifest byte identity")
manifest_entries: dict[str, str] = {}
for line in recovery_manifest.read_text(encoding="utf-8").splitlines():
    digest, relative = line.split("  ", 1)
    manifest_entries[relative] = digest
for relative, expected in sorted(manifest_entries.items()):
    path = recovery / relative
    check(path.is_file(), f"revision-1 recovery exists: {relative}")
    check(path.is_file() and sha(path) == expected, f"revision-1 recovery hash: {relative}")
recovered_files = {
    path.relative_to(recovery).as_posix()
    for path in recovery.rglob("*")
    if path.is_file() and path.name != ".DS_Store"
}
check(recovered_files == set(manifest_entries) | {"SYSTEM/MANIFEST.sha256"}, "complete revision-1 recovery file set")


law_text = (ROOT / "CORE/LAW.md").read_text(encoding="utf-8")
standing_commitments = [
    "**Integrity of kind.** No relation may obtain the force of another merely through convenience, success, custody, proximity, chronology, confidence, or repetition. Source, Matter, evidence, interpretation, contribution, capability, Authority, action, occurrence, consequence, state, and standing remain distinguishable wherever engaged.",
    "**Integrity of boundary.** The Body may bind only the exact Locus, Matter, conduct, relations, and Authority placed under it through supported local basis. Causal power, storage, access, authorship, origination, operational centrality, or external law cannot silently become Body-local legitimacy or unlimited jurisdiction.",
    "**Integrity through time.** Occurrence, present state, correction, withdrawal, cessation, and later interpretation remain attributable. No correction, replacement, revocation, migration, dormancy, recovery, or preferred current view may silently rewrite what occurred or fabricate uninterrupted continuity.",
    "**Relational answerability.** Consequential motion, refusal, non-progression, correction, adaptation, and external crossing must remain attributable to an exact source, function, scope, target, basis, effect, and uncertainty. No participant, carrier, host, interface, or shared field may answer for another locality.",
]
for commitment in standing_commitments:
    check(law_text.count(commitment) == 1, f"exact standing commitment: {commitment.split('.')[0].strip('*')}")
check("would become" not in law_text, "operative law is not conditional candidate grammar")
check("applicability arises only from" in law_text.lower(), "law projection names occurrence-derived applicability")


identity_text = (ROOT / "CORE/IDENTITY.md").read_text(encoding="utf-8")
locus_text = (ROOT / "CORE/LOCUS.md").read_text(encoding="utf-8")
check("RM-LOCAL-REALITY-BODY-001" in identity_text, "formed identity retained")
check("bounded Reality-Model undertaking" in locus_text, "bounded Locus retained")
check("outside unless an exact later body-local occurrence creates a relation" in locus_text.lower(), "positive general Locus boundary")
check("filesystem path" in locus_text and "does not itself change the Locus" in locus_text, "carrier and Locus remain distinct")


applicability = load_json("CORE/APPLICABILITY.json")
check(applicability.get("schema") == "local-reality-body.applicability.v0.1", "applicability schema identity")
check(applicability.get("body_id") == "RM-LOCAL-REALITY-BODY-001", "applicability Body identity")
for key in ("law_floor", "identity_projection", "locus_projection"):
    record = applicability.get(key, {})
    carrier_key = "operative_projection" if key == "law_floor" else "carrier"
    referenced_hash({"carrier": record.get(carrier_key), "sha256": record.get("sha256")}, f"applicability {key}")
check(applicability.get("formation_candidate", {}).get("sha256") == expected_hashes["LINEAGE/CANDIDATES/FIRST_LOCAL_REALITY_BODY_CANDIDATE_001.md"], "applicability formation candidate anchor")
check(applicability.get("formation_occurrence", {}).get("sha256") == expected_hashes["LINEAGE/OCCURRENCES/FORMATION_OCCURRENCE_001.md"], "applicability Formation Occurrence anchor")
check(applicability.get("classification", {}).get("same_body_constitutional_change_authority") is False, "no same-Body constitutional-change Authority")


act_hashes = {
    "LINEAGE/ACTS/FORMATION_ACT_001.md": "2ac21ea8c01b6a4faef206a518daa5c13eaecd78ee4709266cec563d2fe1f9f2",
    "LINEAGE/ACTS/BODY_INTEGRITY_HOST_ADMISSION_ACT_001.md": "2f1009423104a955c398c6ab35535ca6158fd2f78ff7a34ac17c66f37aa38d4a",
    "LINEAGE/ACTS/CARRIER_TOPOLOGY_ACT_001.md": "3b062e2a9a2c1e4e0fc49d6c4c8c99f05d869bb32cc11110f5556dcafda877b1",
}
for relative, expected in act_hashes.items():
    statement = quoted_statement(ROOT / relative)
    check(bool(statement), f"exact Act statement present: {relative}")
    check(hashlib.sha256(statement.encode("utf-8")).hexdigest() == expected, f"canonical Act hash: {relative}")


state_candidate = load_json("CARRIER/VERIFICATION/STATE_REVISION_002_CANDIDATE.json")
check(state_candidate.get("status") == "STAGED_NOT_CURRENT", "preflight state remains explicitly non-current")
check(state_candidate.get("expected_predecessor_revision") == 1, "preflight predecessor revision")
check(state_candidate.get("proposed_currentness_revision") == 2, "preflight proposed revision")
for key in ("identity", "law", "locus", "applicability"):
    referenced_hash(state_candidate.get("core", {}).get(key, {}), f"preflight core {key}")
referenced_hash(state_candidate.get("relations", {}), "preflight relations")
referenced_hash(state_candidate.get("embodiment", {}), "preflight embodiment")
referenced_hash({"carrier": state_candidate.get("epistemic", {}).get("boundary_carrier"), "sha256": state_candidate.get("epistemic", {}).get("boundary_sha256")}, "preflight epistemic boundary")
referenced_hash({"carrier": state_candidate.get("epistemic", {}).get("index_carrier"), "sha256": state_candidate.get("epistemic", {}).get("index_sha256")}, "preflight epistemic index")
check(state_candidate.get("proposed_carrier_topology_occurrence", {}).get("result") == "PENDING_COMPLETE_VERIFICATION", "preflight does not claim occurrence success")


relations = load_json("RELATIONS/INDEX.json")
check(relations.get("represented_currentness_revision") == 2, "relation projection targets revision 2")
check(len(relations.get("human_body_local", [])) == 1, "one preserved human Body-local relation")
check({item.get("function") for item in relations.get("authority_relations", [])} == {
    "LOCAL_GOVERNING_AUTHORITY",
    "OPERATIONAL_AUTHORIZATION",
    "CORRECTION_AUTHORITY",
    "ADAPTATION_AUTHORITY",
    "DORMANCY_AND_CESSATION",
}, "exact continuing Authority functions")
for key in ("inter_body", "relational_fields", "contact_events", "witness_relations", "external_relations"):
    check(relations.get(key) == [], f"no new relation class: {key}")
di_relation = relations.get("di_coagency", [{}])[0]
check(di_relation.get("status") == "ATTRIBUTABLE_BOUNDED_CONTRIBUTION", "DI co-agency remains bounded contribution")
check(all(di_relation.get(key) is False for key in ("persistent_constitutional_identity", "participation", "authority", "consent_claimed", "bearer_continuity_claimed")), "DI non-effects retained")


embodiment = load_json("EMBODIMENT/STATE.json")
mechanisms = embodiment.get("mechanisms", [])
check(len(mechanisms) == 1, "one admitted embodiment mechanism")
mechanism = mechanisms[0] if mechanisms else {}
check(mechanism.get("status") == "ADMITTED_INACTIVE", "Integrity Host remains admitted inactive")
for key in ("implementation", "pressure_verifier"):
    referenced_hash(mechanism.get(key, {}), f"Integrity Host {key}")
for key in ("activation_authorized", "real_body_act_processing_authorized", "real_body_state_mutation_authorized", "direct_state_adapter_present", "persistence_present", "autonomous_invocation_present", "external_update_relation", "body_continuity_requires_mechanism"):
    check(mechanism.get(key) is False, f"Integrity Host effect ceiling: {key}")
check(mechanism.get("replaceable") is True, "Integrity Host remains replaceable")
check(embodiment.get("surfaces") == [], "no embodiment surface invented")


epistemic = load_json("EPISTEMIC/INDEX.json")
for key in ("admitted_records", "source_relations", "observations", "evidence", "inferences", "working_interpretations", "local_consequences", "standing_claims"):
    check(epistemic.get(key) == [], f"epistemic index empty: {key}")
boundary_text = (ROOT / "EPISTEMIC/BOUNDARY.md").read_text(encoding="utf-8").lower()
check("cannot thereby contain or own intelligence" in boundary_text, "epistemic anti-capture boundary")


living_paths = [ROOT / "BODY.md", ROOT / "CORE", ROOT / "CURRENT", ROOT / "RELATIONS", ROOT / "EMBODIMENT", ROOT / "EPISTEMIC"]
living_files: list[Path] = []
for entry in living_paths:
    if entry.is_file():
        living_files.append(entry)
    elif entry.is_dir():
        living_files.extend(path for path in entry.rglob("*") if path.is_file())
for path in living_files:
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        continue
    check("upad" not in text.lower(), f"no named foreign negative coupling: {path.relative_to(ROOT)}")
check("currentness revision" not in (ROOT / "BODY.md").read_text(encoding="utf-8").lower(), "BODY.md excludes mutable revision")
check("ADMITTED_INACTIVE" not in (ROOT / "BODY.md").read_text(encoding="utf-8"), "BODY.md excludes mutable mechanism posture")


for relative in (
    "CARRIER/SCHEMAS/applicability.schema.json",
    "CARRIER/SCHEMAS/current-state.schema.json",
    "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json",
    "CARRIER/SCHEMAS/epistemic-index.schema.json",
):
    schema = load_json(relative)
    check(schema.get("type") == "object", f"schema object ceiling: {relative}")


host_verifier = ROOT / "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py"
host_result = subprocess.run([sys.executable, str(host_verifier)], cwd=host_verifier.parent, capture_output=True, text=True)
check(host_result.returncode == 0, "Integrity Host pressure verifier returns success")
check("22" in (host_result.stdout + host_result.stderr) and "PASS" in (host_result.stdout + host_result.stderr).upper(), "Integrity Host 22-case pressure result")


if FINAL:
    current_path = ROOT / "CURRENT/STATE.json"
    occurrence_path = ROOT / "LINEAGE/OCCURRENCES/CARRIER_TOPOLOGY_OCCURRENCE_001.md"
    manifest_path = ROOT / "CARRIER/MANIFEST.sha256"
    check(current_path.is_file(), "final current state exists")
    check(occurrence_path.is_file(), "final Carrier Topology Occurrence exists")
    check(manifest_path.is_file(), "final manifest exists")
    current = load_json("CURRENT/STATE.json")
    check(current.get("schema") == "local-reality-body.current-state.v0.2", "final current schema identity")
    check(current.get("body_id") == "RM-LOCAL-REALITY-BODY-001", "final current Body identity")
    check(current.get("currentness_revision") == 2, "final currentness revision 2")
    check(current.get("carrier_topology_occurrence", {}).get("result") == "CORRECTED_AND_ADAPTED", "final current result")
    for key in ("identity", "law", "locus", "applicability"):
        referenced_hash(current.get("core", {}).get(key, {}), f"final core {key}")
    referenced_hash(current.get("relations", {}), "final relations")
    referenced_hash(current.get("embodiment", {}), "final embodiment")
    referenced_hash({"carrier": current.get("epistemic", {}).get("boundary_carrier"), "sha256": current.get("epistemic", {}).get("boundary_sha256")}, "final epistemic boundary")
    referenced_hash({"carrier": current.get("epistemic", {}).get("index_carrier"), "sha256": current.get("epistemic", {}).get("index_sha256")}, "final epistemic index")
    occurrence_text = occurrence_path.read_text(encoding="utf-8") if occurrence_path.is_file() else ""
    check("RESULT: CORRECTED_AND_ADAPTED" in occurrence_text, "performed occurrence result recorded")
    check(sha(current_path) in occurrence_text, "occurrence binds final current-state hash")
    check("CORRECTION_AUTHORITY" in occurrence_text and "ADAPTATION_AUTHORITY" in occurrence_text, "occurrence separates Authority functions")
    parsed_manifest: dict[str, str] = {}
    if manifest_path.is_file():
        for line in manifest_path.read_text(encoding="utf-8").splitlines():
            digest, relative = line.split("  ", 1)
            parsed_manifest[relative] = digest
    eligible = {
        path.relative_to(ROOT).as_posix(): sha(path)
        for path in ROOT.rglob("*")
        if path.is_file() and path != manifest_path and path.name != ".DS_Store"
    }
    check(set(parsed_manifest) == set(eligible), "final manifest complete eligible-file coverage")
    check(parsed_manifest == eligible, "final manifest byte agreement")
else:
    check(not (ROOT / "CURRENT/STATE.json").exists(), "preflight has not advanced currentness")
    check(not (ROOT / "LINEAGE/OCCURRENCES/CARRIER_TOPOLOGY_OCCURRENCE_001.md").exists(), "preflight has no success occurrence")


if failures:
    print(f"\nRESULT  FAIL ({len(failures)} failures across {checks} checks)")
    for failure in failures:
        print(f"- {failure}")
    raise SystemExit(1)

mode = "FINAL" if FINAL else "PREFLIGHT"
print(f"\nRESULT  PASS {mode} ({checks} checks)")

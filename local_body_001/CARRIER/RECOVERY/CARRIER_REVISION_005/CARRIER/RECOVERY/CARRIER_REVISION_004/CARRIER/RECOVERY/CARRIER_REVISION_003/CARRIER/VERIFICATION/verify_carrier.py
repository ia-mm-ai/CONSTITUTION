#!/usr/bin/env python3
"""Fail-closed carrier verification for Local-Reality Body revision 3.

This verifier checks byte identity, recovery, semantic ownership, empty initial
indexes, effect ceilings, references, and final manifest agreement. It cannot
create receipt, Matter, regulation, Authority, an Act, occurrence, consequence,
currentness, standing, or any other Body effect.
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
    except Exception as exc:
        failures.append(f"valid JSON: {relative} ({exc})")
        return {}
    check(isinstance(value, dict), f"JSON object: {relative}")
    return value


def referenced_hash(record: dict, label: str) -> None:
    carrier = record.get("carrier")
    expected = record.get("sha256")
    path = ROOT / carrier if isinstance(carrier, str) else ROOT / "__missing__"
    check(path.is_file(), f"reference exists: {label}")
    check(path.is_file() and isinstance(expected, str) and sha(path) == expected, f"reference hash: {label}")


def boundary_reference(record: dict, prefix: str) -> dict:
    return {"carrier": record.get(f"{prefix}_carrier"), "sha256": record.get(f"{prefix}_sha256")}


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


required = [
    "BODY.md",
    "CORE/IDENTITY.md",
    "CORE/LAW.md",
    "CORE/LOCUS.md",
    "CORE/APPLICABILITY.json",
    "CURRENT/STATE.json",
    "LINEAGE/INDEX.json",
    "LINEAGE/CANDIDATES/FIRST_LOCAL_REALITY_BODY_CANDIDATE_001.md",
    "LINEAGE/CANDIDATES/BODY_INTEGRITY_HOST_ADMISSION_CANDIDATE_001.md",
    "LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_CARRIER_TOPOLOGY_CANDIDATE_001.md",
    "LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_REGULATED_MOTION_TOPOLOGY_CANDIDATE_001.md",
    "LINEAGE/ACTS/FORMATION_ACT_001.md",
    "LINEAGE/ACTS/BODY_INTEGRITY_HOST_ADMISSION_ACT_001.md",
    "LINEAGE/ACTS/CARRIER_TOPOLOGY_ACT_001.md",
    "LINEAGE/ACTS/REGULATED_MOTION_TOPOLOGY_ACT_001.md",
    "LINEAGE/OCCURRENCES/FORMATION_OCCURRENCE_001.md",
    "LINEAGE/OCCURRENCES/BODY_INTEGRITY_HOST_ADMISSION_OCCURRENCE_001.md",
    "LINEAGE/OCCURRENCES/CARRIER_TOPOLOGY_OCCURRENCE_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_000.json",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_001.json",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_002.json",
    "MATTER/BOUNDARY.md",
    "MATTER/INDEX.json",
    "REGULATION/MEDIUM.md",
    "REGULATION/INDEX.json",
    "CONSEQUENCES/BOUNDARY.md",
    "CONSEQUENCES/INDEX.json",
    "RELATIONS/INDEX.json",
    "EMBODIMENT/STATE.json",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/body_integrity_host.py",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py",
    "EPISTEMIC/BOUNDARY.md",
    "EPISTEMIC/INDEX.json",
    "CARRIER/SCHEMAS/applicability.schema.json",
    "CARRIER/SCHEMAS/current-state.schema.json",
    "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json",
    "CARRIER/SCHEMAS/epistemic-index.schema.json",
    "CARRIER/SCHEMAS/matter-index.schema.json",
    "CARRIER/SCHEMAS/regulation-index.schema.json",
    "CARRIER/SCHEMAS/consequence-index.schema.json",
    "CARRIER/SCHEMAS/lineage-index.schema.json",
    "CARRIER/VERIFICATION/STATE_REVISION_002_CANDIDATE.json",
    "CARRIER/VERIFICATION/STATE_REVISION_003_CANDIDATE.json",
    "CARRIER/VERIFICATION/LINEAGE_INDEX_REVISION_003_CANDIDATE.json",
    "CARRIER/VERIFICATION/verify_carrier.py",
    "CARRIER/RECOVERY/CARRIER_REVISION_001/SYSTEM/MANIFEST.sha256",
    "CARRIER/RECOVERY/CARRIER_REVISION_002/CARRIER/MANIFEST.sha256",
]
for relative in required:
    check((ROOT / relative).is_file(), f"required carrier: {relative}")


frozen_hashes = {
    "CORE/IDENTITY.md": "47dfa944d02c3668b5385eab078653751f88bda9b847632f9427911639bf94d9",
    "CORE/LAW.md": "3ae0f66a687e054ff56859e8be70e7001a60b48eeeeefc0909523d5881cf3e06",
    "CORE/LOCUS.md": "8a674f1ff787b0d19cac2e732e1b22d79991ef94a4ddc872181b2693d1650d0d",
    "CORE/APPLICABILITY.json": "8fc46922aa0b3dcd174f4dcec1ffc616df928ccffa9bb59abc611ddeae867aa6",
    "LINEAGE/CANDIDATES/FIRST_LOCAL_REALITY_BODY_CANDIDATE_001.md": "f61959c964b768110708d7ca98b28f87b16ea1e496ffd926cfa177b9dff3310d",
    "LINEAGE/CANDIDATES/BODY_INTEGRITY_HOST_ADMISSION_CANDIDATE_001.md": "3751edb1019a7164f1c563a7a3b95df30fd3b2be3e65de6770d7aaad4a8c68d4",
    "LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_CARRIER_TOPOLOGY_CANDIDATE_001.md": "393e00f2292a793fdd64f22379c3d69d22a126f9b17a3db7c3562fa683d498da",
    "LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_REGULATED_MOTION_TOPOLOGY_CANDIDATE_001.md": "c197415d84c271f1bb05ba0b003c954b490a4c870697353103061906b0740564",
    "LINEAGE/ACTS/FORMATION_ACT_001.md": "67ba6d092d9343f9e62e94b0e332b1ff36cd08d1536cda7b6bea33869a2cb9d1",
    "LINEAGE/ACTS/BODY_INTEGRITY_HOST_ADMISSION_ACT_001.md": "5a4758374fedf23ce14a5cb35ce908c26fe0a55fd4e3c3a968ca76531c346b42",
    "LINEAGE/ACTS/CARRIER_TOPOLOGY_ACT_001.md": "947cde3815d1c08bee3d522cf90e907f475bd5a904a49c4657850d911b6d8662",
    "LINEAGE/ACTS/REGULATED_MOTION_TOPOLOGY_ACT_001.md": "d070b1f4b0cf83f7e0f82359ecd61a7099f05096f4e3dd8ed3e5bb973167d2f8",
    "LINEAGE/OCCURRENCES/FORMATION_OCCURRENCE_001.md": "eb419df3af209d1684c8935747b9e9f5f77c33babc978bfec6c49f84b9fb1da6",
    "LINEAGE/OCCURRENCES/BODY_INTEGRITY_HOST_ADMISSION_OCCURRENCE_001.md": "c61b3a710b2cc5be89443602bfa913855c28c0fae2a0ce7b1534a1348b74eaf9",
    "LINEAGE/OCCURRENCES/CARRIER_TOPOLOGY_OCCURRENCE_001.md": "5334a76d7d2a30a0b8c8da2f13a92a9408dca9f99929a3fcfcc618f2fe58abb9",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_000.json": "2a7e7d51aea65ba37d89a46e9b0d78d737d3e4143b68609cb7171e3763ef58f7",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_001.json": "5e7c1f17430a626e5c49db8cfafc345882bd96c58800b2944747b526d4042356",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_002.json": "7cec0c9bc5f9c269d9e28176fe0fd569c270880beb6f66b98240618a2955a7ab",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/body_integrity_host.py": "4f0ae922a9bd8a593b9be05b1e9b37fee1c41092f4c893781ea37096c97f7490",
    "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py": "902bae93d0fa2711f4cf626003f5c0dbdff7539ac12644b64d1046e23dd17d64",
}
for relative, expected in frozen_hashes.items():
    path = ROOT / relative
    check(path.is_file() and sha(path) == expected, f"frozen/current byte identity: {relative}")


recovery = ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_002"
recovery_manifest = recovery / "CARRIER/MANIFEST.sha256"
check(sha(recovery_manifest) == "1bbc4f5cb5da4eec731b80168bddb99b922820817efed32752f1be5932ede8b4", "revision-2 manifest byte identity")
recovery_entries: dict[str, str] = {}
for line in recovery_manifest.read_text(encoding="utf-8").splitlines():
    digest, relative = line.split("  ", 1)
    recovery_entries[relative] = digest
for relative, expected in sorted(recovery_entries.items()):
    path = recovery / relative
    check(path.is_file(), f"revision-2 recovery exists: {relative}")
    check(path.is_file() and sha(path) == expected, f"revision-2 recovery hash: {relative}")
recovered_files = {
    path.relative_to(recovery).as_posix()
    for path in recovery.rglob("*")
    if path.is_file() and path.name != ".DS_Store"
}
check(recovered_files == set(recovery_entries) | {"CARRIER/MANIFEST.sha256"}, "complete revision-2 recovery file set")

recovered_verifier = recovery / "CARRIER/VERIFICATION/verify_carrier.py"
recovered_result = subprocess.run([sys.executable, str(recovered_verifier), "--final"], cwd=recovered_verifier.parent, capture_output=True, text=True)
recovered_output = recovered_result.stdout + recovered_result.stderr
check(recovered_result.returncode == 0, "recovered revision-2 verifier passes")
check("RESULT  PASS FINAL (198 checks)" in recovered_output, "recovered revision-2 exact 198-check result")


law_text = (ROOT / "CORE/LAW.md").read_text(encoding="utf-8")
standing_commitments = [
    "**Integrity of kind.** No relation may obtain the force of another merely through convenience, success, custody, proximity, chronology, confidence, or repetition. Source, Matter, evidence, interpretation, contribution, capability, Authority, action, occurrence, consequence, state, and standing remain distinguishable wherever engaged.",
    "**Integrity of boundary.** The Body may bind only the exact Locus, Matter, conduct, relations, and Authority placed under it through supported local basis. Causal power, storage, access, authorship, origination, operational centrality, or external law cannot silently become Body-local legitimacy or unlimited jurisdiction.",
    "**Integrity through time.** Occurrence, present state, correction, withdrawal, cessation, and later interpretation remain attributable. No correction, replacement, revocation, migration, dormancy, recovery, or preferred current view may silently rewrite what occurred or fabricate uninterrupted continuity.",
    "**Relational answerability.** Consequential motion, refusal, non-progression, correction, adaptation, and external crossing must remain attributable to an exact source, function, scope, target, basis, effect, and uncertainty. No participant, carrier, host, interface, or shared field may answer for another locality.",
]
for commitment in standing_commitments:
    check(law_text.count(commitment) == 1, f"exact standing commitment: {commitment.split('.')[0].strip('*')}")
check("would become" not in law_text, "operative law remains non-conditional")
check("applicability arises only from" in law_text.lower(), "law applicability source retained")
check("RM-LOCAL-REALITY-BODY-001" in (ROOT / "CORE/IDENTITY.md").read_text(encoding="utf-8"), "formed identity retained")
check("bounded Reality-Model undertaking" in (ROOT / "CORE/LOCUS.md").read_text(encoding="utf-8"), "bounded Locus retained")


act_hashes = {
    "LINEAGE/ACTS/FORMATION_ACT_001.md": "2ac21ea8c01b6a4faef206a518daa5c13eaecd78ee4709266cec563d2fe1f9f2",
    "LINEAGE/ACTS/BODY_INTEGRITY_HOST_ADMISSION_ACT_001.md": "2f1009423104a955c398c6ab35535ca6158fd2f78ff7a34ac17c66f37aa38d4a",
    "LINEAGE/ACTS/CARRIER_TOPOLOGY_ACT_001.md": "3b062e2a9a2c1e4e0fc49d6c4c8c99f05d869bb32cc11110f5556dcafda877b1",
    "LINEAGE/ACTS/REGULATED_MOTION_TOPOLOGY_ACT_001.md": "76af2ff65f49f0dd4bff265fa4e176812b2b0579b86ab464249329809463f6e8",
}
for relative, expected in act_hashes.items():
    statement = quoted_statement(ROOT / relative)
    check(bool(statement), f"exact Act statement present: {relative}")
    check(hashlib.sha256(statement.encode("utf-8")).hexdigest() == expected, f"canonical Act hash: {relative}")


stage = load_json("CARRIER/VERIFICATION/STATE_REVISION_003_CANDIDATE.json")
check(stage.get("status") == "STAGED_NOT_CURRENT", "preflight state explicitly non-current")
check(stage.get("expected_predecessor_revision") == 2, "preflight predecessor revision 2")
check(stage.get("expected_predecessor_state_sha256") == frozen_hashes["LINEAGE/STATE_REVISIONS/CURRENT_REVISION_002.json"], "preflight predecessor state anchor")
check(stage.get("expected_predecessor_manifest_sha256") == "1bbc4f5cb5da4eec731b80168bddb99b922820817efed32752f1be5932ede8b4", "preflight predecessor manifest anchor")
check(stage.get("proposed_currentness_revision") == 3, "preflight proposed revision 3")
check(stage.get("proposed_initial_form_posture") == "INITIAL_FORM_STABILIZED", "preflight bounded initial-form posture")
for key in ("identity", "law", "locus", "applicability"):
    referenced_hash(stage.get("core", {}).get(key, {}), f"preflight core {key}")
referenced_hash(stage.get("lineage", {}), "preflight lineage snapshot")
for owner, first, second in (
    ("matter", "boundary", "index"),
    ("regulation", "medium", "index"),
    ("consequences", "boundary", "index"),
    ("epistemic", "boundary", "index"),
):
    referenced_hash(boundary_reference(stage.get(owner, {}), first), f"preflight {owner} {first}")
    referenced_hash(boundary_reference(stage.get(owner, {}), second), f"preflight {owner} {second}")
referenced_hash(stage.get("relations", {}), "preflight relations")
referenced_hash(stage.get("embodiment", {}), "preflight embodiment")
referenced_hash(stage.get("candidate", {}), "preflight candidate")
referenced_hash(stage.get("act", {}), "preflight Act")
check(stage.get("act", {}).get("canonical_statement_sha256") == act_hashes["LINEAGE/ACTS/REGULATED_MOTION_TOPOLOGY_ACT_001.md"], "preflight canonical Act anchor")
check(stage.get("proposed_occurrence", {}).get("result") == "PENDING_COMPLETE_VERIFICATION", "preflight does not claim occurrence success")
check(all(value == 0 for value in stage.get("content_counts", {}).values()), "preflight invents no content, relation, or mechanism")


body_text = (ROOT / "BODY.md").read_text(encoding="utf-8")
for route in ("`MATTER/`", "`REGULATION/`", "`CONSEQUENCES/`", "`LINEAGE/`"):
    check(route in body_text, f"BODY route present: {route}")
check("currentness revision" not in body_text.lower(), "BODY excludes mutable revision")
check("ADMITTED_INACTIVE" not in body_text, "BODY excludes mutable mechanism posture")


matter = load_json("MATTER/INDEX.json")
check(matter.get("schema") == "local-reality-body.matter-index.v0.1", "Matter schema identity")
check(matter.get("represented_currentness_revision") == 3, "Matter index targets revision 3")
check(matter.get("records") == [] and matter.get("record_count") == 0, "Matter index truthfully empty")
matter_boundary = (ROOT / "MATTER/BOUNDARY.md").read_text(encoding="utf-8").lower()
for phrase in ("external pressure or source object", "local receipt record", "referenced content", "admitted body-local matter", "standing remain distinct"):
    check(phrase in matter_boundary, f"Matter boundary distinction: {phrase}")


regulation = load_json("REGULATION/INDEX.json")
check(regulation.get("schema") == "local-reality-body.regulation-index.v0.1", "Regulation schema identity")
check(regulation.get("represented_currentness_revision") == 3, "Regulation index targets revision 3")
check(regulation.get("records") == [] and regulation.get("record_count") == 0, "Regulation index truthfully empty")
check(regulation.get("mechanism_present") is False, "no regulation mechanism present")
medium_text = (ROOT / "REGULATION/MEDIUM.md").read_text(encoding="utf-8").lower()
for phrase in ("non-sovereign interval", "subordinate to the already-applicable law", "regulatory posture is multidimensional", "mechanical success cannot create", "non-progression"):
    check(phrase in medium_text, f"Regulation Medium ceiling: {phrase}")

regulation_schema = load_json("CARRIER/SCHEMAS/regulation-index.schema.json")
record_schema = regulation_schema.get("properties", {}).get("records", {}).get("items", {})
required_dimensions = {
    "local_record_posture", "progression_posture", "locus_assessment",
    "law_compatibility_assessment", "evidence_posture", "authority_presentation",
}
check(required_dimensions.issubset(set(record_schema.get("required", []))), "regulation schema requires independent dimensions")
check("status" not in record_schema.get("required", []), "regulation schema rejects mandatory flat lifecycle status")
expected_enums = {
    "local_record_posture": {"RECEIVED_REFERENCE_ONLY", "ADMITTED_BODY_LOCAL_MATTER", "UNRESOLVED"},
    "progression_posture": {"HELD", "NON_PROGRESSING", "READY_FOR_AUTHORITY_CONSIDERATION", "RELEASED", "UNRESOLVED"},
    "locus_assessment": {"UNASSESSED", "WITHIN_SUPPORTED_LOCUS", "OUTSIDE_SUPPORTED_LOCUS", "UNRESOLVED"},
    "law_compatibility_assessment": {"UNASSESSED", "NO_CONFLICT_IDENTIFIED", "INCOMPATIBLE_WITH_APPLICABLE_LAW", "UNRESOLVED"},
    "evidence_posture": {"NONE_ADMITTED", "PARTIAL", "SUFFICIENT_FOR_STATED_ASSESSMENT", "CONFLICTING", "UNRESOLVED"},
    "authority_presentation": {"NOT_PRESENTED", "PRESENTED", "WITHDRAWN", "UNRESOLVED"},
}
for key, expected in expected_enums.items():
    actual = set(record_schema.get("properties", {}).get(key, {}).get("enum", []))
    check(actual == expected, f"regulation vocabulary exact: {key}")


consequences = load_json("CONSEQUENCES/INDEX.json")
check(consequences.get("schema") == "local-reality-body.consequence-index.v0.1", "Consequence schema identity")
check(consequences.get("represented_currentness_revision") == 3, "Consequence index targets revision 3")
check(consequences.get("records") == [] and consequences.get("record_count") == 0, "Consequence index truthfully empty")
consequence_boundary = (ROOT / "CONSEQUENCES/BOUNDARY.md").read_text(encoding="utf-8").lower()
for phrase in ("actual effects or changed conditions", "may occur without authorization", "recording it does not authorize", "belong under `epistemic/`"):
    check(phrase in consequence_boundary, f"Consequence boundary distinction: {phrase}")


epistemic = load_json("EPISTEMIC/INDEX.json")
check(epistemic.get("schema") == "local-reality-body.epistemic-index.v0.2", "Epistemic schema corrected")
check(epistemic.get("represented_currentness_revision") == 3, "Epistemic index targets revision 3")
for key in ("admitted_records", "source_relations", "observations", "evidence", "inferences", "working_interpretations", "consequence_observations", "standing_claims"):
    check(epistemic.get(key) == [], f"epistemic index empty: {key}")
check("local_consequences" not in epistemic, "actual consequence removed from epistemic ownership")
epistemic_boundary = (ROOT / "EPISTEMIC/BOUNDARY.md").read_text(encoding="utf-8").lower()
check("actual consequence belongs under `consequences/`" in epistemic_boundary, "epistemic boundary routes actual consequence")
check("cannot thereby contain or own intelligence" in epistemic_boundary, "epistemic anti-capture retained")


lineage = load_json("LINEAGE/INDEX.json")
check(lineage.get("schema") == "local-reality-body.lineage-index.v0.1", "Lineage index schema identity")
check(lineage.get("represented_currentness_revision") == 3, "Lineage index targets revision 3")
check(lineage.get("authoritative_evidence_remains_in_exact_objects") is True, "lineage index does not replace evidence")
chains = lineage.get("motion_chains", [])
check(len(chains) == 4, "lineage index maps four actual or authorized motions")
check(all(chain.get("retrospectively_regulated") is False for chain in chains), "no motion is retrospectively regulated")
check(chains[0].get("classification") == "PRECONSTITUTIONAL_FORMATION", "formation remains preconstitutional")
for index, chain in enumerate(chains[:3]):
    for key in ("candidate", "act", "occurrence", "successor_state"):
        referenced_hash(chain.get(key, {}), f"lineage chain {index + 1} {key}")
expected_results = ["FORMED", "ADMITTED_INACTIVE", "CORRECTED_AND_ADAPTED"]
check([chain.get("result") for chain in chains[:3]] == expected_results, "historical lineage results exact")


relations = load_json("RELATIONS/INDEX.json")
check(relations.get("represented_currentness_revision") == 3, "relation projection targets revision 3")
check(len(relations.get("human_body_local", [])) == 1, "one preserved human Body-local relation")
check({item.get("function") for item in relations.get("authority_relations", [])} == {
    "LOCAL_GOVERNING_AUTHORITY", "OPERATIONAL_AUTHORIZATION", "CORRECTION_AUTHORITY",
    "ADAPTATION_AUTHORITY", "DORMANCY_AND_CESSATION",
}, "exact continuing Authority functions")
for key in ("inter_body", "relational_fields", "contact_events", "witness_relations", "external_relations"):
    check(relations.get(key) == [], f"no new relation class: {key}")
di_relation = relations.get("di_coagency", [{}])[0]
check(di_relation.get("status") == "ATTRIBUTABLE_BOUNDED_CONTRIBUTION", "DI co-agency remains bounded contribution")
check(all(di_relation.get(key) is False for key in ("persistent_constitutional_identity", "participation", "authority", "consent_claimed", "bearer_continuity_claimed")), "DI non-effects retained")


embodiment = load_json("EMBODIMENT/STATE.json")
mechanisms = embodiment.get("mechanisms", [])
check(embodiment.get("represented_currentness_revision") == 3, "embodiment projection targets revision 3")
check(len(mechanisms) == 1, "no new embodiment mechanism")
mechanism = mechanisms[0] if mechanisms else {}
check(mechanism.get("mechanism_id") == "RM-BODY-INTEGRITY-HOST-001", "only Integrity Host remains present")
check(mechanism.get("status") == "ADMITTED_INACTIVE", "Integrity Host remains admitted inactive")
for key in ("implementation", "pressure_verifier"):
    referenced_hash(mechanism.get(key, {}), f"Integrity Host {key}")
for key in ("activation_authorized", "real_body_act_processing_authorized", "real_body_state_mutation_authorized", "direct_state_adapter_present", "persistence_present", "autonomous_invocation_present", "external_update_relation", "body_continuity_requires_mechanism"):
    check(mechanism.get(key) is False, f"Integrity Host effect ceiling: {key}")
check(mechanism.get("replaceable") is True, "Integrity Host remains replaceable")
check(embodiment.get("surfaces") == [], "no embodiment surface invented")
check(not (ROOT / "EMBODIMENT/MECHANISMS/BODY_REGULATION_MEDIUM_HOST").exists(), "no Regulation Medium Host path")


for relative in (
    "CARRIER/SCHEMAS/applicability.schema.json", "CARRIER/SCHEMAS/current-state.schema.json",
    "CARRIER/SCHEMAS/relation-index.schema.json", "CARRIER/SCHEMAS/embodiment-state.schema.json",
    "CARRIER/SCHEMAS/epistemic-index.schema.json", "CARRIER/SCHEMAS/matter-index.schema.json",
    "CARRIER/SCHEMAS/regulation-index.schema.json", "CARRIER/SCHEMAS/consequence-index.schema.json",
    "CARRIER/SCHEMAS/lineage-index.schema.json",
):
    schema = load_json(relative)
    check(schema.get("type") == "object", f"schema object ceiling: {relative}")

for path in (ROOT / "MATTER/RECORDS", ROOT / "REGULATION/RECORDS", ROOT / "CONSEQUENCES/RECORDS", ROOT / "MOTION"):
    check(not path.exists(), f"no empty or unselected carrier path: {path.relative_to(ROOT)}")

host_verifier = ROOT / "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py"
host_result = subprocess.run([sys.executable, str(host_verifier)], cwd=host_verifier.parent, capture_output=True, text=True)
host_output = host_result.stdout + host_result.stderr
check(host_result.returncode == 0, "Integrity Host pressure verifier returns success")
check("22" in host_output and "PASS" in host_output.upper(), "Integrity Host 22-case pressure result")

living_roots = [ROOT / name for name in ("BODY.md", "CORE", "CURRENT", "MATTER", "REGULATION", "CONSEQUENCES", "RELATIONS", "EMBODIMENT", "EPISTEMIC")]
living_files: list[Path] = []
for entry in living_roots:
    if entry.is_file():
        living_files.append(entry)
    elif entry.is_dir():
        living_files.extend(path for path in entry.rglob("*") if path.is_file())
for path in living_files:
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        continue
    check("upad" not in text.lower(), f"no named foreign contamination: {path.relative_to(ROOT)}")


if FINAL:
    occurrence_path = ROOT / "LINEAGE/OCCURRENCES/REGULATED_MOTION_TOPOLOGY_OCCURRENCE_001.md"
    manifest_path = ROOT / "CARRIER/MANIFEST.sha256"
    check(occurrence_path.is_file(), "final Regulated-Motion Topology Occurrence exists")
    check(manifest_path.is_file(), "final manifest exists")
    current = load_json("CURRENT/STATE.json")
    check(current.get("schema") == "local-reality-body.current-state.v0.3", "final current schema identity")
    check(current.get("body_id") == "RM-LOCAL-REALITY-BODY-001", "final current Body identity")
    check(current.get("currentness_revision") == 3, "final currentness revision 3")
    check(current.get("initial_form_posture") == "INITIAL_FORM_STABILIZED", "initial form stabilized")
    ceiling = current.get("initial_form_posture_ceiling", {})
    check(ceiling.get("descriptive_only") is True, "initial-form posture descriptive only")
    check(all(ceiling.get(key) is False for key in ("perfection_claimed", "final_architecture_claimed", "operational_activation_claimed", "immunity_from_later_change_claimed")), "initial-form non-finality ceiling")
    for key in ("identity", "law", "locus", "applicability"):
        referenced_hash(current.get("core", {}).get(key, {}), f"final core {key}")
    referenced_hash(current.get("lineage", {}), "final lineage")
    for owner, first, second in (
        ("matter", "boundary", "index"),
        ("regulation", "medium", "index"),
        ("consequences", "boundary", "index"),
        ("epistemic", "boundary", "index"),
    ):
        referenced_hash(boundary_reference(current.get(owner, {}), first), f"final {owner} {first}")
        referenced_hash(boundary_reference(current.get(owner, {}), second), f"final {owner} {second}")
    referenced_hash(current.get("relations", {}), "final relations")
    referenced_hash(current.get("embodiment", {}), "final embodiment")
    motion = current.get("regulated_motion_topology_occurrence", {})
    check(motion.get("result") == "CORRECTED_AND_ADAPTED", "final occurrence result")
    check(motion.get("predecessor_revision") == 2, "final predecessor revision")
    check(motion.get("recovery_carrier") == "CARRIER/RECOVERY/CARRIER_REVISION_002", "final recovery route")
    check(motion.get("canonical_act_sha256") == act_hashes["LINEAGE/ACTS/REGULATED_MOTION_TOPOLOGY_ACT_001.md"], "final canonical Act anchor")
    occurrence_text = occurrence_path.read_text(encoding="utf-8") if occurrence_path.is_file() else ""
    check("RESULT: CORRECTED_AND_ADAPTED" in occurrence_text, "performed occurrence result recorded")
    check(sha(ROOT / "CURRENT/STATE.json") in occurrence_text, "occurrence binds final current-state hash")
    check("CORRECTION_AUTHORITY" in occurrence_text and "ADAPTATION_AUTHORITY" in occurrence_text, "occurrence separates Authority functions")
    check("INITIAL_FORM_STABILIZED" in occurrence_text and "not perfection" in occurrence_text.lower(), "occurrence preserves initial-form ceiling")
    current_chain = chains[3] if len(chains) == 4 else {}
    referenced_hash(current_chain.get("candidate", {}), "final lineage candidate")
    referenced_hash(current_chain.get("act", {}), "final lineage Act")
    occurrence_reference = current_chain.get("occurrence", {})
    check(
        occurrence_reference.get("carrier") == "LINEAGE/OCCURRENCES/REGULATED_MOTION_TOPOLOGY_OCCURRENCE_001.md"
        and occurrence_reference.get("binding") == "CURRENT_STATE_CARRIES_OCCURRENCE_METADATA",
        "final lineage maps occurrence without circular hash",
    )
    check(current_chain.get("result") == "CORRECTED_AND_ADAPTED", "final lineage current result")
    successor = current_chain.get("successor_state", {})
    check(successor.get("carrier") == "CURRENT/STATE.json" and successor.get("binding") == "OCCURRENCE_CARRIES_FINAL_STATE_HASH", "final lineage maps successor state without circular hash")
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
    current_path = ROOT / "CURRENT/STATE.json"
    check(sha(current_path) == "7cec0c9bc5f9c269d9e28176fe0fd569c270880beb6f66b98240618a2955a7ab", "live predecessor state remains exact in preflight")
    check(load_json("CURRENT/STATE.json").get("currentness_revision") == 2, "preflight has not advanced currentness")
    check(not (ROOT / "LINEAGE/OCCURRENCES/REGULATED_MOTION_TOPOLOGY_OCCURRENCE_001.md").exists(), "preflight has no success occurrence")
    current_chain = chains[3] if len(chains) == 4 else {}
    check(current_chain.get("occurrence") is None and current_chain.get("successor_state") is None, "preflight lineage claims no result objects")
    check(current_chain.get("result") == "AUTHORIZED_PENDING_COMPLETE_VERIFICATION", "preflight lineage remains pending")


if failures:
    print(f"\nRESULT  FAIL ({len(failures)} failures across {checks} checks)")
    for failure in failures:
        print(f"- {failure}")
    raise SystemExit(1)

mode = "FINAL" if FINAL else "PREFLIGHT"
print(f"\nRESULT  PASS {mode} ({checks} checks)")

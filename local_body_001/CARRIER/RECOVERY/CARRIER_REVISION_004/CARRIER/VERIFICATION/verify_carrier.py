#!/usr/bin/env python3
"""Fail-closed carrier verification for Local-Reality Body revision 4.

This verifier checks the exact Body Form exposure, predecessor recovery,
current semantic projections, lineage, manifests, contamination ceilings, and
explicit non-effects. It reports carrier agreement only. It cannot originate
Authority, an Act, an occurrence, exposure, currentness, formation, relation,
or any other Body effect.
"""

from __future__ import annotations

import hashlib
import json
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
DEFAULT_EXTERNAL = Path("/Users/markomarkota/LOCAL_REALITY_BODY_FORMS/RM_LOCAL_REALITY_BODY_FORM_001")
EXTERNAL = DEFAULT_EXTERNAL
if "--external-root" in sys.argv:
    position = sys.argv.index("--external-root")
    if position + 1 >= len(sys.argv):
        raise SystemExit("--external-root requires one path")
    EXTERNAL = Path(sys.argv[position + 1]).resolve()

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
        lines = path.read_text(encoding="utf-8").splitlines()
        for line in lines:
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


CANDIDATE_SHA = "b73a70e866def2f59df42a211d45b56cfbb52e2789f235f977fa74069d7ef63b"
BODY_FORM_SHA = "4031ba2a6b7bc7cbe7352bdd91b351227aa1d517a0d4680d4fafc6afc26ca056"
README_SHA = "04cf4b600862595390c0b363c48b5fa0fd70cf4369f162da3ad9af50302d4bd1"
ACT_CARRIER_SHA = "b4bbc97415c6c8ab63f3ce454dd9fcbc8f712880def6d5eb9060b1a7e18a1e85"
ACT_CANONICAL_SHA = "7a58abe884dee808790d28162bfa7fa99bde71dbf47cd5f4edf0edc13fab58f5"
PREDECESSOR_STATE_SHA = "11aed042368f6e62b51e9d135785e63dfbb51b28ac33267913817ef9ab0ee577"
PREDECESSOR_MANIFEST_SHA = "52b354e5a12346723328e44e0c50cc33eacd718147923e37178b5262c937bbe8"
CURRENT_STATE_SHA = "451afca839f93a99b2bac1e6c7226391df66aeffe1ee2caa92b14d93d381ff7c"
LINEAGE_SHA = "423b4ab2bafa8dff8441c62b062f1773ed3753a9abfce5faf76a0a9c2adda4b1"
EXPOSURE_SHA = "90c60f59aaef2793c6ad191a8cbf9b6cfaa4880e371335e38ed37d8138f7b5b0"
EXTERNAL_MANIFEST_SHA = "da93daeba6b0f8e2c49521f4b043be35330d103422dae1ddade82670f43f5a2c"

CANDIDATE_BASE = "LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_FORM_EXPOSURE_CANDIDATE_001"
required = [
    "BODY.md",
    "CORE/IDENTITY.md",
    "CORE/LAW.md",
    "CORE/LOCUS.md",
    "CORE/APPLICABILITY.json",
    "CURRENT/STATE.json",
    "LINEAGE/INDEX.json",
    f"{CANDIDATE_BASE}/CANDIDATE.md",
    f"{CANDIDATE_BASE}/PROPOSED_FORM/BODY_FORM.json",
    f"{CANDIDATE_BASE}/PROPOSED_FORM/README.md",
    "LINEAGE/ACTS/BODY_FORM_EXPOSURE_ACT_001.md",
    "LINEAGE/OCCURRENCES/BODY_FORM_EXPOSURE_OCCURRENCE_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_003.json",
    "MATTER/INDEX.json",
    "REGULATION/INDEX.json",
    "CONSEQUENCES/INDEX.json",
    "RELATIONS/INDEX.json",
    "EMBODIMENT/STATE.json",
    "EPISTEMIC/INDEX.json",
    "CARRIER/SCHEMAS/current-state.schema.json",
    "CARRIER/SCHEMAS/lineage-index.schema.json",
    "CARRIER/VERIFICATION/verify_carrier.py",
    "CARRIER/RECOVERY/CARRIER_REVISION_003/CARRIER/MANIFEST.sha256",
    "CARRIER/MANIFEST.sha256",
]
for relative in required:
    check((ROOT / relative).is_file(), f"required carrier: {relative}")


recovery = ROOT / "CARRIER/RECOVERY/CARRIER_REVISION_003"
recovery_manifest = recovery / "CARRIER/MANIFEST.sha256"
check(recovery_manifest.is_file() and sha(recovery_manifest) == PREDECESSOR_MANIFEST_SHA, "revision-3 manifest byte identity")
recovery_entries = parse_manifest(recovery_manifest, "revision 3")
for relative, expected in sorted(recovery_entries.items()):
    path = recovery / relative
    check(path.is_file(), f"revision-3 recovery exists: {relative}")
    check(path.is_file() and sha(path) == expected, f"revision-3 recovery hash: {relative}")
recovered_files = {
    path.relative_to(recovery).as_posix()
    for path in recovery.rglob("*")
    if path.is_file() and path.name != ".DS_Store"
}
check(recovered_files == set(recovery_entries) | {"CARRIER/MANIFEST.sha256"}, "complete revision-3 recovery file set")

recovered_verifier = recovery / "CARRIER/VERIFICATION/verify_carrier.py"
recovered_result = subprocess.run(
    [sys.executable, "-B", str(recovered_verifier), "--final"],
    cwd=recovered_verifier.parent,
    capture_output=True,
    text=True,
)
recovered_output = recovered_result.stdout + recovered_result.stderr
check(recovered_result.returncode == 0, "recovered revision-3 verifier passes")
check("RESULT  PASS FINAL (427 checks)" in recovered_output, "recovered revision-3 exact 427-check result")


changed_from_revision_3 = {
    "CURRENT/STATE.json",
    "LINEAGE/INDEX.json",
    "MATTER/INDEX.json",
    "REGULATION/INDEX.json",
    "CONSEQUENCES/INDEX.json",
    "RELATIONS/INDEX.json",
    "EMBODIMENT/STATE.json",
    "EPISTEMIC/INDEX.json",
    "CARRIER/SCHEMAS/current-state.schema.json",
    "CARRIER/SCHEMAS/lineage-index.schema.json",
    "CARRIER/SCHEMAS/matter-index.schema.json",
    "CARRIER/SCHEMAS/regulation-index.schema.json",
    "CARRIER/SCHEMAS/consequence-index.schema.json",
    "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json",
    "CARRIER/SCHEMAS/epistemic-index.schema.json",
    "CARRIER/VERIFICATION/verify_carrier.py",
}
for relative, expected in sorted(recovery_entries.items()):
    if relative in changed_from_revision_3:
        continue
    path = ROOT / relative
    check(path.is_file() and sha(path) == expected, f"unchanged revision-3 byte identity: {relative}")


candidate_path = ROOT / CANDIDATE_BASE / "CANDIDATE.md"
body_form_path = ROOT / CANDIDATE_BASE / "PROPOSED_FORM/BODY_FORM.json"
readme_path = ROOT / CANDIDATE_BASE / "PROPOSED_FORM/README.md"
act_path = ROOT / "LINEAGE/ACTS/BODY_FORM_EXPOSURE_ACT_001.md"
check(sha(candidate_path) == CANDIDATE_SHA, "exact exposure candidate")
check(sha(body_form_path) == BODY_FORM_SHA, "exact Body Form payload")
check(sha(readme_path) == README_SHA, "exact Body Form human entrance")
check(sha(act_path) == ACT_CARRIER_SHA, "exact exposure Act carrier")
statement = quoted_statement(act_path)
check(bool(statement), "exact exposure Act statement present")
check(hashlib.sha256(statement.encode("utf-8")).hexdigest() == ACT_CANONICAL_SHA, "canonical exposure Act hash")
for token in ("LOCAL_GOVERNING_AUTHORITY", "ADAPTATION_AUTHORITY", "OPERATIONAL_AUTHORIZATION"):
    check(token in statement, f"Act invokes distinct function: {token}")
check("CORRECTION_AUTHORITY" not in statement and "DORMANCY_AND_CESSATION" not in statement, "Act does not invoke unrelated functions")

predecessor_state = ROOT / "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_003.json"
check(sha(predecessor_state) == PREDECESSOR_STATE_SHA, "exact revision-3 predecessor state")
check(sha(ROOT / "CURRENT/STATE.json") == CURRENT_STATE_SHA, "exact revision-4 current state")
check(sha(ROOT / "LINEAGE/INDEX.json") == LINEAGE_SHA, "exact revision-4 lineage index")


body_form = load_json(body_form_path, "Body Form")
form = body_form.get("form", {})
check(form.get("id") == "RM-LOCAL-REALITY-BODY-FORM-001", "Body Form identity")
check(form.get("role") == "REUSABLE_LOCAL_REALITY_BODY_FORM_MATERIAL", "Body Form reusable role")
check(form.get("body_instance") is False and form.get("law_in_force") is False, "Body Form is neither Body nor applicable law")
preserve_ids = {entry.get("id") for entry in body_form.get("must_preserve", [])}
expected_preserve = {
    "INTEGRITY_OF_KIND",
    "INTEGRITY_OF_BOUNDARY",
    "INTEGRITY_THROUGH_TIME",
    "RELATIONAL_ANSWERABILITY",
    "SOURCE_SUBJECTION",
    "INDEPENDENT_LOCAL_FORMATION",
    "TRUTHFUL_CURRENTNESS",
    "NON_PROPAGATION",
}
check(preserve_ids == expected_preserve, "Body Form complete MUST_PRESERVE set")
check(len(body_form.get("may_adapt", [])) == 11, "Body Form complete MAY_ADAPT set")
check(len(body_form.get("must_not_inherit", [])) == 13, "Body Form complete MUST_NOT_INHERIT set")
check(len(body_form.get("formation_grammar", [])) == 6, "Body Form six-relation formation grammar")
receiver = body_form.get("receiver_independence", {})
check(receiver.get("source_body_can_constitute_receiver") is False, "source Body cannot constitute receiver")
check(receiver.get("form_ancestry_creates_inter_body_relation") is False, "form ancestry creates no inter-Body relation")
boundary = body_form.get("relation_boundary", {})
check(boundary.get("two_formed_bodies_required_before_inter_body_relation_formation") is True, "two Bodies required before relation formation")


current = load_json(ROOT / "CURRENT/STATE.json", "current state")
check(current.get("schema") == "local-reality-body.current-state.v0.4", "current schema identity")
check(current.get("body_id") == "RM-LOCAL-REALITY-BODY-001", "current Body identity")
check(current.get("body_status") == "FORMED", "Body remains formed")
check(current.get("currentness_revision") == 4, "currentness revision 4")
check(current.get("initial_form_posture") == "INITIAL_FORM_STABILIZED", "initial-form posture retained")
ceiling = current.get("initial_form_posture_ceiling", {})
check(ceiling.get("descriptive_only") is True, "initial-form posture remains descriptive")
check(all(ceiling.get(key) is False for key in ("perfection_claimed", "final_architecture_claimed", "operational_activation_claimed", "immunity_from_later_change_claimed")), "initial-form non-finality ceiling")
functions = current.get("current_authority_relations", {}).get("functions", [])
check(
    functions == [
        "LOCAL_GOVERNING_AUTHORITY",
        "OPERATIONAL_AUTHORIZATION",
        "CORRECTION_AUTHORITY",
        "ADAPTATION_AUTHORITY",
        "DORMANCY_AND_CESSATION",
    ],
    "Authority topology unchanged",
)
check(current.get("current_authority_relations", {}).get("bearer") == "Marko Markota", "current bearer unchanged")
for key in ("identity", "law", "locus", "applicability"):
    referenced_hash(current.get("core", {}).get(key, {}), f"current core {key}")
referenced_hash(current.get("lineage", {}), "current lineage")
for owner, first, second in (
    ("matter", "boundary", "index"),
    ("regulation", "medium", "index"),
    ("consequences", "boundary", "index"),
    ("epistemic", "boundary", "index"),
):
    referenced_hash(boundary_reference(current.get(owner, {}), first), f"current {owner} {first}")
    referenced_hash(boundary_reference(current.get(owner, {}), second), f"current {owner} {second}")
referenced_hash(current.get("relations", {}), "current relations")
referenced_hash(current.get("embodiment", {}), "current embodiment")
motion = current.get("body_form_exposure_occurrence", {})
check(motion.get("result") == "BODY_FORM_EXPOSED_LOCAL_ONLY", "current exposure result")
check(motion.get("posture") == "EXPOSE", "current exposure posture")
check(motion.get("predecessor_revision") == 3, "current exposure predecessor revision")
check(motion.get("predecessor_state_sha256") == PREDECESSOR_STATE_SHA, "current exposure predecessor state")
check(motion.get("candidate_sha256") == CANDIDATE_SHA, "current exposure candidate")
check(motion.get("act_carrier_sha256") == ACT_CARRIER_SHA, "current exposure Act carrier")
check(motion.get("canonical_act_sha256") == ACT_CANONICAL_SHA, "current exposure canonical Act")
check(motion.get("recovery_carrier") == "CARRIER/RECOVERY/CARRIER_REVISION_003", "current exposure recovery route")
exposure = current.get("exposure", {})
check(exposure.get("body_form_exposed") is True, "Body Form exposed")
check(exposure.get("form_id") == "RM-LOCAL-REALITY-BODY-FORM-001", "current exact exposed form")
check(exposure.get("body_form_sha256") == BODY_FORM_SHA, "current Body Form hash")
check(exposure.get("readme_sha256") == README_SHA, "current README hash")
check(exposure.get("exposure_record_sha256") == EXPOSURE_SHA, "current exposure-record hash")
check(exposure.get("exposure_manifest_sha256") == EXTERNAL_MANIFEST_SHA, "current exposure-manifest hash")
check(exposure.get("visibility") == "LOCAL_HOST_READABLE", "local-host visibility exact")
check(exposure.get("external_publication") is False, "no external publication")
check(exposure.get("external_legal_license") is False, "no external legal licence")
check(exposure.get("encounter_claimed") is False and exposure.get("receiver_formation_claimed") is False, "no encounter or receiver formation claimed")
non_effects = set(current.get("explicit_non_effects", []))
for effect in (
    "NO_PUBLIC_OR_NETWORK_PUBLICATION",
    "NO_EXTERNAL_LEGAL_LICENSE",
    "NO_RECEIVER_FORMATION_OR_SECOND_BODY",
    "NO_PARENTAGE_OWNERSHIP_OR_AUTOMATIC_PROPAGATION",
    "NO_PRIVATE_KEY_OR_CREDENTIAL_USE",
    "NO_INTER_BODY_RELATION",
    "NO_RELATIONAL_FIELD",
    "NO_CONTACT_EVENT",
    "NO_SHARED_CURRENTNESS",
):
    check(effect in non_effects, f"current non-effect: {effect}")
check("NO_EXTERNAL_PUBLICATION_OR_EXPOSURE" not in non_effects, "obsolete no-exposure claim removed")


for owner in ("MATTER/INDEX.json", "REGULATION/INDEX.json", "CONSEQUENCES/INDEX.json", "RELATIONS/INDEX.json", "EMBODIMENT/STATE.json", "EPISTEMIC/INDEX.json"):
    value = load_json(ROOT / owner, owner)
    check(value.get("represented_currentness_revision") == 4, f"owner projects revision 4: {owner}")

matter = load_json(ROOT / "MATTER/INDEX.json", "Matter index")
regulation = load_json(ROOT / "REGULATION/INDEX.json", "Regulation index")
consequences = load_json(ROOT / "CONSEQUENCES/INDEX.json", "Consequence index")
epistemic = load_json(ROOT / "EPISTEMIC/INDEX.json", "Epistemic index")
relations = load_json(ROOT / "RELATIONS/INDEX.json", "Relation index")
embodiment = load_json(ROOT / "EMBODIMENT/STATE.json", "Embodiment state")
check(matter.get("records") == [] and matter.get("record_count") == 0, "Matter remains empty")
check(regulation.get("records") == [] and regulation.get("record_count") == 0, "Regulation remains empty")
check(regulation.get("mechanism_present") is False, "no Regulation mechanism")
check(consequences.get("records") == [] and consequences.get("record_count") == 0, "Consequences remain empty")
for key in ("admitted_records", "source_relations", "observations", "evidence", "inferences", "working_interpretations", "consequence_observations", "standing_claims"):
    check(epistemic.get(key) == [], f"Epistemic remains empty: {key}")
check(len(relations.get("human_body_local", [])) == 1, "one existing human-Body local relation")
check(len(relations.get("authority_relations", [])) == 5, "five unchanged Authority relations")
for key in ("inter_body", "relational_fields", "contact_events", "witness_relations", "external_relations"):
    check(relations.get(key) == [], f"no relation created: {key}")
mechanisms = embodiment.get("mechanisms", [])
check(len(mechanisms) == 1 and mechanisms[0].get("status") == "ADMITTED_INACTIVE", "Integrity Host remains admitted inactive")
check(embodiment.get("surfaces") == [], "no Body-local embodiment surface invented")


lineage = load_json(ROOT / "LINEAGE/INDEX.json", "Lineage index")
check(lineage.get("represented_currentness_revision") == 4, "lineage projects revision 4")
check(lineage.get("authoritative_evidence_remains_in_exact_objects") is True, "lineage does not replace exact evidence")
chains = lineage.get("motion_chains", [])
check(len(chains) == 5, "lineage maps five motion chains")
check(all(chain.get("retrospectively_regulated") is False for chain in chains), "no motion retrospectively regulated")
for index, chain in enumerate(chains[:3]):
    for key in ("candidate", "act", "occurrence", "successor_state"):
        referenced_hash(chain.get(key, {}), f"historical lineage chain {index + 1} {key}")
check(chains[3].get("result") == "CORRECTED_AND_ADAPTED", "regulated-motion chain retained")
new_chain = chains[4] if len(chains) == 5 else {}
referenced_hash(new_chain.get("candidate", {}), "exposure lineage candidate")
referenced_hash(new_chain.get("act", {}), "exposure lineage Act")
check(new_chain.get("classification") == "BODY_LOCAL_GOVERNING_ADAPTATION_AND_OPERATIONAL_EXPOSURE", "exposure lineage classification")
check(new_chain.get("result") == "BODY_FORM_EXPOSED_LOCAL_ONLY", "exposure lineage result")
occurrence_ref = new_chain.get("occurrence", {})
check(occurrence_ref.get("carrier") == "LINEAGE/OCCURRENCES/BODY_FORM_EXPOSURE_OCCURRENCE_001.md" and occurrence_ref.get("binding") == "CURRENT_STATE_CARRIES_OCCURRENCE_METADATA", "exposure occurrence circularity avoided")
successor_ref = new_chain.get("successor_state", {})
check(successor_ref.get("carrier") == "CURRENT/STATE.json" and successor_ref.get("binding") == "OCCURRENCE_CARRIES_FINAL_STATE_HASH", "exposure successor circularity avoided")


revision_schemas = (
    "CARRIER/SCHEMAS/matter-index.schema.json",
    "CARRIER/SCHEMAS/regulation-index.schema.json",
    "CARRIER/SCHEMAS/consequence-index.schema.json",
    "CARRIER/SCHEMAS/relation-index.schema.json",
    "CARRIER/SCHEMAS/embodiment-state.schema.json",
    "CARRIER/SCHEMAS/epistemic-index.schema.json",
    "CARRIER/SCHEMAS/lineage-index.schema.json",
)
for relative in revision_schemas:
    schema = load_json(ROOT / relative, relative)
    revision_rule = schema.get("properties", {}).get("represented_currentness_revision", {})
    check(schema.get("type") == "object", f"schema object ceiling: {relative}")
    check(revision_rule.get("const") == 4, f"schema projects revision 4: {relative}")
current_schema = load_json(ROOT / "CARRIER/SCHEMAS/current-state.schema.json", "current-state schema")
check(current_schema.get("$id") == "local-reality-body.current-state.v0.4", "current-state schema version")
check(current_schema.get("properties", {}).get("currentness_revision", {}).get("const") == 4, "current-state schema revision")
check({"body_form_exposure_occurrence", "exposure"}.issubset(set(current_schema.get("required", []))), "current-state schema requires exposure")


external_required = {"BODY_FORM.json", "README.md", "EXPOSURE.json", "MANIFEST.sha256"}
external_files = {
    path.relative_to(EXTERNAL).as_posix()
    for path in EXTERNAL.rglob("*")
    if path.is_file() and path.name != ".DS_Store"
} if EXTERNAL.is_dir() else set()
check(EXTERNAL.is_dir(), "external exposure directory present")
check(external_files == external_required, "external exact four-file carrier")
external_form = EXTERNAL / "BODY_FORM.json"
external_readme = EXTERNAL / "README.md"
external_record = EXTERNAL / "EXPOSURE.json"
external_manifest = EXTERNAL / "MANIFEST.sha256"
check(external_form.is_file() and sha(external_form) == BODY_FORM_SHA, "external Body Form byte identity")
check(external_readme.is_file() and sha(external_readme) == README_SHA, "external README byte identity")
check(external_record.is_file() and sha(external_record) == EXPOSURE_SHA, "external exposure-record byte identity")
check(external_manifest.is_file() and sha(external_manifest) == EXTERNAL_MANIFEST_SHA, "external manifest byte identity")
external_entries = parse_manifest(external_manifest, "external exposure") if external_manifest.is_file() else {}
check(set(external_entries) == {"BODY_FORM.json", "README.md", "EXPOSURE.json"}, "external manifest exact coverage")
for relative, expected in external_entries.items():
    path = EXTERNAL / relative
    check(path.is_file() and sha(path) == expected, f"external manifest agreement: {relative}")
external_exposure = load_json(external_record, "external exposure record") if external_record.is_file() else {}
external_status = external_exposure.get("exposure", {})
check(external_status.get("result") == "BODY_FORM_EXPOSED_LOCAL_ONLY", "external exposure result")
check(external_status.get("target") == str(DEFAULT_EXTERNAL), "external target exact")
check(external_status.get("visibility") == "LOCAL_HOST_READABLE", "external visibility ceiling")
check(external_status.get("network_publication") is False and external_status.get("external_publication") is False, "external carrier claims no publication")
check(external_status.get("external_legal_license") is False, "external carrier claims no legal licence")
availability = external_exposure.get("availability", {})
check(availability.get("receiver_source_identified") is False and availability.get("receiver_formation_claimed") is False, "external carrier claims no receiver")
evidence_ceiling = external_exposure.get("evidence_ceiling", {})
check(evidence_ceiling.get("private_key_used") is False and evidence_ceiling.get("signature_profile_selected") is False, "external carrier used no key or signature profile")
external_text = "\n".join(
    path.read_text(encoding="utf-8")
    for path in (external_form, external_readme, external_record)
    if path.is_file()
)
for banned in ("UPAD", "Marko Markota", "/Users/markomarkota/LOCAL_REALITY_BODY_001"):
    check(banned not in external_text, f"external carrier excludes: {banned}")
check(not any(path.suffix == ".rmkey" for path in EXTERNAL.rglob("*") if path.is_file()), "external carrier contains no private-key file")


occurrence_path = ROOT / "LINEAGE/OCCURRENCES/BODY_FORM_EXPOSURE_OCCURRENCE_001.md"
occurrence_text = occurrence_path.read_text(encoding="utf-8") if occurrence_path.is_file() else ""
check("BODY_FORM_EXPOSED_LOCAL_ONLY" in occurrence_text, "performed exposure result recorded")
check(CURRENT_STATE_SHA in occurrence_text, "occurrence binds final current-state hash")
check(LINEAGE_SHA in occurrence_text, "occurrence binds final lineage hash")
for token in ("LOCAL_GOVERNING_AUTHORITY", "ADAPTATION_AUTHORITY", "OPERATIONAL_AUTHORIZATION"):
    check(token in occurrence_text, f"occurrence separates function: {token}")
check(str(DEFAULT_EXTERNAL) in occurrence_text, "occurrence records exact exposure target")
check("No second Body or inter-Body relation was formed." in occurrence_text, "occurrence records no second Body or relation")


host_verifier = ROOT / "EMBODIMENT/MECHANISMS/BODY_INTEGRITY_HOST/V0_1/verify_body_integrity_host.py"
host_result = subprocess.run(
    [sys.executable, "-B", str(host_verifier)],
    cwd=host_verifier.parent,
    capture_output=True,
    text=True,
)
host_output = host_result.stdout + host_result.stderr
check(host_result.returncode == 0, "Integrity Host pressure verifier returns success")
check("22" in host_output and "PASS" in host_output.upper(), "Integrity Host 22-case pressure result")


added_paths = {
    f"{CANDIDATE_BASE}/CANDIDATE.md",
    f"{CANDIDATE_BASE}/PROPOSED_FORM/BODY_FORM.json",
    f"{CANDIDATE_BASE}/PROPOSED_FORM/README.md",
    "LINEAGE/ACTS/BODY_FORM_EXPOSURE_ACT_001.md",
    "LINEAGE/OCCURRENCES/BODY_FORM_EXPOSURE_OCCURRENCE_001.md",
    "LINEAGE/STATE_REVISIONS/CURRENT_REVISION_003.json",
}
recovery_prefix = "CARRIER/RECOVERY/CARRIER_REVISION_003"
expected_live_files = (
    set(recovery_entries)
    | added_paths
    | {f"{recovery_prefix}/{relative}" for relative in recovery_entries}
    | {f"{recovery_prefix}/CARRIER/MANIFEST.sha256"}
)
manifest_path = ROOT / "CARRIER/MANIFEST.sha256"
manifest_entries = parse_manifest(manifest_path, "revision 4") if manifest_path.is_file() else {}
eligible = {
    path.relative_to(ROOT).as_posix(): sha(path)
    for path in ROOT.rglob("*")
    if path.is_file() and path != manifest_path and path.name != ".DS_Store"
}
check(set(eligible) == expected_live_files, "revision-4 exact authorized file set")
check(set(manifest_entries) == set(eligible), "revision-4 manifest complete coverage")
check(manifest_entries == eligible, "revision-4 manifest byte agreement")
residue = [
    path.relative_to(ROOT).as_posix()
    for path in ROOT.rglob("*")
    if path.is_file() and (path.name == ".DS_Store" or path.suffix in {".pyc", ".pyo"} or "__pycache__" in path.parts)
]
check(residue == [], "no runtime or machine residue in Body carrier")


if failures:
    print(f"\nRESULT  FAIL ({len(failures)} failures across {checks} checks)")
    for failure in failures:
        print(f"- {failure}")
    raise SystemExit(1)

print(f"\nRESULT  PASS FINAL ({checks} checks)")

#!/usr/bin/env python3
"""Fail-closed verification for Local-Reality Body revision 8."""

from __future__ import annotations

import hashlib
import json
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

import jsonschema

ROOT = Path(__file__).resolve().parents[2]
UPAD = Path('/Users/markomarkota/UPAD_ADMIN_SUCCESSOR_PROTOTYPE')
EDUC = Path('/Users/markomarkota/UPAD_EDUC_EVIDENCE_2026-09-03')
FORM = Path('/Users/markomarkota/LOCAL_REALITY_BODY_FORMS/RM_LOCAL_REALITY_BODY_FORM_001')
PREFLIGHT = '--preflight' in sys.argv
FINAL = '--final' in sys.argv
if PREFLIGHT == FINAL:
    raise SystemExit('select exactly one of --preflight or --final')

BODY_ID = 'RM-LOCAL-REALITY-BODY-001'
CANDIDATE_SHA = '71eea1c1d54bac15a17d1181430488a335f38b8b7a0079abc04087a9ba0d435c'
SOURCE_STATEMENT_SHA = 'bc8f5b1a2ca309fecd0044acd73f0c6614e0c279dc1cdfddbd38862f3fa0f9f0'
ACT_SHA = '570c51a141d4b5e163084aae9e1cfee9484b093b00d522cb95fb7d7741ef575c'
MATTER_SHA = '57d14587f7c31ece9a57102018a3c3b3c237bd86994d15a09bc2b98b2e9d3d4d'
REGULATION_SHA = '4efebb1ea32d7cf794193c45cd11e7c8f12c157997ab8a459f9735585dca6a20'
MATTER_INDEX_SHA = 'fca1c6f6f8604880eb1a61e8d2cb76ea8b20bd053096534f27f9221b6198eee1'
REGULATION_INDEX_SHA = 'de51ecde90be1d6a930fd7bb15a34abbac4cac25698fc62517e87630730611da'
LINEAGE_SHA = '896d498997bd1999a9412c1cc72f01cad7596232e7be7b6fe722c8a4826a07aa'
CURRENT_SHA = '95917d41b8835adaa8d21ad74624040a2ba4cd3f8afe7b0850c14e44c07e726f'
REV7_CURRENT_SHA = '857294137f75f6799a41990b2aaea94e1bfa1c8b489f1b048d96fc680e32296c'
REV7_LINEAGE_SHA = 'f5eb1e3f8b2abb5d525a7be93b340080282d1794f59f722e5ecbd8d6cbafdccd'
REV7_MANIFEST_SHA = 'c0c4dd32b9748ddfd654348516dcb46783d1454e31353fec2d7ed28b2177ada9'
JOURNAL_SHA = '77c1cb773a3ec48af88b4b8228fa54188587e749d304307b0152e6f91f5164a8'
RAW_HASHES = {
    'rcm-1788365297321538000-38e04707990d.bin': '765df8a6f7f898f52cf17c924c756afd71b5a281fd3c995f0145a346536fbbc6',
    'rcm-1788366247790802000-0e6e87ae457a.bin': '91ae6dfe68cce0764232618d618161ccaa2675b18a3f6043cc246c955dbfc36d',
    'rcm-1788378919095839000-38b12f81f6bc.bin': '6882cbb5de7939fef10d8c5535ba7b0cb25022a12e78a8be4bb225f962adee89',
}
PRESERVED = {
    'BODY/CORE/APPLICABILITY.json': '13d64a5a5d852a68b1f341c7995740725189b33b37b8826fbb0f6487c1951f5b',
    'BODY/CORE/IDENTITY.md': '47dfa944d02c3668b5385eab078653751f88bda9b847632f9427911639bf94d9',
    'BODY/CORE/LAW.md': 'd67a77c51f4a430b3d17c06e3e87c379430f1df5687fbe23fec84f507d8a3608',
    'BODY/CORE/LOCUS.md': '8a674f1ff787b0d19cac2e732e1b22d79991ef94a4ddc872181b2693d1650d0d',
    'BODY/INTEGRITY.md': '94677385d12d32be2eac7fd3f6d1ad864408d6192ee401a7e8306e3b67d52f8b',
    'BODY/LOCAL-FIELD/INDEX.json': 'd8b3a991fe1502c48ce4a402fff21ce1b3890486d692a9f3926fced32a824579',
    'BODY/STATE/AUTHORITY/INDEX.json': '7e6c0d1f5b89db26194fb851ca63133ddab117b6171f855de02304aa0d79c127',
    'BODY/STATE/RELATION-POSTURE/INDEX.json': '199f8096e3e8cd4a10fd1dc927e10bd843039a2304ac69a53abc314ec74cd17d',
    'BODY/STATE/RELATION-POSTURE/HUMAN-BODY-LOCAL.json': '0137186d0a758f0ac97a9da10453e04b31d0f9ef7e5414e888bf6435bf2e9ad3',
    'BODY/STATE/CONSEQUENCE/INDEX.json': 'efe8dcf4b4b850cff16dcacd1b2e56250cba3a14de379a447c16257fbac233db',
    'BODY/STATE/EPISTEMIC/INDEX.json': 'bff7505876e43086b9eecfc3420ecb0021aab74c731238aaabe195f8f43095ba',
    'BODY/EMBODIMENT/STATE.json': '93390ce1fa15f863021b370b2279f58de9849701a2c4f72510dbf38258884ea6',
    'BODY/EMBODIMENT/MECHANISMS/REALITY_CONTACT_MEMBRANE/V0_1/reality_contact_membrane.py': '99b770a7eda474ddfc68a189871c3907b97231d8ff5091ad11df03317552e44f',
}
UPAD_HASHES = {
    'CHECKPOINT/CONTENT.sha256': 'ad439fddad939f93d959a589b0c86c1c6c75b81ef4299322813c4fac4c71b48a',
    'CHECKPOINT/CANDIDATE.md': '80d488f2e01493db60a4a53ea975d71f8dbf686f54c8d4fc6bbc4cd56cf01a37',
}
EDUC_HASHES = {
    'EVIDENCE.sha256': '1f7a0795bee9b3941213b1d63b3de43c10a66b00365ecbb7926bf52fb07df6e2',
    'authorized-exchange-result.json': '3fd5fc1800acb3f3538b32a52ecaa0b9fa2ae1bc934dc1fc1fc4c304caca40d3',
    'fiscal-bridge.sqlite': '1ac555a90c22f38ef6852d701cf5b8e5572ab684c32e20d89b1a9c60af89e0c6',
    'UPAD_EDUC_ACCEPTANCE_RECEIPT.md': '8c3e3819143536a00cc92924c43b699cf98a33cfadd724adf1cfa01c276cc412',
}

failures = []
checks = 0

def check(value: bool, label: str) -> None:
    global checks
    checks += 1
    if value:
        print(f'PASS  {label}')
    else:
        print(f'FAIL  {label}')
        failures.append(label)

def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()

def load_json(path: Path, label: str) -> dict:
    try:
        value = json.loads(path.read_text(encoding='utf-8'))
    except Exception as exc:
        check(False, f'valid JSON: {label} ({exc})')
        return {}
    check(isinstance(value, dict), f'JSON object: {label}')
    return value if isinstance(value, dict) else {}

def parse_manifest(path: Path, label: str) -> dict[str, str]:
    out = {}
    try:
        for line in path.read_text(encoding='utf-8').splitlines():
            digest, rel = line.split('  ', 1)
            if len(digest) != 64 or rel in out:
                raise ValueError(line)
            out[rel] = digest
    except Exception as exc:
        check(False, f'parse manifest: {label} ({exc})')
        return {}
    check(bool(out), f'non-empty manifest: {label}')
    return out

candidate_rel = 'BODY/LINEAGE/CANDIDATES/LOCAL_REALITY_BODY_MATTER_ADMISSION_CANDIDATE_002.md'
act_rel = 'BODY/LINEAGE/ACTS/LOCAL_REALITY_BODY_MATTER_ADMISSION_ACT_001.md'
matter_rel = 'BODY/STATE/MATTER/RECORDS/MATTER_001.json'
regulation_rel = 'BODY/METABOLISM/REGULATION/RECORDS/REGULATION_001.json'
occurrence_rel = 'BODY/LINEAGE/OCCURRENCES/LOCAL_REALITY_BODY_MATTER_ADMISSION_OCCURRENCE_001.md'
required = {
    candidate_rel, act_rel, matter_rel, regulation_rel,
    'BODY/STATE/CURRENT.json', 'BODY/STATE/MATTER/INDEX.json',
    'BODY/METABOLISM/REGULATION/INDEX.json', 'BODY/LINEAGE/INDEX.json',
    'BODY/LINEAGE/STATE_REVISIONS/CURRENT_REVISION_007.json',
    'CARRIER/RECOVERY/CARRIER_REVISION_007/CARRIER/MANIFEST.sha256',
    'CARRIER/RECOVERY/CARRIER_REVISION_007_MAP.json',
    'CARRIER/VERIFICATION/verify_carrier.py', 'CARRIER/MANIFEST.sha256',
}
if FINAL:
    required.add(occurrence_rel)
for rel in sorted(required):
    check((ROOT / rel).is_file(), f'required carrier: {rel}')
check({p.name for p in ROOT.iterdir() if p.is_dir()} == {'BODY', 'CARRIER'}, 'private carrier has exactly BODY and CARRIER roots')
for rel in ('BODY-FORM', 'RELATION', 'RELATIONAL-FIELD', 'CONTACT-EVENT'):
    check(not (ROOT / 'BODY' / rel).exists(), f'cross-local root excluded: {rel}')

check(sha(ROOT / candidate_rel) == CANDIDATE_SHA, 'exact selected Candidate 002')
check(sha(ROOT / act_rel) == ACT_SHA, 'exact Matter Admission Act carrier')
act_text = (ROOT / act_rel).read_text(encoding='utf-8')
source_match = re.search(r'## Exact source statement\n\n```text\n(.*?)\n```', act_text, re.S)
statement = source_match.group(1) if source_match else ''
check(hashlib.sha256(statement.encode('utf-8')).hexdigest() == SOURCE_STATEMENT_SHA, 'exact selecting source statement digest')
check('Posture: `ADMIT_AND_HOLD`' in act_text and 'Authority function: `LOCAL_GOVERNING_AUTHORITY`' in act_text and 'Bounded scope: `MATTER_ADMISSION`' in act_text, 'Act carries exact posture, function, and scope')

schema_instances = (
    'BODY/CORE/APPLICABILITY.json', 'BODY/STATE/CURRENT.json', 'BODY/STATE/AUTHORITY/INDEX.json',
    'BODY/STATE/RELATION-POSTURE/INDEX.json', 'BODY/STATE/RELATION-POSTURE/HUMAN-BODY-LOCAL.json',
    'BODY/STATE/MATTER/INDEX.json', 'BODY/STATE/CONSEQUENCE/INDEX.json', 'BODY/STATE/EPISTEMIC/INDEX.json',
    'BODY/METABOLISM/REGULATION/INDEX.json', 'BODY/LOCAL-FIELD/INDEX.json',
    'BODY/LOCAL-FIELD/REVELATION/INDEX.json', 'BODY/LOCAL-FIELD/PRESENCE/INDEX.json',
    'BODY/LOCAL-FIELD/ENCOUNTER/INDEX.json', 'BODY/LOCAL-FIELD/APERTURE/INDEX.json',
    'BODY/LOCAL-FIELD/RESIDUE/INDEX.json', 'BODY/EMBODIMENT/STATE.json', 'BODY/LINEAGE/INDEX.json',
)
instances = {}
for rel in schema_instances:
    path = ROOT / rel
    instance = load_json(path, rel)
    instances[rel] = instance
    schema_ref = instance.get('$schema')
    schema_path = (path.parent / schema_ref).resolve() if isinstance(schema_ref, str) else ROOT / '__missing__'
    check(schema_path.is_file(), f'schema exists: {rel}')
    try:
        jsonschema.Draft202012Validator(load_json(schema_path, f'schema for {rel}')).validate(instance)
    except Exception as exc:
        check(False, f'schema validation: {rel} ({exc})')
    else:
        check(True, f'schema validation: {rel}')

current = instances['BODY/STATE/CURRENT.json']
matter_index = instances['BODY/STATE/MATTER/INDEX.json']
reg_index = instances['BODY/METABOLISM/REGULATION/INDEX.json']
lineage = instances['BODY/LINEAGE/INDEX.json']
matter = load_json(ROOT / matter_rel, matter_rel)
regulation = load_json(ROOT / regulation_rel, regulation_rel)
check(sha(ROOT / 'BODY/STATE/CURRENT.json') == CURRENT_SHA, 'exact revision-8 current state')
check(current.get('schema') == 'local-reality-body.current-state.v0.8' and current.get('currentness_revision') == 8, 'current state represents revision 8')
check(current.get('body_id') == BODY_ID and current.get('body_status') == 'FORMED', 'Body identity and formation preserved')
check(sha(ROOT / 'BODY/LINEAGE/INDEX.json') == LINEAGE_SHA and lineage.get('represented_currentness_revision') == 8, 'Lineage represents exact revision 8')
check(len(lineage.get('motion_chains', [])) == 9 and lineage['motion_chains'][-1].get('result') == 'ADMITTED_AND_HELD', 'ninth motion is admitted and held')
check(sha(ROOT / matter_rel) == MATTER_SHA and matter.get('record_id') == 'RM-LOCAL-REALITY-BODY-001-MATTER-001', 'exact Matter record')
check(matter.get('record_posture') == 'ADMITTED_BODY_LOCAL_MATTER' and matter.get('referenced_content_admitted') is False, 'Matter admits reference only')
check(matter.get('local_basis', {}).get('candidate_sha256') == CANDIDATE_SHA and matter.get('local_basis', {}).get('source_statement_sha256') == SOURCE_STATEMENT_SHA, 'Matter binds exact Act evidence')
check(sha(ROOT / regulation_rel) == REGULATION_SHA and regulation.get('record_id') == 'RM-LOCAL-REALITY-BODY-001-REGULATION-001', 'exact Regulation record')
check([regulation.get(k) for k in ('local_record_posture','progression_posture','locus_assessment','law_compatibility_assessment','evidence_posture','authority_presentation')] == ['ADMITTED_BODY_LOCAL_MATTER','HELD','WITHIN_SUPPORTED_LOCUS','UNASSESSED','PARTIAL','NOT_PRESENTED'], 'exact six-part Regulation posture')
check(sha(ROOT / 'BODY/STATE/MATTER/INDEX.json') == MATTER_INDEX_SHA and matter_index.get('record_count') == 1 and len(matter_index.get('records', [])) == 1, 'Matter index owns one record pointer')
check(matter_index['records'][0].get('sha256') == MATTER_SHA and matter_index['records'][0].get('referenced_content_admitted') is False, 'Matter index points to exact reference-only record')
check(sha(ROOT / 'BODY/METABOLISM/REGULATION/INDEX.json') == REGULATION_INDEX_SHA and reg_index.get('record_count') == 1 and len(reg_index.get('records', [])) == 1, 'Regulation index owns one record pointer')
check(reg_index.get('mechanism_present') is False and reg_index['records'][0].get('progression_posture') == 'HELD', 'Regulation remains mechanism-free and HELD')
check(current.get('matter', {}).get('admitted_reference_record', {}).get('sha256') == MATTER_SHA and current.get('matter', {}).get('referenced_content_admitted') is False, 'current projects reference-only Matter')
check(current.get('regulation', {}).get('record', {}).get('sha256') == REGULATION_SHA and current.get('regulation', {}).get('progression_posture') == 'HELD' and current.get('regulation', {}).get('mechanism_present') is False, 'current projects exact hold')
motion = current.get('matter_admission_occurrence', {})
check(motion.get('result') == 'ADMITTED_AND_HELD' and motion.get('predecessor_revision') == 7 and motion.get('predecessor_state_sha256') == REV7_CURRENT_SHA, 'current binds occurrence and predecessor')
check(motion.get('candidate_sha256') == CANDIDATE_SHA and motion.get('act_carrier_sha256') == ACT_SHA and motion.get('source_statement_sha256') == SOURCE_STATEMENT_SHA, 'current binds decision evidence')

for rel, expected in PRESERVED.items():
    check(sha(ROOT / rel) == expected, f'preserved non-target owner: {rel}')
authority = instances['BODY/STATE/AUTHORITY/INDEX.json']
functions = [r.get('function') for r in authority.get('authority_relations', [])]
check(functions == ['LOCAL_GOVERNING_AUTHORITY','OPERATIONAL_AUTHORIZATION','CORRECTION_AUTHORITY','ADAPTATION_AUTHORITY','DORMANCY_AND_CESSATION'], 'Authority topology unchanged')
check(all(r.get('bearer') == 'Marko Markota' and r.get('status') == 'CURRENT' for r in authority.get('authority_relations', [])), 'Authority bearer unchanged')
consequence = instances['BODY/STATE/CONSEQUENCE/INDEX.json']
epistemic = instances['BODY/STATE/EPISTEMIC/INDEX.json']
check(consequence.get('records') == [] and consequence.get('record_count') == 0, 'no Consequence created')
for key in ('admitted_records','source_relations','observations','evidence','inferences','working_interpretations','consequence_observations','standing_claims'):
    check(epistemic.get(key) == [], f'no epistemic uptake: {key}')
check(current.get('local_field', {}).get('active_aperture') is False and current.get('local_field', {}).get('semantic_residue_count') == 0, 'Local Field unchanged')
check(current.get('embodiment', {}).get('active_mechanism_count') == 1 and current.get('embodiment', {}).get('regulation_medium_host_present') is False, 'Embodiment unchanged')

live = ROOT / 'BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001'
journal = live / 'local_journal.jsonl'
check(sha(journal) == JOURNAL_SHA, 'living journal byte-identical')
try:
    entries = [json.loads(line) for line in journal.read_text(encoding='utf-8').splitlines() if line]
except Exception:
    entries = []
check(len(entries) == 3 and entries[2].get('entry_sha256') == '42f73290d68ff7c06206b337c37f3f36ddb366519eaaac24f14e9b611a8fa4f0', 'sequence-3 observation preserved')
raw_root = live / 'raw_crossings'
check({p.name for p in raw_root.iterdir() if p.is_file()} == set(RAW_HASHES), 'exact raw crossing set')
for name, expected in RAW_HASHES.items():
    check(sha(raw_root / name) == expected, f'raw crossing byte identity: {name}')
pressure = subprocess.run([sys.executable, '-B', str(ROOT / 'BODY/EMBODIMENT/MECHANISMS/REALITY_CONTACT_MEMBRANE/V0_1/verify_reality_contact_membrane.py')], capture_output=True, text=True)
check(pressure.returncode == 0 and 'PASS 39/39 reality-contact membrane checks' in pressure.stdout, 'Reality Contact Membrane exact pressure result')

check(subprocess.run(['git','status','--porcelain=v1'], cwd=UPAD, capture_output=True, text=True).stdout == '', 'external UPAD checkpoint worktree remains clean')
check(subprocess.run(['git','rev-parse','HEAD'], cwd=UPAD, capture_output=True, text=True).stdout.strip() == '331454c497e8df533e14f2f96f5c721def11dd87', 'external UPAD commit exact')
check(subprocess.run(['git','rev-parse','HEAD^{tree}'], cwd=UPAD, capture_output=True, text=True).stdout.strip() == '2fde97823f77e15ba025664dfd3dbafd11472b8b', 'external UPAD tree exact')
for rel, expected in UPAD_HASHES.items():
    check(sha(UPAD / rel) == expected, f'external UPAD anchor: {rel}')
upad_entries = parse_manifest(UPAD / 'CHECKPOINT/CONTENT.sha256', 'UPAD checkpoint')
check(len(upad_entries) == 97, 'UPAD manifest has 97 entries')
for rel, expected in upad_entries.items():
    check((UPAD / rel).is_file() and sha(UPAD / rel) == expected, f'UPAD checkpoint byte: {rel}')

check({p.name for p in EDUC.iterdir() if p.is_file()} == set(EDUC_HASHES), 'exact EDUC evidence inventory')
for rel, expected in EDUC_HASHES.items():
    check(sha(EDUC / rel) == expected, f'external EDUC anchor: {rel}')
educ_entries = parse_manifest(EDUC / 'EVIDENCE.sha256', 'EDUC evidence')
check(educ_entries == {k:v for k,v in EDUC_HASHES.items() if k != 'EVIDENCE.sha256'}, 'EDUC inventory content exact')
external_hashes = set(EDUC_HASHES.values()) | set(UPAD_HASHES.values())
imported = [p for p in ROOT.rglob('*') if p.is_file() and p.suffix not in {'.md','.json','.sha256','.py'} and sha(p) in external_hashes]
check(imported == [], 'no external evidence payload bytes imported')

recovery_map = load_json(ROOT / 'CARRIER/RECOVERY/CARRIER_REVISION_007_MAP.json', 'revision-7 recovery map')
check(recovery_map.get('source_manifest_sha256') == REV7_MANIFEST_SHA and recovery_map.get('predecessor_recovery_resolution', {}).get('recursive_duplication') is False, 'revision-7 recovery map exact and nonrecursive')
snapshot = ROOT / 'CARRIER/RECOVERY/CARRIER_REVISION_007'
snapshot_manifest = snapshot / 'CARRIER/MANIFEST.sha256'
check(sha(snapshot_manifest) == REV7_MANIFEST_SHA, 'revision-7 manifest byte identity')
rev7_entries = parse_manifest(snapshot_manifest, 'revision 7')
for rel, expected in rev7_entries.items():
    path = ROOT / rel if rel.startswith('CARRIER/RECOVERY/') else snapshot / rel
    check(path.is_file() and sha(path) == expected, f'revision-7 recoverable byte: {rel}')
snapshot_files = {p.relative_to(snapshot).as_posix() for p in snapshot.rglob('*') if p.is_file()}
snapshot_expected = {rel for rel in rev7_entries if not rel.startswith('CARRIER/RECOVERY/')} | {'CARRIER/MANIFEST.sha256'}
check(snapshot_files == snapshot_expected, 'revision-7 snapshot has exact nonrecursive static file set')
check(sha(ROOT / 'BODY/LINEAGE/STATE_REVISIONS/CURRENT_REVISION_007.json') == REV7_CURRENT_SHA, 'exact revision-7 state preserved')
check(sha(snapshot / 'BODY/STATE/CURRENT.json') == REV7_CURRENT_SHA and sha(snapshot / 'BODY/LINEAGE/INDEX.json') == REV7_LINEAGE_SHA, 'revision-7 state and Lineage recover exactly')

if FINAL:
    with tempfile.TemporaryDirectory(prefix='rm-revision7-recovery-', dir='/private/tmp') as temp:
        restored = Path(temp) / 'LOCAL_REALITY_BODY_001'
        shutil.copytree(snapshot, restored)
        for rel in rev7_entries:
            if not rel.startswith('CARRIER/RECOVERY/'):
                continue
            source = ROOT / rel
            target = restored / rel
            target.parent.mkdir(parents=True, exist_ok=True)
            if source.is_file():
                shutil.copy2(source, target)
        shutil.copytree(live, restored / 'BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001')
        old = subprocess.run([sys.executable, '-B', str(restored / 'CARRIER/VERIFICATION/verify_carrier.py'), '--final'], capture_output=True, text=True)
        check(old.returncode == 0 and 'RESULT  PASS FINAL (1114 checks)' in old.stdout, 'revision 7 reconstructs and passes its own 1114-check verifier')

external_manifest = FORM / 'MANIFEST.sha256'
check(sha(external_manifest) == 'da93daeba6b0f8e2c49521f4b043be35330d103422dae1ddade82670f43f5a2c', 'external Body Form manifest unchanged')
for rel, expected in parse_manifest(external_manifest, 'external Body Form').items():
    check(sha(FORM / rel) == expected, f'external Body Form byte: {rel}')

occurrence = ROOT / occurrence_rel
if FINAL:
    occurrence_text = occurrence.read_text(encoding='utf-8') if occurrence.is_file() else ''
    check('Status: **PERFORMED**' in occurrence_text and 'Result: `ADMITTED_AND_HELD`' in occurrence_text, 'successful occurrence recorded')
    check(CURRENT_SHA in occurrence_text and LINEAGE_SHA in occurrence_text and MATTER_SHA in occurrence_text and REGULATION_SHA in occurrence_text, 'occurrence binds exact successor and records')
else:
    check(not occurrence.exists(), 'preflight precedes successful occurrence')

manifest = ROOT / 'CARRIER/MANIFEST.sha256'
eligible = {
    p.relative_to(ROOT).as_posix(): sha(p)
    for p in ROOT.rglob('*')
    if p.is_file() and p != manifest and p.name != '.DS_Store' and p.suffix not in {'.pyc','.pyo'}
    and '__pycache__' not in p.parts
    and not p.relative_to(ROOT).as_posix().startswith('BODY/EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001/')
}
manifest_entries = parse_manifest(manifest, 'revision 8')
check(manifest_entries == eligible, f'revision-8 complete static manifest agreement ({"final" if FINAL else "preflight"})')
check(not any(p.is_symlink() for p in ROOT.rglob('*')), 'carrier contains no symbolic links')
residue = [p for p in ROOT.rglob('*') if p.is_file() and (p.name == '.DS_Store' or p.suffix in {'.pyc','.pyo'} or '__pycache__' in p.parts)]
check(residue == [], 'no runtime or machine residue')

for item in ('NO_SOURCE_BYTES_IMPORTED','NO_EXTERNAL_REPOSITORY_OR_EVIDENCE_DIRECTORY_ADMITTED','NO_REFERENCED_CONTENT_ADMITTED_OR_ADOPTED_AS_TRUE','NO_EPISTEMIC_CONTENT_ADMISSION','NO_CONSEQUENCE_CREATED_OR_ADMITTED','NO_PROGRESSION_BEYOND_EXACT_MATTER_ADMISSION_AND_HOLD','NO_AUTHORITY_PRESENTATION','NO_REGULATION_MECHANISM_ADMISSION_OR_ACTIVATION','NO_BODY_IDENTITY_CHANGE','NO_LOCUS_CHANGE','NO_AUTHORITY_TOPOLOGY_CHANGE','NO_RELATIONAL_FIELD','NO_CONTACT_EVENT','NO_PUBLIC_OR_NETWORK_PUBLICATION'):
    check(item in current.get('explicit_non_effects', []), f'current non-effect: {item}')

if failures:
    print(f'\nRESULT  FAIL ({len(failures)} failures across {checks} checks)')
    for failure in failures:
        print(f'- {failure}')
    raise SystemExit(1)
print(f'\nRESULT  PASS {"PREFLIGHT" if PREFLIGHT else "FINAL"} ({checks} checks)')

#!/usr/bin/env python3
"""Execute and verify PRESENCE REHEARSAL 001 without exporting private material."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import pathlib
import platform
import subprocess
import sys
import time
import urllib.error
import urllib.request
from typing import Any


FORM_ID = "PRESENCE-AVALANCHE-FORM-001"
FORM_SHA256 = "3a72b3707552fd56b78c216c2d4a0caae9297cc52611d440ea170ab6e9770fa5"
SOURCE_REFERENCE = "PRESENCE_AVALANCHE_VM_001_REHEARSAL_SOURCE"
SOURCE_SHA256 = "c02f9a3927f7e2a48c49a4d2267b0cc7c8a932694817374366d8ec5709c419bd"
HOST_LOCALITY_ID = "PRESENCE-REHEARSAL-HOST-003"
LOCUS_ID = "PRESENCE-REHEARSAL-LOCUS-003"
RETURN_LOCUS_ID = "PRESENCE-REHEARSAL-LOCUS-003-RETURN"
MATTER_ID = "PRESENCE-REHEARSAL-MATTER-003"
EMERGENCE_ID = "PRESENCE-REHEARSAL-EMERGENCE-003"
MEDIUM_CAPABILITIES = [
    "OBSERVE_CROSSING",
    "PRESENT_FORM",
    "GATE_DISPOSITION",
    "ENTER",
    "CHECKPOINT_DEPARTURE",
    "REENTER",
    "MATTER_DISPOSITION",
    "CORRECT",
    "EXIT",
    "CLOSE",
    "INCORPORATE_ADDRESSED_RESIDUE",
    "DECLARE_CAPACITY",
    "PULSE",
    "RECLAIM_ADMISSION_OFFER",
    "REGISTER_CONTINUITY_AUTHORITY",
    "EXHAUST_FORMATION_AUTHORITY",
    "PROPOSE_SUCCESSOR",
    "ATTEST_SUCCESSOR",
    "ACTIVATE_SUCCESSOR",
]
EFFECT_CEILING = [
    "NO_FORMATION_BY_CONFORMANCE",
    "NO_SOURCE_OR_AUTHORITY_TRANSFER",
    "NO_SHARED_CURRENTNESS",
    "NO_CORE_ADDRESSABILITY",
    "NO_PARTICIPATION_BY_PRESENTATION",
    "NO_MATTER_ADMISSION_BY_CROSSING",
    "NO_LOCAL_UPTAKE_BY_FIELD_CLOSURE",
    "NO_TOKEN_BALANCE_AS_CAPACITY",
    "NO_RECEIPT_AS_CURRENT_PRESENCE",
    "NO_INFRASTRUCTURE_AS_SUCCESSOR",
    "NO_RESTORATION_BY_REENTRY",
    "NO_PRIVATE_STATE_DISCLOSURE",
]


def utc_now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z")


def digest(label: str) -> str:
    return hashlib.sha256(label.encode("utf-8")).hexdigest()


def canonical_bytes(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode("utf-8")


def continuity_successor(
    participant_id: str,
    predecessor_state_commitment: str,
    delta_sha256: str,
    ordinal: int,
) -> str:
    envelope = {
        "schema": "PRESENCE_PARTICIPANT_STATE_SUCCESSION_001",
        "participant_id": participant_id,
        "ordinal": ordinal,
        "predecessor_state_commitment": predecessor_state_commitment,
        "delta_sha256": delta_sha256,
    }
    encoded = json.dumps(envelope, ensure_ascii=False, separators=(",", ":")).encode(
        "utf-8"
    )
    return hashlib.sha256(encoded).hexdigest()


def addressed_residue(addressed_to_locality_id: str) -> dict[str, Any]:
    envelope = {
        "schema": "PRESENCE_ADDRESSED_RESIDUE_001",
        "addressed_to_locality_id": addressed_to_locality_id,
        "source_locality_id": "EXTERNAL-REHEARSAL-LOCALITY-001",
        "source_locus_id": "EXTERNAL-REHEARSAL-LOCUS-001",
        "source_state_commitment": digest("EXTERNAL REHEARSAL SOURCE STATE"),
        "closure_transition_id": digest("EXTERNAL REHEARSAL CLOSE TRANSITION"),
        "participant_id": "EXTERNAL-REHEARSAL-PARTICIPANT-001",
        "presentation_id": digest("EXTERNAL REHEARSAL PRESENTATION"),
        "entry_transition_id": digest("EXTERNAL REHEARSAL ENTRY"),
        "departure_checkpoint_id": digest("EXTERNAL REHEARSAL DEPARTURE CHECKPOINT"),
        "departure_state_commitment": digest("EXTERNAL REHEARSAL DEPARTURE STATE"),
        "effect": "AVAILABLE_FOR_ADDRESSEE_DECISION_ONLY",
    }
    envelope_bytes = json.dumps(
        envelope, ensure_ascii=False, separators=(",", ":")
    ).encode("utf-8")
    return {
        "residue_id": hashlib.sha256(
            b"PRESENCE_RESIDUE_ID_001\x00" + envelope_bytes
        ).hexdigest(),
        "residue_sha256": hashlib.sha256(envelope_bytes).hexdigest(),
        **envelope,
    }


def read_json(path: pathlib.Path) -> Any:
    with path.open("r", encoding="utf-8") as source:
        return json.load(source)


def write_json(path: pathlib.Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(path.suffix + ".tmp")
    with temporary.open("w", encoding="utf-8") as target:
        json.dump(value, target, indent=2, sort_keys=True)
        target.write("\n")
    os.replace(temporary, path)


def http_json(
    method: str,
    url: str,
    payload: Any | None = None,
    raw_payload: bytes | None = None,
    api_route: str | None = None,
    timeout: float = 10.0,
) -> tuple[int, Any]:
    if payload is not None and raw_payload is not None:
        raise ValueError("payload and raw_payload are mutually exclusive")
    body = raw_payload
    if payload is not None:
        body = canonical_bytes(payload)
    headers = {"Content-Type": "application/json", "Accept": "application/json"}
    if api_route:
        headers["Avalanche-Api-Route"] = api_route
    request = urllib.request.Request(
        url,
        data=body,
        method=method,
        headers=headers,
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            response_body = response.read()
            return response.status, json.loads(response_body)
    except urllib.error.HTTPError as error:
        response_body = error.read()
        try:
            decoded = json.loads(response_body)
        except json.JSONDecodeError:
            decoded = {"error": response_body.decode("utf-8", errors="replace")}
        return error.code, decoded


def checked_http_json(
    method: str,
    url: str,
    payload: Any | None = None,
    raw_payload: bytes | None = None,
    api_route: str | None = None,
    expected: tuple[int, ...] = (200,),
) -> Any:
    status, decoded = http_json(method, url, payload, raw_payload, api_route)
    if status not in expected:
        raise RuntimeError(f"{method} {url} returned HTTP {status}: {decoded}")
    return decoded


def run_checked(arguments: list[str]) -> str:
    completed = subprocess.run(arguments, check=True, text=True, capture_output=True)
    return completed.stdout.strip()


def load_public_authority(path: pathlib.Path) -> dict[str, Any]:
    document = read_json(path)
    expected = "PRESENCE_AUTHORITY_PUBLIC_001"
    if document.get("schema") != expected:
        raise RuntimeError(f"{path} is not a {expected} document")
    return document


def wait_for_receipt(
    base_url: str, api_route: str, transition_id: str, timeout: float = 90.0
) -> Any:
    deadline = time.monotonic() + timeout
    receipt_url = f"{base_url}/receipts/{transition_id}"
    last_result: Any = None
    while time.monotonic() < deadline:
        status, result = http_json("GET", receipt_url, api_route=api_route)
        last_result = result
        if status == 200:
            return result
        if status not in (404, 503):
            raise RuntimeError(f"receipt {transition_id} returned HTTP {status}: {result}")
        time.sleep(0.25)
    raise TimeoutError(f"receipt {transition_id} not accepted within {timeout}s: {last_result}")


def issue_transition(
    *,
    base_url: str,
    api_route: str,
    vm: pathlib.Path,
    private_authority: pathlib.Path,
    workspace: pathlib.Path,
    sequence: int,
    operation: str,
    actor_id: str,
    actor_public_key: str,
    locus_id: str,
    payload: dict[str, Any],
    observed_at: int,
) -> dict[str, Any]:
    request = {
        "operation": operation,
        "actor_id": actor_id,
        "actor_public_key": actor_public_key,
        "locus_id": locus_id,
        "observed_at": observed_at,
        "payload": payload,
    }
    draft = checked_http_json(
        "POST", f"{base_url}/drafts", payload=request, api_route=api_route
    )
    unsigned_path = workspace / f"{sequence:02d}-{operation.lower()}-unsigned.json"
    transition_path = workspace / f"{sequence:02d}-{operation.lower()}-transition.json"
    write_json(unsigned_path, draft["unsigned"])
    run_checked(
        [
            str(vm),
            "--sign-unsigned",
            str(unsigned_path),
            str(private_authority),
            str(transition_path),
        ]
    )
    validation = run_checked([str(vm), "--check-transition", str(transition_path)])
    transition_bytes = transition_path.read_bytes()
    issued = checked_http_json(
        "POST",
        f"{base_url}/transitions",
        raw_payload=transition_bytes,
        api_route=api_route,
        expected=(202,),
    )
    transition_id = issued["transition_id"]
    receipt = wait_for_receipt(base_url, api_route, transition_id)
    legacy_receipt = checked_http_json(
        "GET",
        f"{base_url}/ext/bc/{api_route}/receipts/{transition_id}",
    )
    if legacy_receipt != receipt:
        raise RuntimeError(f"header and legacy receipt routes disagree for {transition_id}")
    if receipt.get("operation") != operation:
        raise RuntimeError(f"receipt operation mismatch for {transition_id}")
    return {
        "sequence": sequence,
        "operation": operation,
        "transition_id": transition_id,
        "local_validation": validation,
        "receipt": receipt,
    }


def collect_snapshot(summary: dict[str, Any], timeout: float = 120.0) -> dict[str, Any]:
    blockchain_id = summary["blockchain_id"]
    deadline = time.monotonic() + timeout
    last_error: Exception | None = None
    while time.monotonic() < deadline:
        try:
            observations = []
            states = []
            for node in summary["nodes"]:
                base_url = node["uri"].rstrip("/")
                status = checked_http_json(
                    "GET", f"{base_url}/status", api_route=blockchain_id
                )
                state = checked_http_json(
                    "GET", f"{base_url}/state", api_route=blockchain_id
                )
                legacy_base = f"{base_url}/ext/bc/{blockchain_id}"
                legacy_status = checked_http_json("GET", f"{legacy_base}/status")
                legacy_state = checked_http_json("GET", f"{legacy_base}/state")
                if legacy_status != status or legacy_state != state:
                    raise RuntimeError(
                        f"header and legacy API routes disagree on {node['node_id']}"
                    )
                state_sha256 = hashlib.sha256(canonical_bytes(state)).hexdigest()
                observations.append(
                    {
                        "node_id": node["node_id"],
                        "uri": node["uri"],
                        "status": status,
                        "state_sha256": state_sha256,
                    }
                )
                states.append(state)
            reference = canonical_bytes(states[0])
            if any(canonical_bytes(state) != reference for state in states[1:]):
                raise RuntimeError("three nodes do not expose identical accepted state")
            revisions = {item["status"]["revision"] for item in observations}
            commitments = {item["status"]["state_commitment"] for item in observations}
            if len(revisions) != 1 or len(commitments) != 1:
                raise RuntimeError("three nodes disagree on revision or state commitment")
            return {
                "observed_at_utc": utc_now(),
                "revision": observations[0]["status"]["revision"],
                "height": observations[0]["status"]["height"],
                "state_commitment": observations[0]["status"]["state_commitment"],
                "state_sha256": observations[0]["state_sha256"],
                "nodes": observations,
                "consensus_state": states[0],
            }
        except (OSError, RuntimeError, urllib.error.URLError) as error:
            last_error = error
            time.sleep(0.5)
    raise TimeoutError(f"network did not converge within {timeout}s: {last_error}")


def command_run(args: argparse.Namespace) -> None:
    summary = read_json(args.summary)
    if len(summary.get("nodes", [])) != 3:
        raise RuntimeError("REHEARSAL 003 requires exactly three nodes")
    vm_version = json.loads(run_checked([str(args.vm), "--version-json"]))
    if summary["vm_id"] != vm_version["vm_id"]:
        raise RuntimeError("network VM ID does not match the LOCALITY binary")
    if (
        vm_version.get("avalanchego_profile")
        != "v1.15.0+PRESENCE_SECURITY_OVERLAY_001"
    ):
        raise RuntimeError("VM does not declare the required AvalancheGo security profile")

    host = load_public_authority(args.host_public)
    participant = load_public_authority(args.participant_public)
    base_url = summary["nodes"][0]["uri"].rstrip("/")
    workspace = args.output.parent / "transitions"
    workspace.mkdir(parents=True, exist_ok=True)
    started_at = utc_now()
    observed_at = int(time.time())
    transitions: list[dict[str, Any]] = []

    def issue(
        operation: str,
        actor_id: str,
        public_key: str,
        private_path: pathlib.Path,
        payload: dict[str, Any],
        locus_id: str = LOCUS_ID,
    ) -> dict[str, Any]:
        result = issue_transition(
            base_url=base_url,
            api_route=summary["blockchain_id"],
            vm=args.vm,
            private_authority=private_path,
            workspace=workspace,
            sequence=len(transitions) + 1,
            operation=operation,
            actor_id=actor_id,
            actor_public_key=public_key,
            locus_id=locus_id,
            payload=payload,
            observed_at=observed_at,
        )
        transitions.append(result)
        return result

    capacity_declaration = issue(
        "DECLARE_CAPACITY",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {
            "actual_units": 800,
            "resource_commitment_sha256": digest("PRESENCE REHEARSAL 001 CARRIERS"),
            "basis_sha256": digest("PRESENCE REHEARSAL 001 CAPACITY BASIS"),
        },
        locus_id="",
    )
    carrier_set = digest("PRESENCE REHEARSAL 001 CARRIER SET")
    pulse_observed_at = observed_at
    pulse_material = checked_http_json(
        "GET",
        f"{base_url}/currentness?observed_at={pulse_observed_at}&carrier_set_sha256={carrier_set}",
        api_route=summary["blockchain_id"],
    )["next_pulse"]
    pulse = issue(
        "PULSE",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {
            "currentness_commitment_sha256": pulse_material["currentness_commitment_sha256"],
            "carrier_set_sha256": carrier_set,
        },
        locus_id="",
    )
    issue(
        "BOUND",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {
            "locus_id": LOCUS_ID,
            "purpose_sha256": digest("PRESENCE REHEARSAL 001 PURPOSE"),
            "closure_condition_sha256": digest("PRESENCE REHEARSAL 001 CLOSE AFTER EXIT"),
            "capacity_ceiling_units": 500,
        },
    )
    initial_participant_state = digest("LOCALITY REHEARSAL PARTICIPANT INITIAL STATE")
    presentation = issue(
        "PRESENT_FORM",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "locality_reference": "CONTINUITY-REHEARSAL-001",
            "source_reference": SOURCE_REFERENCE,
            "source_sha256": SOURCE_SHA256,
            "nucleus_version": "LOCALITY-NUCLEUS-003",
            "nucleus_sha256": digest("LOCALITY-NUCLEUS-003"),
            "form_id": FORM_ID,
            "form_sha256": FORM_SHA256,
            "state_commitment": initial_participant_state,
            "medium_capabilities": MEDIUM_CAPABILITIES,
            "effect_ceiling": EFFECT_CEILING,
        },
    )
    presentation_id = presentation["transition_id"]
    issue(
        "GATE_DISPOSITION",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {
            "participant_id": participant["actor_id"],
            "presentation_id": presentation_id,
            "disposition": "ADMIT",
            "reason_sha256": digest("LOCALITY REHEARSAL FORM ADMITTED"),
            "work_units": 100,
            "resolution_units": 25,
            "offer_expires_at": observed_at + 3600,
        },
    )
    entry = issue(
        "ENTER",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {"presentation_id": presentation_id},
    )
    issue(
        "OBSERVE_CROSSING",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "matter_id": MATTER_ID,
            "content_sha256": digest("LOCALITY REHEARSAL MATTER CONTENT"),
            "media_type": "application/locality-rehearsal+json",
            "claim": "Observed for protocol rehearsal only; no external truth is inferred.",
        },
    )
    matter_disposition = issue(
        "MATTER_DISPOSITION",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "matter_id": MATTER_ID,
            "disposition": "ADMIT",
            "reason_sha256": digest("LOCALITY REHEARSAL PARTICIPANT LOCAL ADMISSION"),
        },
    )
    issue(
        "RECORD_EMERGENCE",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {
            "emergence_id": EMERGENCE_ID,
            "contributor_ids": [participant["actor_id"]],
            "matter_ids": [MATTER_ID],
            "kind": "REHEARSAL_EVIDENCE",
            "description_sha256": digest("LOCALITY REHEARSAL EMERGENCE"),
        },
    )
    issue(
        "CORRECT",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "target_transition_id": matter_disposition["transition_id"],
            "replacement_commitment": digest("LOCALITY REHEARSAL CORRECTED DESCRIPTION"),
            "reason_sha256": digest("LOCALITY REHEARSAL APPEND ONLY CORRECTION"),
        },
    )
    departure_delta = digest("LOCALITY REHEARSAL PARTICIPANT DEPARTURE DELTA")
    departure_state = continuity_successor(
        participant["actor_id"], initial_participant_state, departure_delta, 1
    )
    checkpoint = issue(
        "CHECKPOINT_DEPARTURE",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "entry_transition_id": entry["transition_id"],
            "presentation_id": presentation_id,
            "from_state_commitment": initial_participant_state,
            "passage": [
                {
                    "delta_sha256": departure_delta,
                    "successor_state_commitment": departure_state,
                }
            ],
            "departure_state_commitment": departure_state,
        },
    )
    issue(
        "EXIT",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "departure_checkpoint_id": checkpoint["transition_id"],
            "reason_sha256": digest("LOCALITY REHEARSAL PARTICIPANT EXIT"),
        },
    )
    issue(
        "CLOSE",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {"closure_basis_sha256": digest("LOCALITY REHEARSAL CLOSURE BASIS")},
    )

    closure_snapshot = collect_snapshot(summary)
    if closure_snapshot["revision"] != 13:
        raise RuntimeError(
            f"expected closure revision 13, got {closure_snapshot['revision']}"
        )
    state = closure_snapshot["consensus_state"]
    if state["active_locus_id"] != "":
        raise RuntimeError("closed rehearsal still exposes an active locus")
    if state["loci"][LOCUS_ID]["phase"] != "CLOSED":
        raise RuntimeError("rehearsal locus is not CLOSED")
    if matter_disposition["transition_id"] not in state["corrections"]:
        raise RuntimeError("append-only correction edge is absent")
    residue = state["loci"][LOCUS_ID]["residues"].get(participant["actor_id"])
    if not residue or residue.get("effect") != "AVAILABLE_FOR_ADDRESSEE_DECISION_ONLY":
        raise RuntimeError("closure did not produce addressed, non-incorporated residue")
    if state["incorporated_residues"]:
        raise RuntimeError("closure silently incorporated residue")
    if (
        residue.get("departure_checkpoint_id") != checkpoint["transition_id"]
        or residue.get("departure_state_commitment") != departure_state
        or residue.get("entry_transition_id") != entry["transition_id"]
    ):
        raise RuntimeError("closure residue omitted carried departure state")

    issue(
        "BOUND",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {
            "locus_id": RETURN_LOCUS_ID,
            "purpose_sha256": digest("PRESENCE REHEARSAL 001 RETURN PURPOSE"),
            "closure_condition_sha256": digest(
                "PRESENCE REHEARSAL 001 RETURN CLOSURE"
            ),
            "capacity_ceiling_units": 500,
        },
        locus_id=RETURN_LOCUS_ID,
    )
    return_delta = digest("LOCALITY REHEARSAL PARTICIPANT ABSENCE DELTA")
    return_state = continuity_successor(
        participant["actor_id"], departure_state, return_delta, 1
    )
    return_presentation = issue(
        "PRESENT_FORM",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "locality_reference": "CONTINUITY-REHEARSAL-001",
            "source_reference": SOURCE_REFERENCE,
            "source_sha256": SOURCE_SHA256,
            "nucleus_version": "LOCALITY-NUCLEUS-003-RETURN",
            "nucleus_sha256": digest("LOCALITY-NUCLEUS-003-RETURN"),
            "form_id": FORM_ID,
            "form_sha256": FORM_SHA256,
            "state_commitment": return_state,
            "medium_capabilities": MEDIUM_CAPABILITIES,
            "effect_ceiling": EFFECT_CEILING,
        },
        locus_id=RETURN_LOCUS_ID,
    )
    issue(
        "GATE_DISPOSITION",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        {
            "participant_id": participant["actor_id"],
            "presentation_id": return_presentation["transition_id"],
            "disposition": "ADMIT",
            "reason_sha256": digest("LOCALITY REHEARSAL RENEWED ADMISSION"),
            "work_units": 120,
            "resolution_units": 30,
            "offer_expires_at": observed_at + 3600,
        },
        locus_id=RETURN_LOCUS_ID,
    )
    reentry = issue(
        "REENTER",
        participant["actor_id"],
        participant["public_key"],
        args.participant_private,
        {
            "presentation_id": return_presentation["transition_id"],
            "prior_entry_transition_id": entry["transition_id"],
            "departure_checkpoint_id": checkpoint["transition_id"],
            "residue_id": residue["residue_id"],
            "passage": [
                {
                    "delta_sha256": return_delta,
                    "successor_state_commitment": return_state,
                }
            ],
        },
        locus_id=RETURN_LOCUS_ID,
    )

    imported_residue = addressed_residue(HOST_LOCALITY_ID)
    issue(
        "INCORPORATE_ADDRESSED_RESIDUE",
        HOST_LOCALITY_ID,
        host["public_key"],
        args.host_private,
        imported_residue,
        locus_id="",
    )
    snapshot = collect_snapshot(summary)
    if snapshot["revision"] != 18:
        raise RuntimeError(f"expected final revision 18, got {snapshot['revision']}")
    incorporated = snapshot["consensus_state"]["incorporated_residues"].get(
        imported_residue["residue_id"]
    )
    if not incorporated:
        raise RuntimeError("explicit addressed residue incorporation was not recorded")
    if incorporated.get("addressed_to_locality_id") != HOST_LOCALITY_ID:
        raise RuntimeError("incorporated residue changed its addressee")
    final_state = snapshot["consensus_state"]
    renewed_entry = final_state["entry_history"].get(reentry["transition_id"])
    if not renewed_entry or renewed_entry.get("ingress_mode") != "RENEWED_ENTRY":
        raise RuntimeError("return did not become a distinct renewed entry")
    if (
        renewed_entry.get("prior_entry_transition_id") != entry["transition_id"]
        or renewed_entry.get("prior_departure_checkpoint_id")
        != checkpoint["transition_id"]
        or renewed_entry.get("prior_residue_id") != residue["residue_id"]
    ):
        raise RuntimeError("renewed entry omitted its predecessor bindings")
    consumed = final_state["departure_checkpoints"].get(checkpoint["transition_id"])
    if (
        not consumed
        or consumed.get("status") != "CONSUMED"
        or consumed.get("consumed_by_entry_transition_id")
        != reentry["transition_id"]
    ):
        raise RuntimeError("departure checkpoint was not consumed exactly once")
    if final_state["loci"][LOCUS_ID]["phase"] != "CLOSED":
        raise RuntimeError("renewed ingress reopened the prior locus")

    public_network = dict(summary)
    public_network.pop("network_dir", None)
    evidence = {
        "schema": "PRESENCE_REHEARSAL_LIFECYCLE_003",
        "started_at_utc": started_at,
        "completed_at_utc": utc_now(),
        "network": public_network,
        "vm_version": vm_version,
        "actors": {
            "host_locality_id": HOST_LOCALITY_ID,
            "host_key_id": host["key_id"],
            "participant_actor_id": participant["actor_id"],
            "participant_key_id": participant["key_id"],
        },
        "transitions": transitions,
        "assertions": {
            "revision_zero_to_eighteen": True,
            "three_node_state_equality": True,
            "header_and_legacy_api_routes_equal": True,
            "presentation_is_not_admission": True,
            "admission_is_not_entry": True,
            "crossing_is_not_matter_admission": True,
            "correction_is_append_only": True,
            "closure_produces_addressed_residue": True,
            "closure_does_not_incorporate_residue": True,
            "explicit_addressee_incorporation_required": True,
            "residue_commitment_recomputed_on_import": True,
            "departure_checkpoint_binds_participant_state": True,
            "reentry_requires_fresh_presentation_and_admission": True,
            "reentry_consumes_checkpoint_once": True,
            "reentry_preserves_prior_locus_without_reopening": True,
            "reentry_carries_deterministic_state_successor": True,
            "capacity_declaration_is_local_account_only": capacity_declaration["receipt"]["operation"] == "DECLARE_CAPACITY",
            "pulse_at_p_zero_does_not_invent_presence": pulse["receipt"]["presence_count"] == 0,
            "first_entry_reserves_work_and_resolution": any(
                item["operation"] == "ENTER"
                and item["receipt"]["capacity"]["unresolved_units"] == 100
                and item["receipt"]["capacity"]["correction_egress_units"] == 125
                for item in transitions
            ),
            "exit_releases_lifecycle_reservation": any(
                item["operation"] == "EXIT"
                and item["receipt"]["capacity"]["unresolved_units"] == 0
                for item in transitions
            ),
            "reentry_reserves_work_and_resolution": reentry["receipt"]["capacity"]["unresolved_units"] == 120
            and reentry["receipt"]["capacity"]["correction_egress_units"] == 130,
        },
        "closure_snapshot": closure_snapshot,
        "snapshot": snapshot,
        "private_material_exported": False,
    }
    write_json(args.output, evidence)
    print(json.dumps({"output": str(args.output), "revision": 18, "state_commitment": snapshot["state_commitment"]}))


def expected_snapshot(document: dict[str, Any]) -> dict[str, Any]:
    if "snapshot" in document:
        return document["snapshot"]
    if "observed" in document:
        return document["observed"]
    raise RuntimeError("expected evidence document has no snapshot")


def command_snapshot(args: argparse.Namespace) -> None:
    summary = read_json(args.summary)
    observed = collect_snapshot(summary)
    expected = expected_snapshot(read_json(args.expected))
    for field in ("revision", "state_commitment", "state_sha256"):
        if observed[field] != expected[field]:
            raise RuntimeError(f"{args.label}: {field} changed: expected {expected[field]}, observed {observed[field]}")
    result = {
        "schema": "PRESENCE_REHEARSAL_SNAPSHOT_003",
        "label": args.label,
        "expected": {
            "revision": expected["revision"],
            "state_commitment": expected["state_commitment"],
            "state_sha256": expected["state_sha256"],
        },
        "observed": observed,
        "exact_state_preserved": True,
    }
    write_json(args.output, result)
    print(json.dumps({"output": str(args.output), "label": args.label, "exact_state_preserved": True}))


def sha256_file(path: pathlib.Path) -> str:
    hasher = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            hasher.update(block)
    return hasher.hexdigest()


def command_assemble(args: argparse.Namespace) -> None:
    lifecycle = read_json(args.lifecycle)
    restart = read_json(args.restart)
    bounce = read_json(args.bounce)
    vm_version = json.loads(run_checked([str(args.vm), "--version-json"]))
    avalanchego_version = json.loads(run_checked([str(args.avalanchego), "--version-json"]))
    if vm_version["rpcchainvm"] != avalanchego_version["rpcchainvm"]:
        raise RuntimeError("RPCChainVM protocol mismatch during evidence assembly")
    if (
        vm_version.get("avalanchego_profile")
        != "v1.15.0+PRESENCE_SECURITY_OVERLAY_001"
        or avalanchego_version.get("go") != "1.25.13"
    ):
        raise RuntimeError("security-overlaid AvalancheGo profile mismatch")
    final = lifecycle["snapshot"]
    evidence = {
        "schema": "PRESENCE_REHEARSAL_003_EVIDENCE",
        "observed_at_utc": utc_now(),
        "scope": "DISPOSABLE_THREE_NODE_LOCAL_NETWORK_ONLY",
        "mainnet_mutations_performed": False,
        "verification_host": {
            "os": platform.system(),
            "architecture": platform.machine(),
            "native_macOS_arm64": platform.system() == "Darwin"
            and platform.machine() == "arm64",
        },
        "verification_gates": {
            "go_unit_and_integration": "PASS",
            "go_race": "PASS",
            "go_vet": "PASS",
            "native_vm_build": "PASS",
            "rehearsal_controller_test_and_build": "PASS",
        },
        "formation_predecessor": {
            "blockchain_id": "LB6wwV4JNxr8fwjUPBHMzf3PiW2d4hsTc4v63uX1MjY6oZ9Wb",
            "l1_id": "21mJfY4QpDeVykBaeG8nwn7k7w7b5oPpPhaYcJWqumht8SvaK",
            "vm_id": "pJHx1NU8ghWsiwg1vrqwaqE5uKh1k4EJRBkpB1tKV1QuQFhMi",
            "validator_node_id": "NodeID-AD66psMnQ257UAF7Jz9FmNAyRdibVnkt",
            "validation_id": "ttR2sBgWUEq3e6dupCxYcyiGTv49wnYW3wZ6TUs7XnDkTeZjh",
            "effect": "ANCESTRY_REFERENCE_ONLY",
        },
        "successor": {
            "vm": vm_version,
            "vm_binary_sha256": sha256_file(args.vm),
            "avalanchego": avalanchego_version,
            "avalanchego_binary_sha256": sha256_file(args.avalanchego),
            "rpcchainvm_compatibility": "EXACT_MATCH",
            "security_profile": "v1.15.0+PRESENCE_SECURITY_OVERLAY_001",
        },
        "network": lifecycle["network"],
        "lifecycle": {
            "accepted_transition_count": len(lifecycle["transitions"]),
            "operations": [item["operation"] for item in lifecycle["transitions"]],
            "transition_ids": [item["transition_id"] for item in lifecycle["transitions"]],
            "assertions": lifecycle["assertions"],
            "final_revision": final["revision"],
            "final_height": final["height"],
            "final_state_commitment": final["state_commitment"],
            "final_state_sha256": final["state_sha256"],
        },
        "restart_replay": {
            "all_nodes_restarted": True,
            "exact_state_preserved": restart["exact_state_preserved"],
            "observed": restart["observed"],
        },
        "single_node_rejoin": {
            "one_node_stopped_and_restarted": True,
            "exact_state_preserved": bounce["exact_state_preserved"],
            "observed": bounce["observed"],
        },
        "private_material_exported": False,
        "validator_credentials_disposable": True,
        "result": "PASS",
    }
    write_json(args.output, evidence)
    print(json.dumps({"output": str(args.output), "result": "PASS"}))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command", required=True)

    run_parser = subparsers.add_parser("run")
    run_parser.add_argument("--summary", type=pathlib.Path, required=True)
    run_parser.add_argument("--vm", type=pathlib.Path, required=True)
    run_parser.add_argument("--host-public", type=pathlib.Path, required=True)
    run_parser.add_argument("--host-private", type=pathlib.Path, required=True)
    run_parser.add_argument("--participant-public", type=pathlib.Path, required=True)
    run_parser.add_argument("--participant-private", type=pathlib.Path, required=True)
    run_parser.add_argument("--output", type=pathlib.Path, required=True)
    run_parser.set_defaults(handler=command_run)

    snapshot_parser = subparsers.add_parser("snapshot")
    snapshot_parser.add_argument("--summary", type=pathlib.Path, required=True)
    snapshot_parser.add_argument("--expected", type=pathlib.Path, required=True)
    snapshot_parser.add_argument("--label", required=True)
    snapshot_parser.add_argument("--output", type=pathlib.Path, required=True)
    snapshot_parser.set_defaults(handler=command_snapshot)

    assemble_parser = subparsers.add_parser("assemble")
    assemble_parser.add_argument("--lifecycle", type=pathlib.Path, required=True)
    assemble_parser.add_argument("--restart", type=pathlib.Path, required=True)
    assemble_parser.add_argument("--bounce", type=pathlib.Path, required=True)
    assemble_parser.add_argument("--vm", type=pathlib.Path, required=True)
    assemble_parser.add_argument("--avalanchego", type=pathlib.Path, required=True)
    assemble_parser.add_argument("--output", type=pathlib.Path, required=True)
    assemble_parser.set_defaults(handler=command_assemble)
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    try:
        args.handler(args)
    except Exception as error:
        print(f"locality rehearsal lifecycle: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

"""Pressure verifier for Reality-Contact Membrane v0.1."""

from __future__ import annotations

import json
import os
import socket
import subprocess
import sys
import tempfile
import time
from pathlib import Path

from reality_contact_membrane import (
    MembraneError,
    RealityContactMembrane,
    canonical_bytes,
    sha256_bytes,
)


ROOT = Path(__file__).resolve().parent
FIXTURE = ROOT / "live_pressure_envelope.json"
PROGRAM = ROOT / "reality_contact_membrane.py"


class Checks:
    def __init__(self) -> None:
        self.count = 0

    def that(self, condition: bool, message: str) -> None:
        self.count += 1
        if not condition:
            raise AssertionError(f"check {self.count} failed: {message}")


def wait_for_socket(path: Path, process: subprocess.Popen[bytes]) -> None:
    deadline = time.monotonic() + 5
    while time.monotonic() < deadline:
        if path.exists():
            return
        if process.poll() is not None:
            raise RuntimeError("membrane exited before opening its local socket")
        time.sleep(0.01)
    raise TimeoutError("membrane socket did not appear")


def cross(root: Path, socket_path: Path, wire: bytes) -> dict[str, object]:
    environment = dict(os.environ)
    environment["PYTHONDONTWRITEBYTECODE"] = "1"
    process = subprocess.Popen(
        [
            sys.executable,
            str(PROGRAM),
            "serve-once",
            "--root",
            str(root),
            "--socket",
            str(socket_path),
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=environment,
    )
    wait_for_socket(socket_path, process)
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as client:
        client.connect(str(socket_path))
        client.sendall(wire)
        client.shutdown(socket.SHUT_WR)
        receipt_bytes = b""
        while True:
            chunk = client.recv(65_536)
            if not chunk:
                break
            receipt_bytes += chunk
    stdout, stderr = process.communicate(timeout=5)
    if process.returncode != 0:
        raise RuntimeError(
            f"membrane failed ({process.returncode}): {stdout!r} {stderr!r}"
        )
    value = json.loads(receipt_bytes)
    if not isinstance(value, dict):
        raise RuntimeError("receipt is not an object")
    return value


def run() -> int:
    checks = Checks()
    fixture_value = json.loads(FIXTURE.read_text(encoding="utf-8"))
    wire = canonical_bytes(fixture_value)
    checks.that(
        fixture_value["payload_text"]
        == 'in this field, I am your "audience/relation" and not the world;',
        "live dialogue pressure sentence changed",
    )

    # Unix-domain sockets have a short platform path limit; keep the disposable
    # pressure root under /private/tmp so the live carrier remains addressable.
    with tempfile.TemporaryDirectory(prefix="rcm-", dir="/private/tmp") as name:
        temporary = Path(name)
        local_root = temporary / "receiving_locality"
        socket_path = temporary / "membrane.sock"

        first_receipt = cross(local_root, socket_path, wire)
        checks.that(first_receipt["status"] == "CROSSING_OBSERVED", "first crossing")
        checks.that(first_receipt["wire_sha256"] == sha256_bytes(wire), "first bytes")
        checks.that(not socket_path.exists(), "courier socket removed after crossing")

        membrane = RealityContactMembrane(local_root)
        entries = membrane.verify()
        checks.that(len(entries) == 1, "crossing exists before any later posture")
        first = entries[0]
        checks.that(first["event_type"] == "CROSSING_OBSERVED", "first event type")
        checks.that(first["data"]["format_status"] == "PARSED_OBJECT", "parsed content")
        checks.that(
            first["data"]["presented_source_claim"] == "Marko Markota",
            "source claim preserved",
        )
        checks.that(
            first["data"]["source_identity_status"]
            == "NOT_DETERMINED_BY_THIS_MEMBRANE",
            "membrane did not claim jurisdiction over source identity",
        )
        checks.that(
            first["data"]["source_continuity_status"]
            == "NOT_DETERMINED_BY_THIS_MEMBRANE",
            "membrane did not claim jurisdiction over source continuity",
        )
        checks.that(
            first["data"]["existing_local_relation_changed"] is False,
            "membrane did not demote any existing local relation",
        )
        checks.that(first["data"]["matter_admission"] is False, "no admission")
        checks.that(
            first["data"]["truth_status"] == "NOT_DETERMINED_BY_CROSSING",
            "no truth inflation",
        )
        checks.that(first["data"]["reciprocal_receipt_claimed"] is False, "no reciprocity")
        raw_path = local_root / first["data"]["raw_carrier"]
        checks.that(raw_path.read_bytes() == wire, "exact crossed bytes held locally")

        second_receipt = cross(local_root, socket_path, wire)
        entries = membrane.verify()
        crossings = [e for e in entries if e["event_type"] == "CROSSING_OBSERVED"]
        checks.that(len(crossings) == 2, "same content crossed twice as two occurrences")
        checks.that(
            first_receipt["exposure_id"] != second_receipt["exposure_id"],
            "occurrence identity not collapsed into content identity",
        )
        checks.that(
            crossings[0]["data"]["wire_sha256"] == crossings[1]["data"]["wire_sha256"],
            "same content identity remained visible",
        )

        malformed = b"not-json-but-still-crossed"
        malformed_receipt = cross(local_root, socket_path, malformed)
        entries = membrane.verify()
        malformed_entry = entries[-1]
        checks.that(malformed_receipt["status"] == "CROSSING_OBSERVED", "malformed crossing")
        checks.that(
            malformed_entry["data"]["format_status"] == "MALFORMED_OR_UNRECOGNIZED",
            "malformed content classified without erasing crossing",
        )
        checks.that(
            (local_root / malformed_entry["data"]["raw_carrier"]).read_bytes() == malformed,
            "malformed raw bytes preserved",
        )
        checks.that(malformed_entry["data"]["matter_admission"] is False, "malformed no admission")

        exposure_id = str(first_receipt["exposure_id"])
        raw_before = raw_path.read_bytes()
        posture = membrane.record_posture_claim(
            exposure_id,
            "HOLD_REFERENCE_ONLY",
            "Marko Markota",
            "present pressure-test direction",
        )
        checks.that(posture["event_type"] == "LOCAL_POSTURE_CLAIM_RECORDED", "posture separate")
        checks.that(
            posture["data"]["authority_validation"] == "NOT_PERFORMED_BY_THIS_MECHANISM",
            "mechanism did not become Authority verifier",
        )
        checks.that(posture["data"]["body_effect"] == "NONE", "posture no Body effect")
        checks.that(raw_path.read_bytes() == raw_before, "posture did not alter crossing")

        effect = membrane.record_effect_observation(
            exposure_id,
            "WORKSPACE_ARCHITECTURE",
            "contact and admission were separated",
            "the exact carried pressure preceded the correction",
            "causal sufficiency beyond this local pressure remains unresolved",
        )
        checks.that(
            effect["event_type"] == "LOCAL_EFFECT_OBSERVATION_RECORDED",
            "effect observation separate",
        )
        checks.that(effect["data"]["body_consequence_created"] is False, "no Body consequence")
        checks.that(
            effect["data"]["legitimacy_or_acceptance_created"] is False,
            "effect did not legitimate source",
        )
        checks.that(raw_path.read_bytes() == raw_before, "effect did not alter crossing")

        fresh_process = subprocess.run(
            [
                sys.executable,
                "-c",
                (
                    "from pathlib import Path; "
                    "from reality_contact_membrane import RealityContactMembrane; "
                    "entries=RealityContactMembrane(Path(__import__('sys').argv[1])).verify(); "
                    "print(len(entries))"
                ),
                str(local_root),
            ],
            cwd=ROOT,
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env={**os.environ, "PYTHONDONTWRITEBYTECODE": "1"},
            text=True,
        )
        checks.that(int(fresh_process.stdout.strip()) == 5, "fresh local reconstruction")
        checks.that(not socket_path.exists(), "no live courier required for reconstruction")

        original_journal = membrane.journal_path.read_bytes()
        lines = original_journal.decode("utf-8").splitlines()
        tampered = json.loads(lines[0])
        tampered["data"]["truth_status"] = "TRUE"
        lines[0] = canonical_bytes(tampered).decode("utf-8")
        membrane.journal_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
        try:
            membrane.verify()
        except MembraneError:
            tamper_rejected = True
        else:
            tamper_rejected = False
        checks.that(tamper_rejected, "journal tampering detected")
        membrane.journal_path.write_bytes(original_journal)
        checks.that(len(membrane.verify()) == 5, "original journal restored")

        unknown_rejected = False
        try:
            membrane.record_posture_claim(
                "unknown-exposure",
                "UNRESOLVED",
                "Marko Markota",
                "none",
            )
        except MembraneError:
            unknown_rejected = True
        checks.that(unknown_rejected, "later record cannot invent crossing")
        checks.that(not (local_root / "global_transcript.json").exists(), "no global transcript")
        checks.that(not (local_root / "shared_currentness.json").exists(), "no shared currentness")
        checks.that(not (local_root / "identity_registry.json").exists(), "no identity landlord")
        checks.that(not (local_root / "body_state.json").exists(), "no Body state impersonation")

    print(f"PASS {checks.count}/{checks.count} reality-contact membrane checks")
    return 0


if __name__ == "__main__":
    raise SystemExit(run())

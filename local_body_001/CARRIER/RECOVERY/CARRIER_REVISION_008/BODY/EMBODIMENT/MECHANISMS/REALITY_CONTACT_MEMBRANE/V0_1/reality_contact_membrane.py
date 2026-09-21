"""Recipient-local executable pressure for reality-contact without capture.

The membrane records bytes that cross its local socket before it interprets
them.  It may later append local posture claims or effect observations, but it
cannot turn those records into Body Matter, Authority, consequence, relation,
identity, or shared truth.

This implementation is constitutionally non-authoritative.  Its exact
Body-local admission and bounded activation posture are owned by
EMBODIMENT/STATE.json and BODY_BINDING.json, not by executable custody.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import secrets
import socket
import time
from pathlib import Path
from typing import Any


PRESSURE_ID = "REALITY-CONTACT-MEMBRANE-001"
SCHEMA_PREFIX = "reality-contact-membrane"
MAX_WIRE_BYTES = 1_048_576
POSTURES = {
    "ADMIT_AS_MATTER_CLAIM",
    "HOLD_REFERENCE_ONLY",
    "REFUSE_REQUESTED_MOTION",
    "WITHHOLD_POSTURE",
    "UNRESOLVED",
}


def canonical_bytes(value: Any) -> bytes:
    return json.dumps(
        value,
        ensure_ascii=False,
        separators=(",", ":"),
        sort_keys=True,
    ).encode("utf-8")


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def sha256_value(value: Any) -> str:
    return sha256_bytes(canonical_bytes(value))


def atomic_write(path: Path, value: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_name(f".{path.name}.tmp-{os.getpid()}-{secrets.token_hex(4)}")
    temporary.write_bytes(value)
    temporary.replace(path)


class MembraneError(RuntimeError):
    pass


class RealityContactMembrane:
    """Owns one receiver-local append-only account of crossings."""

    def __init__(self, root: Path):
        self.root = root
        self.raw_root = root / "raw_crossings"
        self.journal_path = root / "local_journal.jsonl"
        self.root.mkdir(parents=True, exist_ok=True)

    def _entries(self) -> list[dict[str, Any]]:
        if not self.journal_path.exists():
            return []
        entries: list[dict[str, Any]] = []
        for line in self.journal_path.read_text(encoding="utf-8").splitlines():
            if line.strip():
                value = json.loads(line)
                if not isinstance(value, dict):
                    raise MembraneError("journal entry is not an object")
                entries.append(value)
        return entries

    def verify(self) -> list[dict[str, Any]]:
        entries = self._entries()
        previous: str | None = None
        crossings: set[str] = set()
        for expected_sequence, entry in enumerate(entries, start=1):
            if entry.get("sequence") != expected_sequence:
                raise MembraneError("journal sequence mismatch")
            if entry.get("previous_entry_sha256") != previous:
                raise MembraneError("journal predecessor mismatch")
            claimed = entry.get("entry_sha256")
            unhashed = {k: v for k, v in entry.items() if k != "entry_sha256"}
            if claimed != sha256_value(unhashed):
                raise MembraneError("journal hash mismatch")
            data = entry.get("data", {})
            if entry.get("event_type") == "CROSSING_OBSERVED":
                exposure_id = data.get("exposure_id")
                if not isinstance(exposure_id, str) or exposure_id in crossings:
                    raise MembraneError("invalid or repeated exposure id")
                raw_path = self.root / data["raw_carrier"]
                raw = raw_path.read_bytes()
                if len(raw) != data.get("wire_bytes"):
                    raise MembraneError("raw crossing byte count mismatch")
                if sha256_bytes(raw) != data.get("wire_sha256"):
                    raise MembraneError("raw crossing hash mismatch")
                crossings.add(exposure_id)
            elif entry.get("event_type") in {
                "LOCAL_POSTURE_CLAIM_RECORDED",
                "LOCAL_EFFECT_OBSERVATION_RECORDED",
            }:
                if data.get("exposure_id") not in crossings:
                    raise MembraneError("later record precedes its crossing")
            previous = claimed
        return entries

    def _append(self, event_type: str, data: dict[str, Any]) -> dict[str, Any]:
        entries = self.verify()
        previous = entries[-1]["entry_sha256"] if entries else None
        entry_without_hash = {
            "schema": f"{SCHEMA_PREFIX}.local-journal-entry.v0.1",
            "pressure_id": PRESSURE_ID,
            "sequence": len(entries) + 1,
            "previous_entry_sha256": previous,
            "recorded_at_unix_ns": time.time_ns(),
            "event_type": event_type,
            "data": data,
        }
        entry = {
            **entry_without_hash,
            "entry_sha256": sha256_value(entry_without_hash),
        }
        with self.journal_path.open("a", encoding="utf-8") as handle:
            handle.write(canonical_bytes(entry).decode("utf-8") + "\n")
            handle.flush()
            os.fsync(handle.fileno())
        return entry

    def observe_wire_crossing(self, wire: bytes) -> dict[str, Any]:
        if len(wire) > MAX_WIRE_BYTES:
            raise MembraneError("wire input exceeds bounded local capacity")
        exposure_id = f"rcm-{time.time_ns()}-{secrets.token_hex(6)}"
        raw_relative = Path("raw_crossings") / f"{exposure_id}.bin"
        atomic_write(self.root / raw_relative, wire)

        decoded: dict[str, Any] | None = None
        format_status = "MALFORMED_OR_UNRECOGNIZED"
        try:
            candidate = json.loads(wire.decode("utf-8"))
            if isinstance(candidate, dict):
                decoded = candidate
                format_status = "PARSED_OBJECT"
        except (UnicodeDecodeError, json.JSONDecodeError):
            pass

        data = {
            "exposure_id": exposure_id,
            "raw_carrier": raw_relative.as_posix(),
            "wire_bytes": len(wire),
            "wire_sha256": sha256_bytes(wire),
            "format_status": format_status,
            "presented_source_claim": decoded.get("source_claim") if decoded else None,
            "presented_source_continuity_claim": (
                decoded.get("source_continuity_claim") if decoded else None
            ),
            "source_identity_status": "NOT_DETERMINED_BY_THIS_MEMBRANE",
            "source_continuity_status": "NOT_DETERMINED_BY_THIS_MEMBRANE",
            "existing_local_relation_changed": False,
            "target_claim": decoded.get("target_claim") if decoded else None,
            "requested_effect": decoded.get("requested_effect") if decoded else None,
            "crossing_effect_ceiling": "RECIPIENT_LOCAL_OBSERVATION_ONLY",
            "matter_admission": False,
            "truth_status": "NOT_DETERMINED_BY_CROSSING",
            "reciprocal_receipt_claimed": False,
        }
        return self._append("CROSSING_OBSERVED", data)

    def record_posture_claim(
        self,
        exposure_id: str,
        posture: str,
        attributed_bearer: str,
        presented_basis: str,
    ) -> dict[str, Any]:
        if posture not in POSTURES:
            raise MembraneError(f"unsupported posture claim: {posture}")
        self._require_crossing(exposure_id)
        return self._append(
            "LOCAL_POSTURE_CLAIM_RECORDED",
            {
                "exposure_id": exposure_id,
                "posture_claim": posture,
                "attributed_bearer": attributed_bearer,
                "presented_basis": presented_basis,
                "authority_validation": "NOT_PERFORMED_BY_THIS_MECHANISM",
                "body_effect": "NONE",
            },
        )

    def record_effect_observation(
        self,
        exposure_id: str,
        target: str,
        observed_change: str,
        attribution_basis: str,
        uncertainty: str,
    ) -> dict[str, Any]:
        self._require_crossing(exposure_id)
        return self._append(
            "LOCAL_EFFECT_OBSERVATION_RECORDED",
            {
                "exposure_id": exposure_id,
                "target": target,
                "observed_change": observed_change,
                "attribution_basis": attribution_basis,
                "uncertainty": uncertainty,
                "body_consequence_created": False,
                "legitimacy_or_acceptance_created": False,
            },
        )

    def _require_crossing(self, exposure_id: str) -> dict[str, Any]:
        for entry in self.verify():
            if (
                entry["event_type"] == "CROSSING_OBSERVED"
                and entry["data"]["exposure_id"] == exposure_id
            ):
                return entry
        raise MembraneError("unknown exposure id")


def serve_once(root: Path, socket_path: Path) -> None:
    if socket_path.exists():
        socket_path.unlink()
    socket_path.parent.mkdir(parents=True, exist_ok=True)
    membrane = RealityContactMembrane(root)
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as server:
        server.bind(str(socket_path))
        server.listen(1)
        connection, _ = server.accept()
        with connection:
            chunks: list[bytes] = []
            total = 0
            while True:
                chunk = connection.recv(65_536)
                if not chunk:
                    break
                total += len(chunk)
                if total > MAX_WIRE_BYTES:
                    receipt = {"status": "REFUSED_OVERSIZE", "recorded": False}
                    connection.sendall(canonical_bytes(receipt))
                    return
                chunks.append(chunk)
            entry = membrane.observe_wire_crossing(b"".join(chunks))
            receipt = {
                "schema": f"{SCHEMA_PREFIX}.local-receipt.v0.1",
                "status": "CROSSING_OBSERVED",
                "exposure_id": entry["data"]["exposure_id"],
                "wire_sha256": entry["data"]["wire_sha256"],
                "effect_ceiling": "RECIPIENT_LOCAL_OBSERVATION_ONLY",
            }
            connection.sendall(canonical_bytes(receipt))
    socket_path.unlink(missing_ok=True)


def main() -> int:
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command", required=True)
    serve_parser = subparsers.add_parser("serve-once")
    serve_parser.add_argument("--root", required=True, type=Path)
    serve_parser.add_argument("--socket", required=True, type=Path)
    args = parser.parse_args()
    if args.command == "serve-once":
        serve_once(args.root, args.socket)
        return 0
    return 2


if __name__ == "__main__":
    raise SystemExit(main())

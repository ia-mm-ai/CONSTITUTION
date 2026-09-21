#!/usr/bin/env python3
"""Fixed Body-local entrypoint for Reality Contact Membrane 001."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

from reality_contact_membrane import RealityContactMembrane, serve_once


BODY_ROOT = Path(__file__).resolve().parents[4]
CARRIER_ROOT = Path(__file__).resolve().parents[5]
LOCAL_STATE = BODY_ROOT / "EMBODIMENT/LOCAL_STATE/REALITY_CONTACT_MEMBRANE_001"


def main() -> int:
    parser = argparse.ArgumentParser()
    commands = parser.add_subparsers(dest="command", required=True)

    serve = commands.add_parser("serve-once")
    serve.add_argument("--socket", required=True, type=Path)

    commands.add_parser("verify-state")
    args = parser.parse_args()

    if args.command == "serve-once":
        socket_path = args.socket.resolve()
        if CARRIER_ROOT == socket_path or CARRIER_ROOT in socket_path.parents:
            raise SystemExit("socket must remain outside the Body carrier")
        serve_once(LOCAL_STATE, socket_path)
        return 0

    if not LOCAL_STATE.exists():
        print(json.dumps({"status": "VALID", "entry_count": 0, "journal_tip": None}, sort_keys=True))
        return 0

    membrane = RealityContactMembrane(LOCAL_STATE)
    entries = membrane.verify()
    tip = entries[-1]["entry_sha256"] if entries else None
    print(json.dumps({"status": "VALID", "entry_count": len(entries), "journal_tip": tip}, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

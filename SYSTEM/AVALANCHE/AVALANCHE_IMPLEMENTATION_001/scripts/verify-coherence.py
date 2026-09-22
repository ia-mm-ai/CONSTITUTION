#!/usr/bin/env python3
"""Fail closed unless the implementation has one operative identity and topology."""

import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CANONICAL_VM_ID = "cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8"
HISTORICAL_VM_ID = "cQYXygFUVutQucm4s8pr8M51sRdS1" + "UfrTbpMdEgpRYt2JvEzr"


def main():
    go_modules = sorted(path.relative_to(ROOT).as_posix() for path in ROOT.rglob("go.mod"))
    if go_modules != ["vm/go.mod", "vm/rehearsal/go.mod"]:
        raise ValueError(f"unexpected Go module topology: {go_modules}")
    if list(ROOT.glob("*.go")):
        raise ValueError("root-level Go source competes with vm/")

    contracts = sorted(path.relative_to(ROOT).as_posix() for path in ROOT.rglob("operations.json"))
    if contracts != ["vm/protocol/operations.json"]:
        raise ValueError(f"unexpected operation-contract topology: {contracts}")

    historical_hits = []
    for path in ROOT.rglob("*"):
        if path != Path(__file__) and path.is_file() and HISTORICAL_VM_ID.encode() in path.read_bytes():
            historical_hits.append(path.relative_to(ROOT).as_posix())
    if historical_hits != ["records/DERIVATION_COHERENCE_001.json"]:
        raise ValueError(f"historical VM ID escaped bounded coherence record: {historical_hits}")

    config = json.loads((ROOT / "field/examples/config.template.json").read_bytes())
    compatibility = json.loads((ROOT / "compat/COMPATIBILITY_001.json").read_bytes())
    genesis = json.loads((ROOT / "vm/genesis/PRESENCE_AVALANCHE_RUNTIME_GENESIS_TEMPLATE_001.json").read_bytes())
    if config["expected_vm_id"] != CANONICAL_VM_ID:
        raise ValueError("default FIELD configuration does not target the canonical VM")
    if compatibility["vm_id_derivation"]["result"] != CANONICAL_VM_ID:
        raise ValueError("compatibility declaration does not derive the canonical VM")
    if genesis["domain"]["id"] != compatibility["vm_protocol"]:
        raise ValueError("genesis protocol does not resolve through the canonical VM derivation")

    build = (ROOT / "scripts/build.sh").read_text()
    if 'cd "${root}/vm"' not in build or "source_dir" in build:
        raise ValueError("default build is not uniquely bound to vm/")

    print(json.dumps({
        "status": "COHERENT_SINGLE_IMPLEMENTATION",
        "vm_id": CANONICAL_VM_ID,
        "go_modules": go_modules,
        "operation_contracts": contracts,
        "historical_vm_id_locations": historical_hits,
        "default_build": "vm/",
        "default_field_vm_id": config["expected_vm_id"],
        "genesis_protocol": genesis["domain"]["id"],
    }, indent=2))


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Verify exact current Source, canonical STATE, and predecessor carrier bytes."""

import hashlib
import json
from pathlib import Path


IMPLEMENTATION = Path(__file__).resolve().parents[1]
REPOSITORY = IMPLEMENTATION.parents[2]
BINDING = IMPLEMENTATION / "vm/protocol/binding.json"


def digest(path):
    if path.is_symlink() or not path.is_file():
        raise ValueError(f"not an ordinary file: {path}")
    data = path.read_bytes()
    return {"sha256": hashlib.sha256(data).hexdigest(), "bytes": len(data)}


def verify(repository=REPOSITORY, binding_path=BINDING):
    binding = json.loads(binding_path.read_bytes())
    for group in ("source", "predecessors"):
        for relative, expected in binding[group].items():
            actual = digest(repository / relative)
            if actual != expected:
                raise ValueError(f"immutable binding mismatch: {relative}")
    source_paths = {
        p.relative_to(repository).as_posix()
        for p in (repository / "SOURCE").rglob("*")
        if p.is_file() or p.is_symlink()
    }
    if source_paths != set(binding["source"]):
        raise ValueError("Source file inventory changed")
    human = repository / "SOURCE/CONSTITUTION_0()1.md"
    machine = json.loads((repository / "SOURCE/CONSTITUTION_0()1.json").read_bytes())
    internal = machine["binding"]
    if (
        internal["human_path"] != human.name
        or internal["hash_algorithm"] != "sha256"
        or internal["human_sha256"] != digest(human)["sha256"]
        or internal["human_byte_length"] != human.stat().st_size
        or internal["hash_scope"] != "exact_file_bytes"
    ):
        raise ValueError("machine Constitution exact-human-byte binding failed")
    inventory = {}
    for path in sorted((repository / "STATE").rglob("*")):
        if path.is_symlink():
            raise ValueError(f"symlink in canonical STATE: {path}")
        if path.is_file():
            relative = path.relative_to(repository).as_posix()
            if path.suffix == ".zip":
                if relative not in binding["predecessors"]:
                    raise ValueError(f"unbound STATE archive: {relative}")
            else:
                inventory[relative] = digest(path)
    encoded = "".join(
        f"{entry['sha256']}  {path}\n" for path, entry in sorted(inventory.items())
    ).encode("utf-8")
    state = binding["canonical_state"]
    if (
        len(inventory) != state["file_count"]
        or sum(entry["bytes"] for entry in inventory.values()) != state["total_bytes"]
        or hashlib.sha256(encoded).hexdigest() != state["inventory_sha256"]
    ):
        raise ValueError("canonical STATE inventory binding failed")
    return {
        "status": "EXACT_BYTE_BINDINGS_VERIFIED",
        "starting_commit": binding["starting_commit"],
        "source": binding["source"],
        "predecessors": binding["predecessors"],
        "canonical_state": state,
        "effect": "INTEGRITY_ONLY_NOT_CONSTITUTIONAL_EFFECT",
    }


if __name__ == "__main__":
    print(json.dumps(verify(), indent=2))

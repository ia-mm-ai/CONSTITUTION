#!/usr/bin/env python3
import hashlib
import json
import pathlib
import sys


IMPLEMENTATION_ROOT = pathlib.Path(__file__).resolve().parents[1]
REPOSITORY_ROOT = IMPLEMENTATION_ROOT.parents[2]
BINDING_PATH = (
    IMPLEMENTATION_ROOT / "bindings" / "SOURCE_STATE_BINDING_001.json"
)


def sha256_file(path):
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def tree_sha256(path):
    digest = hashlib.sha256()
    files = sorted(candidate for candidate in path.rglob("*") if candidate.is_file())
    for candidate in files:
        relative = candidate.relative_to(REPOSITORY_ROOT).as_posix()
        digest.update(f"{sha256_file(candidate)}  {relative}\n".encode("utf-8"))
    return digest.hexdigest()


def verify_file(record):
    path = REPOSITORY_ROOT / record["path"]
    if not path.is_file():
        raise ValueError(f"missing bound file: {record['path']}")
    if path.stat().st_size != record["bytes"]:
        raise ValueError(f"byte length mismatch: {record['path']}")
    if sha256_file(path) != record["sha256"]:
        raise ValueError(f"SHA-256 mismatch: {record['path']}")


def main():
    binding = json.loads(BINDING_PATH.read_text(encoding="utf-8"))
    if binding["schema"] != "PRESENCE_AVALANCHE_SOURCE_STATE_BINDING_001":
        raise ValueError("unsupported binding schema")
    for record in binding["source"]["files"]:
        verify_file(record)
    for record in binding["state"]["predecessor_archives"]:
        verify_file(record)
    for directory, expected in (
        ("SOURCE", binding["source"]["tree_sha256"]),
        ("STATE", binding["state"]["tree_sha256"]),
    ):
        actual = tree_sha256(REPOSITORY_ROOT / directory)
        if actual != expected:
            raise ValueError(f"{directory} tree SHA-256 mismatch: {actual}")
    print("Source, STATE, and predecessor archive bindings verified")


if __name__ == "__main__":
    try:
        main()
    except (KeyError, OSError, ValueError, json.JSONDecodeError) as error:
        print(f"binding verification failed: {error}", file=sys.stderr)
        raise SystemExit(1)

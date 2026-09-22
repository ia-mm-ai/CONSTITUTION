"""Revision and byte checks; core semantics remain with STATE's checker."""

import argparse
import hashlib
import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path, PurePosixPath

from jsonschema import Draft202012Validator


ROOT = Path(__file__).resolve().parents[1]
CHECKER = "STATE/LINEAGE/CONFORMANCE.py"
HUMAN = "SOURCE/CONSTITUTION_0()1.md"
MACHINE = "SOURCE/CONSTITUTION_0()1.json"
CEILING = "Byte correspondence and checked representation only; no adoption, Authority, or present CSC/DCR."


def load_json(data):
    def pairs(items):
        result = {}
        for key, value in items:
            if key in result:
                raise ValueError(f"Duplicate JSON key: {key}")
            result[key] = value
        return result

    def reject(value):
        raise ValueError(f"Non-finite JSON number: {value}")

    return json.loads(data.decode("utf-8"), object_pairs_hook=pairs, parse_constant=reject)


def encode(value):
    return (json.dumps(value, ensure_ascii=False, indent=2, allow_nan=False) + "\n").encode("utf-8")


def binding(data):
    return {"sha256": hashlib.sha256(data).hexdigest(), "byte_length": len(data)}


def safe_path(name):
    if not isinstance(name, str) or not name or "\\" in name:
        raise ValueError("Invalid relative path")
    path = PurePosixPath(name)
    if path.is_absolute() or path.as_posix() != name or any(
        part in (".", "..", ".git") or any(ord(c) < 32 for c in part)
        for part in name.split("/")
    ):
        raise ValueError(f"Unsafe path: {name}")
    return path


def read_file(root, name):
    path = root
    for part in safe_path(name).parts:
        path = path / part
        if path.is_symlink():
            raise ValueError(f"Symlink not permitted: {name}")
    if not path.is_file():
        raise ValueError(f"Missing regular file: {name}")
    return path.read_bytes()


def revision_files(repository, revision):
    if not re.fullmatch(r"(?:[0-9a-f]{40}|[0-9a-f]{64})", revision):
        raise ValueError("Select a full Git commit ID, not a branch, tag, or abbreviated ID")

    def git(*args):
        return subprocess.check_output(["git", "-C", str(repository), *args], stderr=subprocess.PIPE)

    if git("cat-file", "-t", revision).strip() != b"commit":
        raise ValueError("Revision must identify a commit")
    files = {}
    tree = git("ls-tree", "-r", "-z", revision, "--", "SOURCE", "STATE", "SYSTEM", "SURFACE")
    for entry in tree.split(b"\0"):
        if not entry:
            continue
        metadata, raw_name = entry.split(b"\t", 1)
        mode, kind, object_id = metadata.split()
        name = raw_name.decode("utf-8")
        safe_path(name)
        if mode not in (b"100644", b"100755") or kind != b"blob":
            raise ValueError(f"Only regular Git blobs may be exported: {name}")
        files[name] = git("cat-file", "blob", object_id.decode("ascii"))
    return files


def walk(value, pointer=""):
    if isinstance(value, dict):
        yield value, pointer
        for key, child in value.items():
            escaped = key.replace("~", "~0").replace("/", "~1")
            yield from walk(child, f"{pointer}/{escaped}")
    elif isinstance(value, list):
        for i, child in enumerate(value):
            yield from walk(child, f"{pointer}/{i}")


def inventory(files):
    declaration = load_json(files["SURFACE/surface.json"])
    if set(declaration) != {"point", "resources", "operations"}:
        raise ValueError("surface.json must declare point, resources, and operations")
    resources, references = [], {}
    paths = set()
    for item in declaration["resources"]:
        if set(item) != {"path", "ref"} or not item["path"].startswith("../"):
            raise ValueError("Resource paths are relative to SURFACE/surface.json")
        name = item["path"][3:]
        safe_path(name)
        if not name.startswith(("SOURCE/", "STATE/")) or not name.endswith((".md", ".json")):
            raise ValueError(f"Not selected core material: {name}")
        if name in paths or item["ref"] in references:
            raise ValueError(f"Duplicate resource: {name}")
        paths.add(name)
        data = files[name]
        resources.append({
            "ref": item["ref"], "path": name,
            "media_type": "application/json" if name.endswith(".json") else "text/markdown; charset=utf-8",
            **binding(data),
        })
        references[item["ref"]] = {"path": name, "pointer": ""}
        if name.endswith(".json"):
            document = load_json(data)
            identifier = document.get("$id", document.get("id"))
            if identifier and identifier != item["ref"]:
                raise ValueError(f"Resource reference differs from core identifier: {name}")
    for node, pointer in walk(load_json(files[MACHINE])):
        if "ref" in node:
            ref = node["ref"]
            if ref in references:
                raise ValueError(f"Duplicate Source reference: {ref}")
            references[ref] = {"path": MACHINE, "pointer": pointer}
    if HUMAN not in paths or MACHINE not in paths or "STATE/LINEAGE/ORIGIN.json" not in paths:
        raise ValueError("The bound Source pair and LINEAGE origin must be selected")
    for path in declaration["point"]["provenance"]:
        if not path.startswith("../") or path[3:] not in paths:
            raise ValueError(f"Unselected provenance: {path}")
    expected = {"read": "serve.read_resource", "resolve": "serve.resolve_reference", "check": "serve.check_resource"}
    if {op["id"]: op["handler"] for op in declaration["operations"]} != expected:
        raise ValueError("Only fixed read, resolve, and check handlers are supported")
    if len(declaration["operations"]) != len(expected):
        raise ValueError("Duplicate operation")
    for op in declaration["operations"]:
        parameter = "ref" if op["id"] == "resolve" else "resource"
        if op["method"] != "GET" or op["route"] != f"/api/{op['id']}":
            raise ValueError("Unsupported operation route or method")
        contract = {
            "type": "object", "additionalProperties": False, "required": [parameter],
            "properties": {parameter: {"type": "string", "minLength": 1, "maxLength": 512}},
        }
        if op["input"] != contract:
            raise ValueError("Operation input must retain its bounded contract")
        Draft202012Validator.check_schema(op["input"])
    return declaration, resources, references


def check_material(files):
    machine = load_json(files[MACHINE])
    expected = machine["binding"]
    if expected["hash_algorithm"] != "sha256" or expected["hash_scope"] != "exact_file_bytes":
        raise ValueError("Unsupported Source binding")
    if expected["human_path"] != PurePosixPath(HUMAN).name or binding(files[HUMAN]) != {
        "sha256": expected["human_sha256"], "byte_length": expected["human_byte_length"]
    }:
        raise ValueError("Source human/machine byte binding mismatch")
    origin = load_json(files["STATE/LINEAGE/ORIGIN.json"])
    for field, name in (("human_content_binding", HUMAN), ("machine_content_binding", MACHINE)):
        expected = origin[field]
        if expected["path"] != f"../../{name}" or expected["hash_algorithm"] != "sha256" or expected["hash_scope"] != "exact_file_bytes":
            raise ValueError(f"Unsupported LINEAGE binding: {field}")
        if binding(files[name]) != {key: expected[key] for key in ("sha256", "byte_length")}:
            raise ValueError(f"LINEAGE byte binding mismatch: {name}")
    declaration, resources, references = inventory(files)
    schemas = {}
    for item in resources:
        name = item["path"]
        if not name.endswith(".json"):
            continue
        document = load_json(files[name])
        if name.endswith(".schema.json"):
            Draft202012Validator.check_schema(document)
            schemas[document["$id"]] = document
        for node, _ in walk(document):
            for key in ("source", "x-source"):
                source = node.get(key)
                if isinstance(source, dict) and isinstance(source.get("refs"), list):
                    for ref in source["refs"]:
                        if ref not in references or references[ref]["path"] != MACHINE:
                            raise ValueError(f"Unbound Source reference in {name}: {ref}")
    for document in schemas.values():
        for node, _ in walk(document):
            ref = node.get("$ref", "")
            if ref and not ref.startswith("#") and ref.split("#")[0] not in schemas:
                raise ValueError(f"Schema not selected for offline resolution: {ref}")
    return declaration, resources, references


def check_revision(repository, revision):
    files = revision_files(repository, revision)
    declaration, resources, references = check_material(files)
    conformance = {"status": "UNAVAILABLE", "checker": CHECKER,
                   "reason": "The selected revision has no core conformance checker; semantic rules and vectors were not evaluated."}
    if CHECKER in files:
        # Execute only code explicitly selected from local Git, never from an export.
        with tempfile.TemporaryDirectory(prefix="presence-core-") as directory:
            root = Path(directory)
            for name, data in files.items():
                if name.startswith(("SOURCE/", "STATE/")):
                    target = root / name
                    target.parent.mkdir(parents=True, exist_ok=True)
                    target.write_bytes(data)
            process = subprocess.run(
                [sys.executable, str(root / CHECKER)], cwd=root, capture_output=True,
                text=True, timeout=60, check=False,
            )
        conformance = {"status": "PASSED" if process.returncode == 0 else "FAILED",
                       "checker": CHECKER, "returncode": process.returncode,
                       "stdout": process.stdout, "stderr": process.stderr}
    report = {
        "revision": revision,
        "status": "PASSED" if conformance["status"] == "PASSED" else
                  "FAILED" if conformance["status"] == "FAILED" else "INCOMPLETE",
        "checked": ["Source exact bytes", "LINEAGE Source bindings", "selected resource references",
                    "schema definitions and offline schema dependencies"],
        "resources": [item["path"] for item in resources],
        "conformance": conformance, "effect_ceiling": CEILING,
    }
    return files, declaration, resources, references, report


def publication(repository, revision, allow_incomplete=False):
    files, declaration, resources, references, report = check_revision(repository, revision)
    if report["status"] == "FAILED" or (report["status"] != "PASSED" and not allow_incomplete):
        raise ValueError("Core conformance did not pass; restore the checker or explicitly request --allow-incomplete for a labelled preview")
    output = {item["path"]: files[item["path"]] for item in resources}
    output["index.html"] = files["SURFACE/index.html"]
    output["index.json"] = encode({
        "format": "presence.index.v1", "revision": revision, "point": declaration["point"],
        "resources": resources, "references": references, "operations": declaration["operations"],
        "verification": report,
    })
    output["manifest.json"] = encode({
        "format": "presence.manifest.v1", "revision": revision,
        "files": {name: binding(data) for name, data in sorted(output.items())},
        "verification": report,
        "effect_ceiling": "Exact bytes, not authenticated origin. The manifest does not hash itself.",
    })
    return output, report


def check_export(root, expected_manifest):
    if not re.fullmatch(r"[0-9a-f]{64}", expected_manifest):
        raise ValueError("Supply the trusted manifest SHA-256")
    raw_manifest = read_file(root, "manifest.json")
    if binding(raw_manifest)["sha256"] != expected_manifest:
        raise ValueError("Manifest differs from the independently supplied digest")
    manifest = load_json(raw_manifest)
    if manifest["format"] != "presence.manifest.v1" or not re.fullmatch(
        r"(?:[0-9a-f]{40}|[0-9a-f]{64})", manifest["revision"]
    ):
        raise ValueError("Unsupported manifest format or revision")
    files = {"manifest.json": raw_manifest}
    for name, expected in manifest["files"].items():
        safe_path(name)
        if name not in ("index.html", "index.json") and not (
            name.startswith(("SOURCE/", "STATE/")) and name.endswith((".md", ".json"))
        ):
            raise ValueError(f"Unexpected exported file: {name}")
        data = read_file(root, name)
        if binding(data) != expected:
            raise ValueError(f"Export binding mismatch: {name}")
        files[name] = data
    actual = set()
    for path in root.rglob("*"):
        if path.is_symlink():
            raise ValueError("Symlinks are not permitted in an export")
        if not path.is_dir():
            actual.add(path.relative_to(root).as_posix())
    if actual != set(files):
        raise ValueError("Export has missing or unlisted files")
    index = load_json(files["index.json"])
    if index["format"] != "presence.index.v1" or index["revision"] != manifest["revision"]:
        raise ValueError("Index and manifest edition mismatch")
    if index["verification"] != manifest["verification"] or index["verification"]["revision"] != index["revision"]:
        raise ValueError("Verification report edition mismatch")
    declaration = {
        "point": index["point"], "operations": index["operations"],
        "resources": [{"path": "../" + item["path"], "ref": item["ref"]} for item in index["resources"]],
    }
    material = {**files, "SURFACE/surface.json": encode(declaration)}
    _, resources, references = check_material(material)
    if resources != index["resources"] or references != index["references"]:
        raise ValueError("Generated inventory does not match selected material")
    if set(files) != {item["path"] for item in resources} | {"index.html", "index.json", "manifest.json"}:
        raise ValueError("Manifest and resource inventory disagree")
    return files, {
        "revision": index["revision"], "status": "INTEGRITY_VERIFIED",
        "checked": sorted(files), "manifest_sha256": expected_manifest,
        "conformance": {"status": "NOT_RERUN", "recorded": index["verification"]["conformance"]},
        "effect_ceiling": CEILING,
    }


def write_fresh(destination, files):
    if destination.exists() or destination.is_symlink():
        raise ValueError("Destination must not already exist")
    destination.mkdir(parents=False)
    try:
        for name, data in files.items():
            target = destination / safe_path(name)
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(data)
    except BaseException:
        import shutil
        shutil.rmtree(destination)
        raise


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--revision", help="Full, locally available, trusted Git commit ID")
    source.add_argument("--export", type=Path, help="Portable edition directory; never executes its code")
    parser.add_argument("--repository", type=Path, default=ROOT)
    parser.add_argument("--manifest-sha256", help="Independently trusted digest, required with --export")
    args = parser.parse_args()
    try:
        if args.export:
            if not args.manifest_sha256:
                raise ValueError("--export requires --manifest-sha256")
            _, report = check_export(args.export, args.manifest_sha256)
        else:
            *_, report = check_revision(args.repository, args.revision)
        print(encode(report).decode(), end="")
        return 0 if report["status"] in ("PASSED", "INTEGRITY_VERIFIED") else 2
    except (ValueError, KeyError, TypeError, OSError, subprocess.SubprocessError) as error:
        print(encode({"status": "FAILED", "error": str(error)}).decode(), end="")
        return 1


if __name__ == "__main__":
    sys.exit(main())

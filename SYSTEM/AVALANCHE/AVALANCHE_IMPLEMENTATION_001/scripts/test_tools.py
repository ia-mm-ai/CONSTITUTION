import importlib.util
import json
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest.mock import patch


ROOT = Path(__file__).resolve().parents[1]


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, ROOT / "scripts" / filename)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


binding = load("binding", "verify-binding.py")
preflight = load("preflight", "validator-preflight.py")


class BindingTests(unittest.TestCase):
    def test_current_exact_bindings(self):
        self.assertEqual(binding.verify()["status"], "EXACT_BYTE_BINDINGS_VERIFIED")

    def test_mutations_are_rejected(self):
        with tempfile.TemporaryDirectory(prefix="presence-binding-") as directory:
            root = Path(directory)
            for name in ("SOURCE", "STATE"):
                shutil.copytree(binding.REPOSITORY / name, root / name)
            for name in (
                "SOURCE/CONSTITUTION_0()1.md",
                "SOURCE/CONSTITUTION_0()1.json",
                "STATE/STATE.json",
                "STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip",
                "STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip",
            ):
                with self.subTest(path=name):
                    path = root / name
                    original = path.read_bytes()
                    path.write_bytes(original + b"\n")
                    with self.assertRaises(ValueError):
                        binding.verify(root)
                    path.write_bytes(original)
            added = root / "STATE/unbound.json"
            added.write_text("{}")
            with self.assertRaises(ValueError):
                binding.verify(root)
            added.unlink()
            added = root / "STATE/unbound.zip"
            added.write_bytes(b"not a bound carrier")
            with self.assertRaises(ValueError):
                binding.verify(root)
            added.unlink()
            added = root / "SOURCE/unbound.txt"
            added.write_text("not Source")
            with self.assertRaises(ValueError):
                binding.verify(root)

    def test_symlink_rejected(self):
        with tempfile.TemporaryDirectory(prefix="presence-binding-") as directory:
            path = Path(directory) / "link"
            path.symlink_to(binding.BINDING)
            with self.assertRaises(ValueError):
                binding.digest(path)


class PreflightTests(unittest.TestCase):
    def test_binary_version_profile(self):
        vm = {
            "vm_id": "test-vm-id",
            "version": "1.0.0",
            "implementation": "AVALANCHE_IMPLEMENTATION_001",
            "protocol": "PRESENCE_AVALANCHE_VM_001",
            "rpcchainvm": 46,
        }
        node = {"application": "avalanchego/1.15.0", "rpcchainvm": 46, "go": "1.25.13"}
        arguments = [
            "preflight", "--plugin", "/tmp/test-vm-id", "--plugin-sha256", "a" * 64,
            "--avalanchego", "/tmp/test-node", "--avalanchego-sha256", "b" * 64,
            "--vm-id", "test-vm-id",
        ]
        info = "go1.25.13\n\tdep\tgoogle.golang.org/grpc\tv1.83.2\th1:test\n"
        for valid in (True, False):
            with self.subTest(valid_protocol=valid):
                candidate = {**vm, "rpcchainvm": 46 if valid else 45}
                with (
                    patch("sys.argv", arguments),
                    patch.object(preflight, "checked_binary", side_effect=[Path("/tmp/test-vm-id"), Path("/tmp/test-node")]),
                    patch.object(preflight, "version", side_effect=[candidate, node]),
                    patch.object(preflight.subprocess, "check_output", return_value=info),
                    patch("builtins.print"),
                ):
                    if valid:
                        preflight.main()
                    else:
                        with self.assertRaises(ValueError):
                            preflight.main()

    def test_requires_digest_and_executable(self):
        with tempfile.TemporaryDirectory(prefix="presence-preflight-") as directory:
            binary = Path(directory) / "binary"
            binary.write_bytes(b"not executable")
            with self.assertRaises(ValueError):
                preflight.checked_binary(binary, "missing")
            with self.assertRaises(ValueError):
                preflight.checked_binary(binary, "0" * 64)
            binary.chmod(0o700)
            with self.assertRaises(ValueError):
                preflight.checked_binary(binary, "0" * 64)
            expected = preflight.hashlib.sha256(binary.read_bytes()).hexdigest()
            self.assertEqual(preflight.checked_binary(binary, expected), binary)

    def test_administrative_schema_is_closed_and_public_only(self):
        from jsonschema import Draft202012Validator
        schema = json.loads((ROOT / "vm" / "genesis" / "administrative-input.schema.json").read_bytes())
        Draft202012Validator.check_schema(schema)
        validator = Draft202012Validator(schema)
        public = {
            "schema": "PRESENCE_AVALANCHE_DEPLOYMENT_DESCRIPTOR_001",
            "genesis_id": "test-genesis",
            "locality_id": "test-locality",
            "purpose": "schema test only",
            "source_reference": "test-reference",
            "source_sha256": "a" * 64,
        }
        self.assertTrue(validator.is_valid(public))
        self.assertFalse(validator.is_valid({**public, "private_key": "not permitted"}))
        self.assertFalse(validator.is_valid({**public, "locality_id": ""}))


if __name__ == "__main__":
    unittest.main()

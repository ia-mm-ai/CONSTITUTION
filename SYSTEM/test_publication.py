"""Portable edition regressions; no historical carrier code is executed."""

import copy
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import verify


class PublicationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.files = {
            path.relative_to(verify.ROOT).as_posix(): path.read_bytes()
            for directory in ("SOURCE", "STATE", "SYSTEM", "SURFACE")
            for path in (verify.ROOT / directory).rglob("*")
            if path.is_file() and "__pycache__" not in path.parts
        }
        cls.revision = "1" * 40
        with patch.object(verify, "revision_files", return_value=cls.files):
            cls.edition, cls.report = verify.publication(verify.ROOT, cls.revision)
        cls.digest = verify.binding(cls.edition["manifest.json"])["sha256"]

    def changed_document(self, path, change):
        files = dict(self.files)
        document = verify.load_json(files[path])
        change(document)
        files[path] = verify.encode(document)
        return files

    def test_system_checker_and_complete_inventory(self):
        self.assertEqual(self.report["status"], "PASSED")
        self.assertEqual(self.report["conformance"]["checker"], "SYSTEM/conformance.py")
        self.assertEqual(self.report["conformance"]["checker_binding"],
                         verify.binding(self.files["SYSTEM/conformance.py"]))
        self.assertTrue(all(not name.endswith(".py") for name in self.edition))
        index = verify.load_json(self.edition["index.json"])
        self.assertIn("urn:presence:lineage:implementation:vm003-medium001:1", index["references"])
        self.assertEqual(sum(r["media_type"] == "application/zip" for r in index["resources"]), 2)
        for item in index["resources"]:
            self.assertEqual(self.edition[item["path"]], self.files[item["path"]])

    def test_missing_account_companion_carrier_or_vector_is_rejected(self):
        for target in (
            "STATE/LINEAGE/IMPLEMENTATIONS/VM003-MEDIUM001.json",
            "STATE/LINEAGE/IMPLEMENTATIONS/VM003-MEDIUM001.md",
            "STATE/LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip",
            "STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip",
            "STATE/LINEAGE/VECTORS/PROVENANCE.json",
        ):
            files = self.changed_document("SURFACE/surface.json", lambda d: d.update(
                resources=[r for r in d["resources"] if r["path"] != "../" + target]))
            with self.subTest(target=target), self.assertRaises(ValueError):
                verify.check_material(files)

    def test_corrupt_carriers_are_rejected(self):
        for name in self.files:
            if name.endswith(".zip"):
                files = dict(self.files)
                files[name] = files[name] + b"tamper"
                with self.subTest(name=name), self.assertRaisesRegex(ValueError, "binding mismatch"):
                    verify.check_material(files)

    def test_source_only_selection_cannot_omit_state(self):
        roots = {verify.HUMAN, verify.MACHINE, "STATE/LINEAGE/ORIGIN.json"}
        files = self.changed_document("SURFACE/surface.json", lambda d: (
            d.update(resources=[r for r in d["resources"] if r["path"][3:] in roots]),
            d["point"].update(provenance=["../STATE/LINEAGE/ORIGIN.json"]),
        ))
        with self.assertRaisesRegex(ValueError, "State core"):
            verify.check_material(files)

    def test_origin_policy_relation_must_be_selected(self):
        files = self.changed_document("STATE/LINEAGE/ORIGIN.json",
                                      lambda d: d.update(policy="../MISSING.json"))
        with self.assertRaisesRegex(ValueError, "Unselected material reference"):
            verify.check_material(files)

    def test_implementation_basis_and_attestation_relations_must_be_selected(self):
        path = "STATE/LINEAGE/IMPLEMENTATIONS/VM003-MEDIUM001.json"
        for field in ("current_basis_refs", "key_ref", "signature_ref"):
            def change(document):
                if field == "current_basis_refs":
                    document["present_status"][field] = ["missing.json"]
                else:
                    document["attestation"][field] = "missing.json"
            files = self.changed_document(path, change)
            with self.subTest(field=field), self.assertRaisesRegex(ValueError, "Unselected material reference"):
                verify.check_material(files)
        files = self.changed_document(path, lambda d: d["present_status"].update(
            current_basis_refs=["urn:external:unverified-basis"]))
        verify.check_material(files)

    def test_implementation_digest_mismatch_is_rejected(self):
        path = "STATE/LINEAGE/IMPLEMENTATIONS/VM003-MEDIUM001.json"
        files = self.changed_document(path, lambda d: d["identity"]["identity_refs"].__setitem__(
            0, "../../LOCALITY_VM_003-v3.0.0-GITHUB_UPLOAD.zip#sha256=" + "0" * 64 + "&byte_length=6201060"))
        with self.assertRaisesRegex(ValueError, "Implementation carrier binding mismatch"):
            verify.check_material(files)

    def test_unsafe_and_missing_relations_are_rejected(self):
        for reference in ("/etc/passwd", "../../../../outside.json", "missing.json"):
            files = self.changed_document("STATE/LINEAGE/LINEAGE.json",
                                          lambda d: d["related"].append(reference))
            with self.subTest(reference=reference), self.assertRaises(ValueError):
                verify.check_material(files)

    def test_absent_system_checker_cannot_be_replaced_by_state_code(self):
        files = dict(self.files)
        del files[verify.CHECKER]
        files["STATE/LINEAGE/CONFORMANCE.py"] = b"raise RuntimeError('must not execute')"
        with patch.object(verify, "revision_files", return_value=files):
            *_, report = verify.check_revision(verify.ROOT, self.revision)
            self.assertEqual(report["status"], "INCOMPLETE")
            self.assertEqual(report["conformance"]["status"], "UNAVAILABLE")
            with self.assertRaises(ValueError):
                verify.publication(verify.ROOT, self.revision)
            _, report = verify.publication(verify.ROOT, self.revision, allow_incomplete=True)
            self.assertEqual(report["status"], "INCOMPLETE")

    def test_vector_mismatch_cannot_be_bypassed(self):
        files = self.changed_document("STATE/LINEAGE/VECTORS/COMPOSITION.json", lambda d: d["cases"][0].update(
            expected={"schema_valid": True, "semantic_result": "INCONSISTENT", "failure_codes": ["WRONG_RESULT"]}))
        with patch.object(verify, "revision_files", return_value=files):
            with self.assertRaisesRegex(ValueError, "conformance did not pass"):
                verify.publication(verify.ROOT, self.revision, allow_incomplete=True)

    def test_roundtrip_is_exact_and_does_not_rerun_code(self):
        with tempfile.TemporaryDirectory() as directory:
            original = Path(directory) / "original"
            recovered = Path(directory) / "recovered"
            verify.write_fresh(original, self.edition)
            with patch.object(verify.subprocess, "run", side_effect=AssertionError("no execution")):
                files, report = verify.check_export(original, self.digest)
                verify.write_fresh(recovered, files)
                recovered_files, recovered_report = verify.check_export(recovered, self.digest)
            self.assertEqual(files, recovered_files)
            self.assertEqual(report, recovered_report)
            self.assertEqual(report["conformance"]["status"], "NOT_RERUN")
            self.assertEqual(report["conformance"]["recorded"]["status"], "PASSED")

    def test_export_tampering_and_wrong_trust_digest_fail(self):
        with tempfile.TemporaryDirectory() as directory:
            destination = Path(directory) / "edition"
            verify.write_fresh(destination, self.edition)
            with self.assertRaises(ValueError):
                verify.check_export(destination, "0" * 64)
            carrier = destination / "STATE/LOCALITY_MEDIUM_001-v2.0.0-GITHUB_UPLOAD.zip"
            carrier.write_bytes(carrier.read_bytes() + b"tamper")
            with self.assertRaisesRegex(ValueError, "Export binding mismatch"):
                verify.check_export(destination, self.digest)

    def test_self_consistent_manifest_cannot_hide_dangling_relation(self):
        files = copy.copy(self.edition)
        target = "STATE/LINEAGE/IMPLEMENTATIONS/VM003-MEDIUM001.md"
        del files[target]
        index = verify.load_json(files["index.json"])
        index["resources"] = [r for r in index["resources"] if r["path"] != target]
        index["references"] = {k: v for k, v in index["references"].items() if v["path"] != target}
        files["index.json"] = verify.encode(index)
        manifest = verify.load_json(files["manifest.json"])
        del manifest["files"][target]
        manifest["files"]["index.json"] = verify.binding(files["index.json"])
        files["manifest.json"] = verify.encode(manifest)
        with tempfile.TemporaryDirectory() as directory:
            destination = Path(directory) / "edition"
            verify.write_fresh(destination, files)
            with self.assertRaisesRegex(ValueError, "Unselected material reference"):
                verify.check_export(destination, verify.binding(files["manifest.json"])["sha256"])

    def test_surface_returns_exact_zip_bytes_and_keeps_check_ceiling(self):
        spec = importlib.util.spec_from_file_location("surface", verify.ROOT / "SURFACE/serve.py")
        surface = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(surface)
        index = verify.load_json(self.edition["index.json"])
        for resource in index["resources"]:
            if resource["media_type"] == "application/zip":
                request = {"resource": resource["ref"]}
                media, data = surface.read_resource(index, self.edition, request)
                self.assertEqual(media, "application/zip")
                self.assertEqual(data, self.files[resource["path"]])
                _, result = surface.check_resource(index, self.edition, request)
                result = verify.load_json(result)
                self.assertEqual(result["result"], "MATCH")
                self.assertEqual(result["conformance"], "NOT_EVALUATED")
                _, result = surface.resolve_reference(index, self.edition, {"ref": resource["ref"]})
                self.assertEqual(verify.load_json(result)["path"], resource["path"])


if __name__ == "__main__":
    unittest.main()

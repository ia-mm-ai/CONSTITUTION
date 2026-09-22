"""Tests for the core expression notation, not demonstrations of CSC or DCR."""

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

from jsonschema import Draft202012Validator, RefResolver


SPEC = importlib.util.spec_from_file_location(
    "core_conformance", Path(__file__).with_name("CONFORMANCE.py")
)
core = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(core)


class ExpressionTests(unittest.TestCase):
    def test_missing_differs_from_null_and_false(self):
        self.assertIs(core.lookup({}, "value"), core.MISSING)
        self.assertEqual(core.evaluate({"missing": ["x", "y"]}, {"y": None}), ["x"])
        self.assertFalse(core.evaluate({"===": [{"var": "x"}, None]}, {}))
        self.assertFalse(core.evaluate({"===": [{"var": "x"}, {"var": "y"}]}, {}))
        self.assertFalse(core.equal(False, 0))
        self.assertFalse(core.equal("1", 1))
        self.assertTrue(core.equal(1, 1.0))

    def test_deep_equality(self):
        self.assertTrue(core.equal({"b": [1, "x"], "a": False}, {"a": False, "b": [1, "x"]}))
        self.assertFalse(core.equal([1, 2], [2, 1]))
        self.assertFalse(core.equal({"a": [True]}, {"a": [1]}))

    def test_boolean_and_short_circuit(self):
        self.assertFalse(core.evaluate({"and": [False, {"invalid": "never evaluated"}]}, {}))
        self.assertTrue(core.evaluate({"or": [True, {"invalid": "never evaluated"}]}, {}))
        for value in (None, False, 0, "", [], core.MISSING):
            self.assertFalse(core.truthy(value))
        self.assertTrue(core.truthy({}))
        self.assertTrue(core.evaluate({"!": {"var": "x"}}, {}))

    def test_quantifier_scope_and_empty_array(self):
        expr = {"all": [{"var": "items"}, {"===": [{"var": "value"}, 2]}]}
        self.assertTrue(core.evaluate(expr, {"items": [{"value": 2}]}))
        self.assertFalse(core.evaluate(expr, {"value": 2, "items": [{}]}))
        self.assertTrue(core.evaluate({"all": [[], False]}, {}))
        self.assertFalse(core.evaluate({"some": [[], True]}, {}))
        self.assertTrue(core.evaluate({"none": [[], True]}, {}))
        self.assertTrue(core.evaluate({"some": [[1, 2], {"===": [{"var": ""}, 2]}]}, {}))

    def test_membership_and_count(self):
        self.assertTrue(core.evaluate({"in": [2, [1, 2, 3]]}, {}))
        self.assertFalse(core.evaluate({"in": [True, [1]]}, {}))
        self.assertEqual(core.evaluate({"count": {"var": "items"}}, {"items": []}), 0)
        with self.assertRaises(ValueError):
            core.evaluate({"in": ["a", "abc"]}, {})

    def test_ordering(self):
        for operation in (">", ">=", "<", "<="):
            self.assertIsInstance(core.evaluate({operation: [2, 3]}, {}), bool)
        self.assertTrue(core.evaluate({"<": ["2026-09-21T00:00:00Z", "2026-09-22T00:00:00Z"]}, {}))
        for values in ([True, 1], ["1", 1], [None, None]):
            with self.assertRaises(ValueError):
                core.evaluate({"<": values}, {})

    def test_bad_operations_never_pass(self):
        for expression in ({"exec": "bad"}, {"===": [1]}, {"count": None}, {"all": [None, True]}):
            with self.assertRaises(ValueError):
                core.evaluate(expression, {})

    def test_failure_order_and_no_universal_permission(self):
        rules = [
            {"assert": False, "failure_code": "B"},
            {"assert": False, "failure_code": "A"},
            {"assert": False, "failure_code": "B"},
            {"when": False, "assert": False, "failure_code": "C"},
        ]
        self.assertEqual(core.semantic_result(rules, {})["failure_codes"], ["B", "A"])
        self.assertEqual(core.semantic_result([], {})["semantic_result"], "CONSISTENT_AT_DECLARED_SCOPE")

    def test_strict_json_bytes(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "input.json"
            for data in (b'{"x":1,"x":2}', b'NaN', b'Infinity', b'"\xff"'):
                path.write_bytes(data)
                with self.assertRaises((ValueError, UnicodeError)):
                    core.load(path)
            path.write_bytes(b'{"value":null}')
            self.assertEqual(core.load(path), {"value": None})
            path.write_bytes(b'[1.0000000000000001,1,1e999,1.0]')
            values = core.load(path)
            self.assertFalse(core.equal(values[0], values[1]))
            self.assertTrue(core.equal(values[1], values[3]))
            self.assertTrue(core.ExactValidator({"type": "integer"}).is_valid(values[3]))
            self.assertTrue(core.ExactValidator({"type": "number"}).is_valid(values[2]))

    def test_remote_schema_reference_refused(self):
        with self.assertRaises(ValueError):
            core.no_remote("https://example.invalid/schema")
        resolver = core.OfflineResolver.from_schema({})
        with self.assertRaises(ValueError):
            resolver.resolve_remote("ftp://example.invalid/schema")

    def test_exact_integer_through_schema_reference(self):
        target = {
            "$schema": "https://json-schema.org/draft/2020-12/schema",
            "$id": "urn:vector:integer",
            "type": "integer",
        }
        schema = {"$ref": "urn:vector:integer"}
        resolver = core.OfflineResolver.from_schema(schema, store={target["$id"]: target})
        validator = core.ExactValidator(schema, resolver=resolver)
        self.assertTrue(validator.is_valid(core.Decimal("1.0")))
        self.assertFalse(validator.is_valid(core.Decimal("1.0000000000000001")))


class CompositionTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        directory = Path(__file__).parent / "SCHEMAS"
        cls.schemas = {}
        for stem in ("CORE", "RULE", "VECTOR", "INTERFACE", "ORIGIN", "DERIVATION", "MODULE", "STATE"):
            schema = core.load(directory / f"{stem}.schema.json")
            Draft202012Validator.check_schema(schema)
            cls.schemas[schema["$id"]] = schema

    def validator(self, stem):
        schema = self.schemas[core.BASE + stem + ":1"]
        return Draft202012Validator(
            schema,
            resolver=RefResolver.from_schema(schema, store=self.schemas, handlers={"urn": core.no_remote}),
        )

    def test_rules_and_vectors(self):
        model = core.load(core.ROOT / "STATE/STATE.json")
        self.validator("state").validate(model)
        suite = core.load(Path(__file__).parent / "VECTORS/COMPOSITION.json")
        self.validator("vector").validate(suite)
        for rule in model["rules"]:
            self.validator("rule").validate(rule)
        for case in suite["cases"]:
            valid = self.validator("interface").is_valid(case["input"])
            self.assertEqual(valid, case["expected"]["schema_valid"], case["id"])
            if valid:
                self.assertEqual(core.semantic_result(model["rules"], case["input"]), case["expected"])

    def test_locator_and_exact_source_bytes(self):
        locator = core.load(Path(__file__).parent / "ORIGIN.json")
        self.validator("origin").validate(locator)
        for key in ("human_content_binding", "machine_content_binding"):
            binding = locator[key]
            data = (Path(__file__).parent / binding["path"]).read_bytes()
            self.assertEqual(len(data), binding["byte_length"])
            self.assertEqual(core.hashlib.sha256(data).hexdigest(), binding["sha256"])

    def test_unknown_and_missing_core_properties_rejected(self):
        suite = core.load(Path(__file__).parent / "VECTORS/COMPOSITION.json")
        document = suite["cases"][0]["input"]
        for field in document:
            changed = {key: value for key, value in document.items() if key != field}
            self.assertFalse(self.validator("interface").is_valid(changed), field)
        changed = dict(document, permission=True)
        self.assertFalse(self.validator("interface").is_valid(changed))
        changed = json.loads(json.dumps(document))
        changed["input"]["authority"] = "inherited"
        self.assertFalse(self.validator("interface").is_valid(changed))


if __name__ == "__main__":
    unittest.main()

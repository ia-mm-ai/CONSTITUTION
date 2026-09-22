"""Regression tests for the derivative evaluator, not demonstrations of CSC or DCR."""

import json
from pathlib import Path
import tempfile
import unittest

import conformance as core


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

    def test_quantified_scope_is_bound_to_original_input(self):
        expression = {"all": [
            {"var": "relations"},
            {"===": [{"var": "scope"}, {"root": "scope"}]},
        ]}
        self.assertTrue(core.evaluate(expression, {"scope": "motion", "relations": [{"scope": "motion"}]}))
        self.assertFalse(core.evaluate(expression, {"scope": "motion", "relations": [{"scope": "other"}]}))
        nested = {"all": [
            {"var": "groups"},
            {"all": [{"var": ""}, {"===": [{"var": ""}, {"root": "scope"}]}]},
        ]}
        self.assertTrue(core.evaluate(nested, {"scope": "motion", "groups": [["motion"]]}))
        self.assertFalse(core.evaluate(nested, {"scope": "motion", "groups": [["other"]]}))
        self.assertIs(core.evaluate({"root": "absent"}, {}), core.MISSING)

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


class CoreTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.core = core.Core()

    def test_all_schemas_vectors_and_records(self):
        for identifier in self.core.schemas:
            self.core.check_schema(identifier)
        for module in core.MODULES:
            self.assertGreater(self.core.run_module(module), 0)
        self.core.check_records()

    def test_historical_account_cannot_claim_present_demonstration(self):
        claim = core.load(core.ROOT / "STATE/LINEAGE/IMPLEMENTATIONS/VM003-MEDIUM001.json")
        model = core.load(core.ROOT / core.MODULES["LINEAGE"])
        self.assertEqual(core.semantic_result(model["rules"], claim)["failure_codes"], [])
        claim["claimed_effects"].append("DECLARED_DEMONSTRATION")
        self.assertIn("IMPLEMENTATION_HISTORICAL_TO_CURRENT_CONVERSION",
                      core.semantic_result(model["rules"], claim)["failure_codes"])

    def test_unknown_and_missing_core_properties_rejected(self):
        suite = core.load(core.ROOT / "STATE/LINEAGE/VECTORS/COMPOSITION.json")
        document = suite["cases"][0]["input"]
        validator = self.core.validator(core.BASE + "interface:1")
        for field in document:
            changed = {key: value for key, value in document.items() if key != field}
            self.assertFalse(validator.is_valid(changed), field)
        self.assertFalse(validator.is_valid(dict(document, permission=True)))
        changed = json.loads(json.dumps(document))
        changed["input"]["authority"] = "inherited"
        self.assertFalse(validator.is_valid(changed))


if __name__ == "__main__":
    unittest.main()

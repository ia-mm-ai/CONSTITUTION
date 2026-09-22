"""Offline core-notation conformance; never an implementation or truth oracle."""

import argparse
from decimal import Decimal
import hashlib
import json
from pathlib import Path
import sys

from jsonschema import Draft202012Validator, FormatChecker, RefResolver, validators


ROOT = Path(__file__).resolve().parents[2]
BASE = "urn:presence:core:lineage:"
MISSING = object()
NUMBER_TYPES = (int, float, Decimal)
MODULES = {
    "STATE": "STATE/STATE.json",
    "CSC": "STATE/CONTINUITY-STATE-CAPABILITY/CONTINUITY-STATE-CAPABILITY.json",
    "DCR": "STATE/DYNAMIC-CAPACITY-REGULATION/DYNAMIC-CAPACITY-REGULATION.json",
    "LINEAGE": "STATE/LINEAGE/LINEAGE.json",
}


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"Duplicate JSON key: {key}")
        result[key] = value
    return result


def exact_decimal(text):
    value = Decimal(text)
    if not value.is_finite():
        raise ValueError("Non-finite JSON number")
    return value


def invalid_constant(text):
    raise ValueError(f"Non-JSON number: {text}")


def load(path):
    return json.loads(
        Path(path).read_bytes().decode("utf-8"),
        object_pairs_hook=unique_object,
        parse_float=exact_decimal,
        parse_constant=invalid_constant,
    )


def nodes(value):
    if isinstance(value, dict):
        yield value
        for child in value.values():
            yield from nodes(child)
    elif isinstance(value, list):
        for child in value:
            yield from nodes(child)


def truthy(value):
    if value is MISSING or value is None:
        return False
    if isinstance(value, (bool, int, float, Decimal, str, list)):
        return bool(value)
    return True


def equal(left, right):
    if left is MISSING or right is MISSING:
        return False
    if type(left) in NUMBER_TYPES and type(right) in NUMBER_TYPES:
        return left == right
    if type(left) is not type(right):
        return False
    if isinstance(left, dict):
        return left.keys() == right.keys() and all(
            equal(left[key], right[key]) for key in left
        )
    if isinstance(left, list):
        return len(left) == len(right) and all(
            equal(a, b) for a, b in zip(left, right)
        )
    return left == right


def lookup(document, path):
    if not isinstance(path, str):
        raise ValueError("var needs a string path")
    if path == "":
        return document
    for component in path.split("."):
        if not isinstance(document, dict) or component not in document:
            return MISSING
        document = document[component]
    return document


def array(value):
    if not isinstance(value, list):
        raise ValueError("Expected an array operand")
    return value


def evaluate(expression, document):
    if isinstance(expression, list):
        return [evaluate(item, document) for item in expression]
    if not isinstance(expression, dict):
        return expression
    if len(expression) != 1:
        raise ValueError("An operation must have exactly one key")
    operator, args = next(iter(expression.items()))
    if operator == "var":
        return lookup(document, args)
    if operator == "missing":
        return [path for path in array(args) if lookup(document, path) is MISSING]
    if operator == "!":
        return not truthy(evaluate(args, document))
    if operator == "count":
        return len(array(evaluate(args, document)))
    if operator == "and":
        return all(truthy(evaluate(arg, document)) for arg in array(args))
    if operator == "or":
        return any(truthy(evaluate(arg, document)) for arg in array(args))
    if operator not in ("===", "!==", "in", ">", ">=", "<", "<=", "all", "some", "none"):
        raise ValueError(f"Unknown operation: {operator}")
    if len(array(args)) != 2:
        raise ValueError(f"{operator} needs two operands")
    left = evaluate(args[0], document)
    if operator in ("all", "some", "none"):
        values = (truthy(evaluate(args[1], item)) for item in array(left))
        return all(values) if operator == "all" else (
            any(values) if operator == "some" else not any(values)
        )
    right = evaluate(args[1], document)
    if operator == "===":
        return equal(left, right)
    if operator == "!==":
        return not equal(left, right)
    if operator == "in":
        return any(equal(left, item) for item in array(right))
    numbers = type(left) in NUMBER_TYPES and type(right) in NUMBER_TYPES
    strings = isinstance(left, str) and isinstance(right, str)
    if not (numbers or strings):
        raise ValueError("Ordering requires two numbers or two strings")
    return {
        ">": lambda: left > right,
        ">=": lambda: left >= right,
        "<": lambda: left < right,
        "<=": lambda: left <= right,
    }[operator]()


def semantic_result(rules, document):
    failures = []
    for rule in rules:
        if truthy(evaluate(rule.get("when", True), document)):
            if not truthy(evaluate(rule["assert"], document)):
                code = rule["failure_code"]
                if code not in failures:
                    failures.append(code)
    return {
        "schema_valid": True,
        "semantic_result": "INCONSISTENT" if failures else "CONSISTENT_AT_DECLARED_SCOPE",
        "failure_codes": failures,
    }


def no_remote(uri):
    raise ValueError(f"Non-local schema reference is forbidden: {uri}")


class OfflineResolver(RefResolver):
    def resolve_remote(self, uri):
        return no_remote(uri)


def is_integer(checker, instance):
    if isinstance(instance, Decimal):
        return instance.is_finite() and instance == instance.to_integral_value()
    return Draft202012Validator.TYPE_CHECKER.is_type(instance, "integer")


ExactValidator = validators.extend(
    Draft202012Validator,
    type_checker=Draft202012Validator.TYPE_CHECKER.redefine("integer", is_integer),
    # Keep exact-number semantics when a referenced schema declares its dialect.
    version="presence-exact-draft2020-12",
)


class Core:
    def __init__(self, root=ROOT):
        self.root = Path(root).resolve()
        self.schemas = {}
        self.schema_paths = {}
        for path in sorted((self.root / "STATE").glob("*/SCHEMAS/*.schema.json")):
            schema = load(path)
            identifier = schema.get("$id")
            if not identifier:
                raise ValueError(f"Missing schema ID: {path}")
            if identifier in self.schemas:
                raise ValueError(f"Duplicate schema ID: {identifier}")
            self.schemas[identifier] = schema
            self.schema_paths[identifier] = path
        machine = load(self.root / "SOURCE/CONSTITUTION_0()1.json")
        self.refs = {node["ref"] for node in nodes(machine) if "ref" in node}
        human = (self.root / "SOURCE/CONSTITUTION_0()1.md").read_bytes()
        binding = machine["binding"]
        if binding["human_byte_length"] != len(human) or (
            binding["human_sha256"] != hashlib.sha256(human).hexdigest()
        ):
            raise ValueError("SOURCE_BINDING_MISMATCH")
        origin = load(self.root / "STATE/LINEAGE/ORIGIN.json")
        for key in ("human_content_binding", "machine_content_binding"):
            binding = origin[key]
            path = self.local(self.root / "STATE/LINEAGE", binding["path"])
            data = path.read_bytes()
            if len(data) != binding["byte_length"] or (
                hashlib.sha256(data).hexdigest() != binding["sha256"]
            ):
                raise ValueError(f"ORIGIN_BINDING_MISMATCH: {key}")

    def local(self, directory, relative):
        path = (directory / relative).resolve()
        if not path.is_relative_to(self.root):
            raise ValueError(f"Path escapes core: {relative}")
        return path

    def source_refs(self, document, skip_inputs=False):
        if isinstance(document, dict):
            for key, value in document.items():
                if key in ("source", "x-source") and isinstance(value, dict) and "refs" in value:
                    for ref in value["refs"]:
                        if ref not in self.refs:
                            raise ValueError(f"Unresolved constitutional reference: {ref}")
                if not (skip_inputs and key == "input"):
                    self.source_refs(value, skip_inputs)
        elif isinstance(document, list):
            for value in document:
                self.source_refs(value, skip_inputs)

    def validator(self, identifier):
        if identifier not in self.schemas:
            raise ValueError(f"Unresolved core schema: {identifier}")
        schema = self.schemas[identifier]
        resolver = OfflineResolver.from_schema(schema, store=self.schemas)
        return ExactValidator(schema, resolver=resolver, format_checker=FormatChecker())

    def validate(self, identifier, document):
        self.validator(identifier).validate(document)
        self.source_refs(document)

    def check_schema(self, identifier):
        schema = self.schemas[identifier]
        if schema.get("$schema") != "https://json-schema.org/draft/2020-12/schema":
            raise ValueError(f"Not Draft 2020-12: {identifier}")
        if schema.get("x-status") != "CANONICAL_SCHEMA":
            raise ValueError(f"Not canonical schema: {identifier}")
        Draft202012Validator.check_schema(schema)
        self.source_refs(schema)
        for node in nodes(schema):
            if "$ref" in node:
                target = node["$ref"].split("#")[0]
                if target and target not in self.schemas:
                    raise ValueError(f"Non-local or missing reference: {node['$ref']}")

    def run_module(self, module):
        model_path = self.root / MODULES[module]
        model = load(model_path)
        self.validate(BASE + ("state:1" if module == "STATE" else "module:1"), model)
        rule_ids = [rule["id"] for rule in model["rules"]]
        if len(rule_ids) != len(set(rule_ids)):
            raise ValueError(f"Duplicate rule IDs in {module}")
        for rule in model["rules"]:
            self.validate(BASE + "rule:1", rule)
        covered = set()
        declared_schemas = set()
        for relative in model["schemas"]:
            path = self.local(model_path.parent, relative)
            identifier = load(path)["$id"]
            self.check_schema(identifier)
            declared_schemas.add(identifier)
        count = 0
        identifiers = set()
        for relative in model["vectors"]:
            path = self.local(model_path.parent, relative)
            suite = load(path)
            self.validator(BASE + "vector:1").validate(suite)
            self.source_refs(suite, skip_inputs=True)
            for case in suite["cases"]:
                if case["id"] in identifiers:
                    raise ValueError(f"Duplicate vector ID: {case['id']}")
                identifiers.add(case["id"])
                if case["schema"] not in declared_schemas:
                    raise ValueError(f"Vector schema not declared by {module}: {case['schema']}")
                self.check_schema(case["schema"])
                covered.add(case["schema"])
                valid = self.validator(case["schema"]).is_valid(case["input"])
                if valid:
                    try:
                        self.source_refs(case["input"])
                    except ValueError:
                        valid = False
                actual = semantic_result(model["rules"], case["input"]) if valid else {
                    "schema_valid": False,
                    "semantic_result": "NOT_EVALUATED",
                    "failure_codes": ["SCHEMA_INVALID"],
                }
                if not equal(actual, case["expected"]):
                    raise ValueError(
                        f"{path.name}:{case['id']}: expected {case['expected']}, got {actual}"
                    )
                count += 1
        if not covered or not count:
            raise ValueError(f"No evaluated vectors in {module}")
        print(f"{module}: {count} vectors passed; {len(covered)} schemas exercised")
        return count

    def check_records(self):
        self.validate(BASE + "origin:1", load(self.root / "STATE/LINEAGE/ORIGIN.json"))
        records = sorted((self.root / "STATE/LINEAGE/RECORDS").glob("*.json"))
        if not {"VM003-TO-CSC.json", "MEDIUM-TO-DCR.json"}.issubset({path.name for path in records}):
            raise ValueError("The two extraction records must be present")
        for path in records:
            record = load(path)
            self.validate(BASE + "derivation:1", record)
            inventory = [item["path"] for item in record["inspected_material"]]
            if len(inventory) != len(set(inventory)):
                raise ValueError(f"Duplicate inspected path: {path}")
            target_rules = set()
            for relative in record["target_files"]:
                target = self.local(self.root, relative)
                if not target.is_file():
                    raise ValueError(f"Missing derivation target: {relative}")
                if target.suffix == ".json" and target.name in (
                    "CONTINUITY-STATE-CAPABILITY.json", "DYNAMIC-CAPACITY-REGULATION.json"
                ):
                    target_rules.update(rule["id"] for rule in load(target)["rules"])
            for mapping in record["extracted_rule_map"]:
                if not set(mapping["input_paths"]).issubset(inventory):
                    raise ValueError(f"Extraction map has uninspected input: {path}")
                if not set(mapping["target_rules"]).issubset(target_rules):
                    raise ValueError(f"Extraction map has unknown target rule: {path}")
        print("Origin and derivation records passed (claims remain unverified)")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--module", choices=MODULES, action="append")
    parser.add_argument("--records", action="store_true")
    args = parser.parse_args()
    core = Core()
    if args.records:
        core.check_records()
    else:
        if not args.module:
            for identifier in core.schemas:
                core.check_schema(identifier)
        for module in args.module or MODULES:
            core.run_module(module)
        if not args.module:
            core.check_records()
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (ValueError, KeyError, OSError) as error:
        print(f"CONFORMANCE ERROR: {error}", file=sys.stderr)
        sys.exit(1)

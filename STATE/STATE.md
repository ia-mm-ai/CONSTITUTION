# STATE

## Standing and authority

This is the canonical State core: a normative representation contract, not a
Body's current State, a new Constitution, or an implementation. The
[human Constitution](../SOURCE/CONSTITUTION_0%28%291.md) governs meaning. The
[bound machine Constitution](../SOURCE/CONSTITUTION_0%28%291.json) supplies
references, not independent law. Verify its internal exact-human-byte binding
before using those references. A mismatch invalidates the binding, not the
human Constitution. An irreconcilable Form conflict remains `UNRESOLVED`; the
dependent consequential motion is held or refused at its exact scope, never
given the broader effect (Readings 2–7).

“MUST” below concerns this representation contract. A constitutional obligation
is identified by its exact `source.refs`; an interoperable representation
choice is identified as `CORE_ENCODING_RULE`. Neither silently amends Source.
`CANONICAL_CORE`, `CANONICAL_STANDARD`, `CANONICAL_SCHEMA`, and
`NORMATIVE_VECTOR` are repository metadata, not constitutional vocabulary.

## Composition without a universal pipeline

| Branch | Job | What it cannot supply |
| --- | --- | --- |
| [CSC](CONTINUITY-STATE-CAPABILITY/CONTINUITY-STATE-CAPABILITY.md) | Requirements for preserving supported State succession through transition, correction, dormancy, and resumption | Missing present capacity, identity, Agency, Authority, or a successor's independent basis |
| [DCR](DYNAMIC-CAPACITY-REGULATION/DYNAMIC-CAPACITY-REGULATION.md) | What an exact Locality presently supports, distinguishing structural and situated capacity | State succession, truth, general permission, or Authority |
| [LINEAGE](LINEAGE/LINEAGE.md) | Exact attributable occurrences and relations, including derivation and implementation claims | The occurrence itself, reciprocity, inheritance, adoption, or present capability |

[STATE.json](STATE.json) specifies these interfaces. There is no compulsory
CSC→DCR→LINEAGE order, shared runtime, universal gate, ledger, or global State.
A result becomes another module's input only through an explicit attributable
relation identifying the producing record, consuming record, exact target,
scope, Locality, effective boundary, conditions, supporting basis and evidence,
and effect ceiling. Use [INTERFACE](LINEAGE/SCHEMAS/INTERFACE.schema.json) for
this narrow cross-module input relation; LINEAGE's general relations remain
distinct. A transfer of a representation is not transfer of the represented
function or its effect (sections 5.1–5.2, 11.10–11.12, 19.2–19.4).

A DCR disposition may be referred to in a CSC proposal, but cannot establish
the proposal's predecessor, preservation, Agency, or Authority. A CSC record
may be referred to by DCR, but does not refresh its time-sensitive observations.
A LINEAGE record makes either reference explicit; it does not certify either
claim. No module fills another module's missing basis. Unconnected records
remain unconnected; matching identifiers are not an implicit relation.

## Common representation discipline

The following are **CORE_ENCODING_RULES**, limited to interoperable notation.

1. Core documents use JSON and Draft 2020-12 schemas. Objects are closed unless
   a schema explicitly defines an extension boundary. Unknown fields and
   missing required fields are syntax failures, not constitutional negatives.
   Each schema has a stable `urn:presence:core:…:1` identifier; resolve it from
   the exact checked-out core revision, never by fetching arbitrary references.
2. `source.refs` and `x-source.refs` name exact stable `ref` entries in the
   bound machine Constitution. A constitutional rule MUST cite at least one.
   Structural choices, enumerations, notation, and test execution semantics are
   encoding rules even when their purpose is supported by a cited clause.
3. Omission means no statement was supplied; a required omission is malformed.
   `null` is not a placeholder or an alternative to `UNKNOWN`: it is rejected
   unless a particular schema explicitly assigns it a bounded meaning.
   `UNKNOWN` means unestablished; `UNRESOLVED` preserves an identified conflict.
   `UNSUPPORTED` means the stated support does not establish this scoped claim;
   `REFUSED` is an attributable scoped disposition. None means general absence,
   falsehood, incapacity, consent, or permanent exclusion (Reading 5;
   sections 3.4–3.5, 11.12, 13.2).
4. Identifiers are explicit references, not content hashes or claims of global
   uniqueness. SHA-256 bindings always cover **the exact stored file bytes**,
   with byte length, no newline conversion, Unicode normalization, JSON
   reserialization, archive decompression, or implicit canonicalization. To
   bind a newly serialized document, first fix its UTF-8 file bytes, then hash
   those bytes; different serializations are different byte objects. Internal
   record IDs and semantic equality never depend on serialized byte hashes.
   Duplicate JSON keys, non-finite numbers and invalid UTF-8 are rejected.
5. Repository file references resolve relative to their containing document,
   except `target_files` in derivations, which are repository-root-relative.
   Archive inventory paths use `!/` to cross a nested archive boundary.
   Digests of inventory members cover uncompressed member bytes; outer archive
   digests cover the original archive bytes. Evidence remains outside the
   canonical tree; a digest is a locator, not proof of a claim.

## Formal rule and vector evaluation

The module models contain independent rules, not a state-changing program.
The [rule schema](LINEAGE/SCHEMAS/RULE.schema.json) defines a small closed
expression language. An expression is a JSON literal, an array of expressions,
or a single-key operation. Objects as data are obtained through `var` or `root`, not
executed as arbitrary code. There are no network, clock, filesystem, random,
evaluation, or external-execution operators.

**CORE_ENCODING_RULE — expression semantics:**

- `var` takes a dotted property path; `""` denotes the current evaluation
  object. Missing paths produce a distinct internal missing value, not `null`.
  `missing` takes an array of paths and returns the missing paths in order.
- `root` takes the same path notation but always reads the immutable original
  input, including inside nested quantifiers. This explicit encoding operation
  binds a quantified item's scope to its enclosing motion; an item's own
  internally consistent scope is not sufficient. It supplies no ambient
  runtime state or external evidence.
- `===` and `!==` compare two evaluated values by deep JSON equality: no
  coercion of strings, booleans, numbers, arrays, or objects; object key order
  is immaterial, array order is material. A missing operand never establishes
  equality, including with another missing operand. Finite JSON numbers compare
  as exact base-ten values, without binary floating-point rounding; booleans
  are not numbers. Numeric `1.0` equals `1` and satisfies the integer type.
- `and` / `or` take arrays and return a Boolean; `!` negates one expression.
  False, null, missing, zero, empty string and empty array are falsey; all
  other values are truthy. Logical operations short-circuit.
- `in` takes `[value, array]` and uses deep equality, not substring matching.
  `count` takes an array expression and returns its length.
- `>`, `>=`, `<`, `<=` take two finite numbers or two strings of the same
  type; string comparison is Unicode code-point order. Any time comparison
  therefore requires a schema-constrained uniform time representation.
- `all`, `some`, `none` take `[array, predicate]`; the predicate's `var` paths
  are relative to each item, not an ambient parent. `all([])` is true,
  `some([])` false, and `none([])` true; required evidence must separately
  require nonempty arrays. An operand of the wrong type is an evaluation
  error, never a passing rule.

For each vector, validate `input` against the exact named schema with format
checking enabled. Invalid syntax yields `schema_valid: false`,
`semantic_result: NOT_EVALUATED`, `failure_codes: ["SCHEMA_INVALID"]`.
For valid syntax, evaluate every applicable rule (`when`, default true) in
model array order; a false `assert` appends its failure code, once, in that
order. No failures yields `CONSISTENT_AT_DECLARED_SCOPE`; otherwise the result
is `INCONSISTENT`. These labels describe represented claims only. Unknown
conditions may be consistently represented without supporting action.
Errors in a model, unresolved references or evaluation errors stop the check;
they are not converted to ordinary refusal or success.

The suites contain exact inputs and expected triples, source references and
forbidden inferences. They are normative tests, not actual events, simulated
Localities, present capability demonstrations, or evidence of adoption.
[CONFORMANCE.py](LINEAGE/CONFORMANCE.py) is an offline notation/vector checker,
not a CSC or DCR implementation. It uses Python's standard library and the
existing `jsonschema` validator (tested with 4.10.3); it performs no local
dispositions, deployments or State mutation. Run from any directory:

```sh
python /home/runner/work/PRESENCE/PRESENCE/STATE/LINEAGE/CONFORMANCE.py
python -m unittest discover -s /home/runner/work/PRESENCE/PRESENCE/STATE/LINEAGE -p 'test_conformance.py'
```

## Qualification and proof boundaries

Source, evidence, records, representations and supported State remain
distinct. Correction preserves predecessor truth and actual consequence;
validity never erases an occurrence (sections 14.1–14.5, 15.1–15.5).
An implementation claim must identify the exact core commit and Source byte
binding, scope, conditions, dependencies, demonstrations, failure boundaries,
present status and proof ceiling. Schema acceptance or historical qualification
does not establish present CSC or capacity (sections 16.7–16.8, 18.1–18.4).

[ORIGIN](LINEAGE/ORIGIN.json) locates this canonical core and the verified
Source snapshot. The derivation records preserve the two predecessor packs'
inspected materials, extracted rules, exclusions and refused claims; they
are semantic provenance, not duplicate Git history or succession of a Body.
The archives and their executable contents are not part of this core.

This crossing neither amends Source nor forms, adopts, deploys or demonstrates
a Body or implementation. No bearer, platform, repository or verifier becomes
parent, owner, sovereign or keeper by publication, ancestry, technical control
or necessity (sections 1.2–1.5, 6.4–6.9, 19.1–19.4, 20.1–20.6; Reading 7).

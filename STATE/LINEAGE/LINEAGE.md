# LINEAGE

## 1. Standing, authority, and scope

**Canonical, platform-neutral Core standard.** This document specifies a
normative representation and bounded consistency profile, not a new Constitution,
an adoption instrument, or an assertion that a repository or bearer presently
demonstrates CSC. Read with [STATE](../STATE.md), the authoritative
[human Constitution](../../SOURCE/CONSTITUTION_0%28%291.md), and its bound
[machine references](../../SOURCE/CONSTITUTION_0%28%291.json).

The human Form supplies substantive meaning. Stable references identify that
meaning; neither references nor this encoding may widen, replace, or silently
narrow it. An irreconcilable Form conflict remains `UNRESOLVED`; dependent
consequential motion must be held or refused at its exact scope, not assigned the
broader effect (Readings 2–7). A local adoption basis is still necessary. Shared
code, ancestry, custody, or this standard's publication creates no external
jurisdiction or parental sovereignty (§§ 1.2–1.5, 17.5, 20.3).

**MUST**, **MUST NOT**, and **SHALL** specify obligations of a declaration in this
Core profile. Substantive obligations are sourced below and in each machine
rule. Requirements concerning JSON syntax, field presence, ordering, identifiers,
hash representation, and closed objects are **CORE_ENCODING_RULE** choices,
not added constitutional laws. Failure to fit this profile does not establish
that an external reality is false, illegitimate, ineligible, or absent.

The standard covers exact occurrences, attributable relations, target-specific
succession, incompatible streams, re-entry, and implementation claims. It has no
dependency on a particular host, operating system, ledger, runtime, virtual
machine, network, signature system, or database. Human procedures can embody
bounded functions as well as software; no technology is privileged (§ 19.3).

## 2. Contract and evaluation

[LINEAGE.json](LINEAGE.json) is the executable normative rule model.
[The schemas](#3-record-contracts) use JSON Schema Draft 2020-12. Schema IDs
are `urn:presence:core:lineage:<lowercase-file-stem>:1`. Schemas are marked
`CANONICAL_SCHEMA`; the model is `CANONICAL_STANDARD`; executable suites are
`NORMATIVE_VECTOR`. These statuses describe artifacts, not constitutional force.

An evaluator SHALL:

1. Select the exact schema by its URN, without guessing from filename or accepting
   an unregistered schema. Validate the entire input, with no coercion, defaults,
   property dropping, or replacement of an unsupported value.
2. On schema failure, return `schema_valid: false`,
   `semantic_result: NOT_EVALUATED`, `failure_codes: ["SCHEMA_INVALID"]`.
   Do not run semantic rules on malformed input.
3. On schema success, evaluate applicable rules in array order. `when` omitted
   means true. Collect every false assertion's `failure_code`, deduplicating
   while preserving first occurrence. There is no first-error shortcut.
4. Return `INCONSISTENT` with those codes, or
   `CONSISTENT_AT_DECLARED_SCOPE` with an empty list. Both carry
   `schema_valid: true`.

Expressions use only `var`, strict deep `===` / `!==`, `and`, `or`, `!`, array
membership `in`, ordering `> >= < <=`, `all`, `some`, `none`, `missing`, and
`count`. A variable's dotted path starts at input; inside a quantifier it starts
at that item, and `var: ""` means that item. There is no implicit access to the
outer record from a quantifier, code execution, reference traversal, network
fetch, or custom operation. Objects in expressions have exactly one operator.
Equality never converts strings, numbers, and booleans. Arrays compare in order.
`all` and `none` on an empty array are true; `some` is false. Rules explicitly
check nonemptiness where affirmative support requires evidence.

The result checks **declared internal consistency only**. It does not authenticate
evidence, validate an independent constitutional basis, resolve the interior of a
Locality, establish adoption, verify a signature, retrieve remote bytes, grant
Authority, or prove present CSC/DCR. A declaration can be internally consistent
and factually unsupported. Conversely, inconsistent authorization claims do not
make their real consequences unreal (§§ 5.2, 14.2, 18.1–18.4, 19.3).

### Representation discipline

- Every record and nested object is closed; all documented fields are required,
  except fields belonging only to the signed attestation alternative. Unknown
  keys and unknown enum members are rejected, not interpreted as extensions.
- Null is disallowed throughout these six record contracts. Missing required
  structure is malformed; it does not encode an uncertain proposition.
  `UNKNOWN` means unestablished; `UNRESOLVED` means an open conflict or material
  unresolved question; `UNSUPPORTED` means support has not been supplied for the
  claim; `REFUSED` records an exact refusal. None means false, absent, permanent
  exclusion, or permission. No affirmative absence finding is encoded here.
- Nonempty strings identify exact qualified referents, supported temporal
  descriptions, or bounded descriptions. A temporal reference can be an exact
  occurrence or a supported interval; fabricated clock precision is forbidden.
  `UNKNOWN` / `UNRESOLVED` may be used as explicit unresolved references in an
  incomplete declaration, never as evidence establishing an effect.
- Evidence and history arrays contain exact references. An empty array states
  that no entries are declared in that category at this scope, not that none
  exist in reality. Unestablished material mapping MUST be explicitly marked
  unresolved rather than disguised as a complete empty map.
- Ordered predecessor snapshot arrays are copied exactly into their retained
  counterparts. This is a deterministic preservation check, not a demand that
  reality have a total order. Later additions and corrections receive new exact
  references; they do not replace the predecessor snapshot.
- References name objects but do not import their truth, status, permission,
  applicability, or effect. Validation does not follow an arbitrary graph to
  create constitutional conversions (§§ 5.1–5.2, 19.2).

## 3. Record contracts

| Schema | Record kind | Required distinction |
| --- | --- | --- |
| [OCCURRENCE](SCHEMAS/OCCURRENCE.schema.json) | `occurrence` | Exact event, evidence, record, and carrier |
| [RELATION](SCHEMAS/RELATION.schema.json) | `relation` | Distinct terms and each attributed side's basis |
| [SUCCESSION](SCHEMAS/SUCCESSION.schema.json) | `succession` | Independent basis, exact target, establishment, mapping, boundary, remainder |
| [DIVERGENCE](SCHEMAS/DIVERGENCE.schema.json) | `divergence` | Incompatible streams, preserved history, bounded treatment |
| [RE-ENTRY](SCHEMAS/RE-ENTRY.schema.json) | `re_entry` | Historical relation versus new current binding |
| [IMPLEMENTATION-CLAIM](SCHEMAS/IMPLEMENTATION-CLAIM.schema.json) | `implementation_claim` | Exact implementation and Core binding versus demonstrated scope and effect ceiling |

The separately specified [DERIVATION contract](SCHEMAS/DERIVATION.schema.json)
and [ORIGIN locator](ORIGIN.json) are linked, not redefined here. Reading their
contents requires their own contracts. Derivation does not by itself decide
adaptation, amendment, transformation, succession, or new formation (§ 17.8).
An origin locator is not a genesis event, worldwide priority claim, identity
certificate, Source transfer, or grant of continuing availability.

### 3.1 Exact occurrence

An occurrence SHALL identify its holding Locality, attribution, exact target,
kind, time, scope, and support posture. Exposure, emission, crossing, delivery,
receipt, inspection, admission, uptake, integration, and response remain
different occurrences (§§ 3.2, 7.2–7.3).

The occurrence, carrier, preserved record, and evidence have different qualified
referents even if a single file physically carries representations of them.
`evidence.object`, `occurrence_type`, and `scope` MUST match the event claim;
the proof declaration names method, time, basis, result, and effect ceiling.
Affirmative support MUST reference that evidence in the attribution account.
Evidence time is the time of evidence, not automatically the event time.

The only effect of this record is its `EXACT_OCCURRENCE` claim. For a directly
supported receipt, declare an occurrence of type `RECEIPT`; do not add `RECEIPT`
as an inferred effect of delivery. Other occurrences require their own support.
Self-report supports an attributable claim, signature a bounded key/byte/
attribution relation, replay a reproducible trace, inspection an external
correlate, and conformance only tested conditions (§§ 18.1–18.4).

### 3.2 Attributable relation

A relation identifies exact left and right terms, holding Locality, kind,
direction, scope, each side's basis, duration, asymmetry, refusal, withdrawal,
correction, suspension, exit, history, and uncertainty (§§ 9.2–9.5).

This is a **bilateral encoding profile**, not a claim that constitutional
relations can only have two terms. A multilateral account needs explicit support
for every relevant side; a collection of binary records does not itself
establish a shared multilateral relation. An attributable set of pairwise records
may represent a multi-side relation when its exact shared basis covers every
relevant side; graph connectivity or transitive reference is never that basis.

The left side holds the local record. Each basis identifies the correct bearer,
counterpart, Locality, scope, origin, evidence, and posture. A local claim needs
the left side's valid basis; `RECIPROCAL` or `SHARED` additionally requires the
right side's basis. `HOST_INFERENCE` and `REFERENCE_CHAIN` do not supply it.
Unknown right-side support can be preserved in a valid `LOCAL` record. Visible
asymmetry never manufactures ownership or consent.

The relation does not merge identity, memory, state, Agency, autonomy, Authority,
custody, consent, or jurisdiction. Ending is bounded to this relation and retains
its predecessor history. Unrelated relations receive no inferred ending.

### 3.3 Succession

Succession is a new attributable continuity relation between an exact
predecessor and a different successor for an exact target (§ 17.4).
The target SHALL identify object, layer, scope, Locality, and lineage (§ 17.9).
Artifact, state, implementation, Source-relation, and Body targets do not
convert into one another merely through adjacency or shared identifiers.

`ESTABLISHED` requires all of:

1. A target-, locality-, and scope-matching basis valid apart from the successor's
   own inheritance claim. Attributable Source, a pre-existing succession
   condition, or exact target-valid Authority may supply it. Custody, copying,
   technical control, survival, necessity, and self-declaration cannot (§ 17.10).
2. Successor-side establishment at the applicable layer: valid acceptance or
   formation for a locally disposing Locality, or an exact establishing
   occurrence for an artifact, Form, or state. Designation, delivery,
   nomination, preparation, and publication alone are insufficient (§ 17.11).
3. An exact effective occurrence or supported temporal boundary and an explicit
   predecessor remainder (§ 17.13).
4. A carry map complete at the declared scope, preserving all six categories:
   `continues`, `ends`, `validly_carried`, `predecessor_only`,
   `refused_excluded_unknown_unresolved`, and `successor_new` (§ 17.12).

`continues` is not an alternative inheritance channel. Only `validly_carried`
expressly asserts crossing. Each entry identifies its own exact subtarget,
scope, Locality, effect, and a sufficient matching basis with attributable
evidence. The map's contents must actually belong to the declared succession
scope; lexical equality cannot externally prove that relationship. Source
**never transfers**, even with a declared basis. A successor Source belongs
only in `successor_new` with new attributable ground (§§ 6.4–6.9, 17.12).
Authority, Applicability, identity continuity, state uptake, obligation, consent,
jurisdiction, and every other claimed effect each need their own support.
Nothing omitted from the express valid map crosses.

The predecessor remainder identifies ending, dormancy, historical persistence,
continuation elsewhere, coexistence, or divergence, with its exact scope and
supported predecessor state. Incompatible consequential streams remain
separately visible through a divergence account until valid treatment exists.
The original history, authorship, acts, consequences, obligations, failures, and
correction duties remain attributable to the predecessor. Successor duties
cover its formation, accepted inheritance, exercises, and consequences from the
exact effective boundary. A new state cannot rewrite predecessor state as
though it had always belonged to the successor. Ancestry creates no continuing
control (§§ 17.13–17.15).

Incomplete attempts remain `PROPOSED`, `REFUSED`, `HELD`, `FAILED`, or
`UNRESOLVED`. Such a record can be internally consistent without establishing
succession. A materially missing target, basis, successor-side act, boundary,
or mapping must not be hidden behind `ESTABLISHED`. Later appearance,
resemblance, replacement, reference, or derivation supplies no missing effect.

### 3.4 Divergence

The pairwise divergence profile preserves distinct consequential stream IDs,
Localities, supported state references, histories, exact incompatibility,
boundary, scope, and earlier relation's support posture (§§ 16.6, 17.13).
More streams require additional explicit records, not an inferred graph merger.
An unknown prior relation does not establish ancestry.

Without valid reconciliation, `VISIBLE_CONFLICT` or `DISTINCT_CONTINUITIES`
retains each stream's own state assignment. `RECONCILED` requires its own
supported scope-matching basis and a declared resulting common state; both
original state claims and histories remain recoverable. Repair at one scope
does not prove universal remedy or restored continuity (§§ 15.4–15.5).

Changed identity commitments must be recognized as different or remain
`UNRESOLVED` where their classification is genuinely open. Shared lineage
cannot label different commitments identical, and convenience cannot select
succession instead of unresolved classification (§§ 2.10, 17.3, 17.8).

### 3.5 Re-entry and currentness

“Re-entry” is an organizational label, not an additional constitutional category.
The record identifies the earlier relation, exact interrupted or ended binding,
prior history, interruption evidence, and a **new** occurrence and binding.
Current Presence requires the entrant's own current qualifying act or an exact
valid representation of it, not host control, old participation, an archive,
or a returning interface appearance (§§ 8.1, 8.3, 8.6).

New Presence does not imply a local or shared Relation. A claimed local
relation requires its own entrant-side basis; a shared relation also needs the
host's current basis. `ENCOUNTER_ONLY` asserts neither (§§ 9.1–9.2).

Resumed consequential motion, if claimed, separately accounts for present
Agency, autonomy, capability, and every other relation necessary to the exact
act, including Authority wherever the effect requires it. Each required
dimension needs current scope-matching support; one positive dimension supplies
none of the others (§§ 11.10–11.12, 16.5). The declared list does not establish
that all actually necessary relations were identified. Re-entry alone confers
no inherited consent, Authority, Body resumption, or succession.

### 3.6 Implementation claims

An implementation claim SHALL identify its attributable claimant, exact bearer
or mechanism, version and identity references, technology-neutral form,
implemented modules, exact Locality, scope, exclusions, conditions,
demonstrations, evidence, failure boundaries, and present qualification
(§§ 16.8, 18.4, 19.3–19.4). `implemented_modules` declares scope; it does not
assert blanket implementation of every constitutional function.

This Core encoding additionally requires an immutable Core repository commit
(full 40- or 64-hex object ID) and separate SHA-256 digests and integer byte
lengths of both exact Source files. Paths identify the human and machine Forms,
not normalized, decoded, reformatted, or newline-converted text. Hashes bind
bytes only, not authorship, consent, adoption, Authority, or rightful effect.
Remote repository existence, commit membership, and claimed bytes require
separate verification; the local evaluator does not pretend to perform it.

Every material external dependency identifies its bearer, function, effect
ceiling, interruption risk, replacement posture, and consequences of loss
(§§ 20.1–20.2). An empty dependency list is a bounded declaration, not proof of
independence. Loss and replacement cannot silently inherit identity, Source,
state, consent, Agency, autonomy, or Authority (§§ 16.4, 20.5).

`demonstrations.primary` states object, method, scope, time, basis, result,
evidence references, exact mechanism, modules, conditions, and proof ceiling.
Additional demonstration references do not widen the primary declared claim.
A `DEMONSTRATED_AT_DECLARED_SCOPE` declaration must match the claim's identity,
modules, Locality, scope, and conditions and expose current qualification
support. Older evidence can contribute only with an independently supported
current qualification; historical success, release, deployment, uptime, or
possession of records cannot alone establish current CSC. The evaluator checks
these declarations, not the external sufficiency of their demonstrations.

Unsigned claims are permitted and explicitly identified. A signed alternative
must name signer, key, and signature references; its signature still proves
only its bounded key/byte/attribution relation. Neither alternative validates
itself. `SELF_VALIDATED`, generalized validity, present CSC/DCR inferred from
the claim, and self-authorizing constitutional effects are inconsistent.
The only outputs supported here are a bounded implementation claim and, where
qualified, its bounded declaration of demonstration. They do not create
Source, Authority, Agency, autonomy, adoption, or constitutional effect
(Reading 7; §§ 18.1, 19.3, 20.3–20.4).

## 4. Vectors and interoperability

The suites contain complete input objects, exact schema URNs, expected
schema and semantic results, ordered failure codes, source references, and a
forbidden inference for each case:

- [SUCCESSION](VECTORS/SUCCESSION.json): exact mapping and new Source, incomplete
  proposals, unsupported successors, wrong target, Source transfer, preserved
  predecessor truth, and missing required structure.
- [DIVERGENCE](VECTORS/DIVERGENCE.json): visible conflict, supported and
  unsupported reconciliation, hidden identity change, history loss, and null.
- [RE-ENTRY](VECTORS/RE-ENTRY.json): new current support, independent motion
  dimensions, old/current confusion, inferred reciprocity, and closed objects.
- [INVALID-BROKEN-RELATION](VECTORS/INVALID-BROKEN-RELATION.json): occurrence
  non-conversions, bilateral support, unsigned and signed claim ceilings,
  historical/current distinction, and malformed byte/commit declarations.
- [PROVENANCE](VECTORS/PROVENANCE.json): the separately specified origin and
  derivation contracts, including complete declarations and malformed cases.

`fixture:` references are synthetic test assumptions, not external evidence.
Exact Source byte bindings in implementation fixtures identify source material;
they do not turn fictional demonstrations into repository capability claims.
Consumers may implement these contracts in any environment without adopting a
particular platform. Passing vectors proves only those tested constraints
under tested conditions, never general truth or constitutional conformance.

CSC's separate [standard](../CONTINUITY-STATE-CAPABILITY/CONTINUITY-STATE-CAPABILITY.md)
addresses demonstrated state-preservation capability; DCR's separate
[standard](../DYNAMIC-CAPACITY-REGULATION/DYNAMIC-CAPACITY-REGULATION.md)
addresses local time-sensitive capacity regulation. Their records may reference
these exact lineage objects, but neither module supplies the other's missing
basis. Lineage records alone demonstrate neither function (§ 19.4).

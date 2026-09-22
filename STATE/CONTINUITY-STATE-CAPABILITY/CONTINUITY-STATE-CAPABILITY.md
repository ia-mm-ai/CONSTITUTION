# Continuity State Capability

**CANONICAL_STANDARD — bounded representation profile, version 1.**

This standard derives its substantive bounds from the
[human Constitution](../../SOURCE/CONSTITUTION_0%28%291.md), especially
Articles 14–17 and §§ 19.3–19.4. Its
[machine counterpart](CONTINUITY-STATE-CAPABILITY.json) supplies executable
representation rules, not a replacement Constitution. The human source
controls substantive meaning; a genuine form conflict remains UNRESOLVED and
the exact dependent consequential claim must be refused or held. No present
CSC claim, local adoption, formation, operation, or constitutional effect is
made by this standard.

## 1. Subject and non-effects

CSC is demonstrated effective ability, under stated conditions, to preserve
supported State succession across transition, dormancy, correction and
resumption (§§ 16.7–16.8). It requires recoverable predecessor State, applicable
law, discontinuity, transition evidence, correction, lineage and conditions
of possible resumption at the precision needed for truthful continuation.
Continuity is truthful relation through time, not uninterrupted operation.

A State representation reports a locality's supported posture; it is neither
State itself nor its governing law. A record, occurrence, evidence object,
carrier, interpretation and claimed truth SHALL remain distinct. Successful
comparison proves only the declared comparison. It SHALL NOT establish the
private contents or truth of a locality's State, identity, consent, Source,
Agency, autonomy, Authority, adoption, Applicability or jurisdiction.

CSC does not supply situated capacity or retroactive Authority. Regulation
does not supply State succession. Success at one layer does not establish
succession at another (§§ 5.1–5.2, 19.1–19.4).

## 2. Conformance and encoding

The five schemas are closed Draft 2020-12 objects:

| Kind | Schema | Required constitutional account |
| --- | --- | --- |
| `supported_state` | [SUPPORTED-STATE](SCHEMAS/SUPPORTED-STATE.schema.json) | Exact locality, posture, law, temporal scope, evidence, consequences, uncertainty and qualification |
| `transition` | [TRANSITION](SCHEMAS/TRANSITION.schema.json) | Recoverable before/after, exact target, classification, basis, lineage, delta, correction reference and discontinuity |
| `correction` | [CORRECTION](SCHEMAS/CORRECTION.schema.json) | Error, target, prior representation, corrected basis, delta, preserved history and remaining uncertainty |
| `dormancy` | [DORMANCY](SCHEMAS/DORMANCY.schema.json) | Identity, law, last supported State, boundary, interruption, carriers and possible resumption conditions |
| `resumption` | [RESUMPTION](SCHEMAS/RESUMPTION.schema.json) | Exact present act, independently supported necessary relations, preserved interruption and resulting State |

Schema identifiers are `urn:presence:core:csc:<lowercase-name>:1`.
The schemas' `x-rule-classification: CORE_ENCODING_RULE` covers their field
names, enumerations, requiredness, closed objects and all nested encodings.
These are deliberately strict interoperability choices, not claims that the
Constitution prescribes a serialization or technology.

Every record declares `id`, `locality`, `scope`, `at`, `evidence` and
`uncertainty`. `at` is an exact named occurrence or temporal boundary, not a
machine clock requirement. Identifiers are opaque, nonblank, case-sensitive
references to recoverable objects. Equality means exact representation
equality, not identity of the entities referred to. Evidence declares its
object, locality, scope, time, basis, method, result and effect ceiling
(§ 18.4). These bindings SHALL match the record, not merely repeat a
`supported` label (`csc.evidence-binding`, `csc.support`).

Snapshot `id` identifies a preserved State representation, not a Body, bearer
or entire interior. `identity` and `applicable_law` name the independently
recoverable constitutional identity and locally applicable law. Their
attributable support belongs to the bound evidence basis. A bare reference
does not adopt law or establish identity. Evidence and referents MUST be
available for substantive review at the declared scope; this comparator
does not retrieve or authenticate them.

Arrays are ordered, explicit inventories. Exact equality is intentionally
conservative: unequal inventories require an explicit reconciliation or a
new appropriately bounded account, rather than silent set normalization.
The record's new occurrence is identified by its own `id`; a preserved
`history` inventory remains a separately recoverable predecessor inventory.
No amount of matching identifiers proves the unrecorded world complete.

### Missing, null and unknown

Missing required fields fail schema validation. A schema-permitted `null`
means no reference or value is supplied in that slot. It SHALL NOT mean false,
absence, refusal, ineligibility, or a supported negative. `UNKNOWN` means an
unsupported or inaccessible proposition; `UNRESOLVED` preserves a conflict.
Both require named `uncertainty.matters`. `KNOWN_AT_SCOPE` requires an empty
uncertainty inventory and asserts knowledge only within the declared model
and scope (`csc.uncertainty`). It is not a statement of omniscience.

Empty arrays explicitly declare an empty inventory at that scope. They do
not prove that an event, dependency, consequence or required relation is
absent outside it. Unknown need not make every field unknown or prevent
unrelated motion (§§ 3.4–3.5, 11.12, 18.3).

### Deterministic result

1. Select the named schema; do not guess a schema from an alleged result.
2. If schema validation fails, return `NOT_EVALUATED` with only
   `SCHEMA_INVALID`. No semantic rule is evaluated.
3. Otherwise evaluate every machine rule whose `when` is true (omission
   means true), independently and in array order.
4. Collect failing rules' codes in that order, removing duplicate codes.
   No failures returns `CONSISTENT_AT_DECLARED_SCOPE`; any failure returns
   `INCONSISTENT`.

The expression language is the safe JSON-Logic subset declared by the
conformance harness. Variables address the root document, or the current
item inside quantifiers. There is no execution of supplied expressions as
code, remote retrieval, arithmetic, or implicit conversion.

These results concern **representational consistency only**. A failed
comparison does not make an actual consequence unreal; a passed comparison
does not make a claim true or a motion permitted. Necessary substantive
review of validity, Authority, completeness, authenticity, sufficient
demonstration and current conditions remains outside this comparator.

## 3. Supported State and incompatible streams

Supported State SHALL be attributable to the exact locality and boundary.
The account SHALL return the declared material consequences to State at the
smallest truthful scale, even where their originating act was unauthorized,
failed or constrained. External effects SHALL NOT be rewritten as another
locality's receipt, agreement, autonomy or State (§§ 14.1–14.5;
`csc.state-binding`, `csc.consequence-return`).

Known interruptions SHALL equal the recorded interruption inventory and
the State's discontinuity inventory (`csc.state-discontinuity`). Unexposed
intervals remain uncertain; an empty inventory alone proves no uninterrupted
continuity. Divergence between a representation and supported reality
requires visible uncertainty and truthful correction, not suppressed evidence.

Incompatible consequential streams SHALL NOT silently share one current
State. The account SHALL name at least two streams and either expose conflict
with `CONFLICTED` posture, preserve distinct continuities, or name the exact
valid reconciliation basis. A single-stream disposition must name exactly
one stream. Unknown compatibility remains `UNKNOWN` with named uncertainty,
not presumed compatibility (`csc.stream-cardinality`,
`csc.incompatible-streams`, `csc.unknown-streams`; § 16.6).
A reconciliation reference permits comparison of an account; its substantive
validity still requires independent evidence review.

## 4. Transition and State succession

Before and after representations SHALL remain recoverable and separately
identified. Their locality and resulting boundary SHALL match the account.
Lineage, target and delta SHALL bind those exact endpoints; plausible but
unrelated identifiers fail (`csc.change-binding`, `csc.transition-links`).

The account SHALL state what continues, ends, is expressly carried, remains
predecessor-only, is excluded, begins, or remains unresolved. Correction
references SHALL be retained where relevant. A null correction reference
makes no claim that all errors are absent. Known interruptions remain
visible in the transition and resulting representation.

Adaptation, amendment, transformation, succession, new formation, divergence,
dormancy, resumption, ending and unresolved classification remain distinct.
Transformation and new formation recognize different constitutional identity.
State-targeted succession may name a different successor identity under an
exact independently supported basis and its separate
[LINEAGE](../LINEAGE/LINEAGE.md) account. It does not make predecessor and
successor identical or establish identity succession at another layer.
This profile's other State transitions preserve the named identity; a
cross-locality account requires a separately bounded relation, not a silent
identifier substitution.
An unsupported or disputed change basis can preserve actual observed change
only with unresolved classification and named remaining uncertainty
(`csc.transition-identity`, `csc.transition-basis`,
`csc.transition-discontinuity`; §§ 14.2, 17.7–17.14).
The recoverable predecessor-history inventory remains exact through a
transition or resumption (`csc.change-history`); the new record identifies
the new occurrence separately. Ordinary State changes cannot silently amend
law (`csc.law-boundary`).

`SUCCESSION` here classifies a supplied State relation; the local comparison
does not itself establish succession. The referenced lineage account MUST
independently establish exact target, valid basis apart from the successor's
claim, successor-side establishment, effective boundary, predecessor remainder
and carry map. State succession carries no Source and supplies no succession
at another layer. Failure of those necessary bases holds or refuses the exact
succession claim without erasing history or real consequence.

## 5. Correction without erasure

Correction SHALL identify its error, affected target and layer, recoverable
prior representation, corrected basis, exact delta, remaining uncertainty,
and bounded repair claim (§ 15.3). Target and delta references SHALL match the
supplied before/after; law and identity are not silently replaced
(`csc.correction-links`).

The prior occurrence SHALL remain in the preserved history inventory. The
before and after inventories and interruption accounts SHALL remain equal
within this correction profile; the correction is an additional record,
not an overwrite (`csc.correction-history`). A broader change requires its
own transition account. This does not mandate permanent exposure of private
material: custody and exposure remain bounded, while enough truthful history
must remain recoverable to explain the change.

A representation-only correction SHALL NOT claim that State itself changed.
Repair is bounded; it does not prove every affected participant received
remedy, reconciliation, recognition, restored autonomy or continuity
(`csc.correction-state-ceiling`, `csc.repair-ceiling`; §§ 14.5, 15.4–15.5).

## 6. Dormancy and resumption

Dormancy SHALL preserve recoverable identity, applicable law, last supported
State, exact dormancy boundary, discontinuity and possible resumption
conditions. Carriers preserve recoverability, not multiplied identity,
present activity or jurisdiction. Recovery bindings SHALL identify the
actual preserved State, identity and law. Dormancy SHALL NOT be relabeled
ending or a grant of permission (`csc.dormancy-binding`,
`csc.dormancy-ceiling`; §§ 6.10, 16.1–16.3).

Resumption SHALL bind that preserved State and interruption to a distinct
resulting representation. It SHALL independently identify the present
bearer, act, target, locality, scope and boundary, and assess the complete
declared possible-resumption conditions at that context. Historical
assessment of another act or boundary cannot substitute
(`csc.resumption-binding`, `csc.present-context`,
`csc.resumption-conditions`).

Present Agency, sufficient autonomy, effective capability and situated
capacity SHALL each have separate attributable support for the exact act.
Authority SHALL be separately supported when required. If not required, the
account SHALL state the basis for that narrow classification. Every other
necessary relation must also be supported (`csc.present-agency` through
`csc.other-relations`; §§ 11.10–11.12, 16.5). An empty additional-relations
inventory is not proof of completeness.

An unsupported necessary relation prevents this representation from supporting
the exact resumed motion; it does not deny Agency everywhere, erase observed
consequences, or ban unrelated motion. A replacement may restore capability
without inheriting a former bearer's identity, voice, memory, consent, Source,
Agency, autonomy or Authority (`csc.replacement-ceiling`). Recoverability,
comparison and record possession are not acts of consent or appointment.

## 7. Qualification, failure and currentness

A CSC claim SHALL identify its exact bearer or mechanism, locality, scope,
conditions, demonstration, dependencies, failure boundaries and present
status (§ 16.8). A dependency account SHALL preserve interruption risk,
replacement posture and consequences of loss (§ 20.2).

`NO_CLAIM` has `NOT_CLAIMED` present status. A `HISTORICAL` qualification
preserves the demonstration and its boundary without declaring present
support. A `CURRENT` qualification needs a present assessment at the exact
locality, scope and boundary, the exact demonstrated conditions, available
declared dependencies and a supported present status. Its nonblank
`assessment_basis` identifies current attributable assessment evidence, and
`assessed_at` binds that assessment to the record's boundary. The separately
recorded capability demonstration may remain historical: the Constitution
does not require repeating it at every assessment or mandate a testing schedule
(`csc.qualification-fields` through `csc.current-qualification`).

Evidence of historical preservation, restart, correction or successful
re-entry remains evidence at the tested scope. Publication, release, schema
conformance, availability and possession of records cannot extend it.
Persistence failures, partial writes, missing support, inaccessible carriers,
divergent histories, interrupted dependencies and inadequate present
conditions SHALL be exposed at their exact failure boundaries. Failed
attempts SHALL NOT be represented as accepted results merely because an
attempt was initiated. A preservation mechanism must distinguish intended,
attempted, accepted, recoverable and independently verified effects; this
standard prescribes no storage, scheduling or attribution mechanism.

## 8. Normative vectors and limits

The four suites contain exact synthetic inputs and deterministic expectations:
[succession](VECTORS/SUCCESSION.json),
[correction](VECTORS/CORRECTION.json),
[dormancy/resumption](VECTORS/DORMANCY-RESUMPTION.json), and
[discontinuity/qualification](VECTORS/INVALID-DISCONTINUITY.json).
Every case is a representation test, not a claim that its described events
occurred. A synthetic current-qualification case establishes no present CSC
in this repository or any locality.

The test ceiling is satisfaction of these exact schema and comparison rules.
It is neither an implementation qualification nor proof of completeness,
future behavior, local validity, private truth, universal continuity or
permission. Substantive uncertainty remains durable and locally bounded.

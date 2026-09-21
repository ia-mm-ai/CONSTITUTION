"""Non-operative executable pressure candidate for one Local-Reality Body.

The host evaluates an already-originated proposed Act against an immutable
snapshot. It cannot originate an Act, Authority, law, Body state, or real
occurrence. A caller must provide Act identity, time, basis, uncertainty, and
the expected current revision; the host generates none of them. Its attempt
record identifier is derived only from the supplied matter and resulting
before/after hashes.

The narrow seam is:

    CURRENT SNAPSHOT + ATTRIBUTED PROPOSED ACT
        -> APPLIED | REFUSED | HOLD | UNRESOLVED
        -> ONE APPEND-ONLY ATTEMPT RECORD

The implementation deliberately contains no generic CANDIDATE/PRESENT/
STANDING lifecycle, relation engine, persistence layer, autonomous decision
maker, or constitutional-change path. It is a replaceable integrity organ,
not the Body or the source of the Body's law.
"""

from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass, fields, is_dataclass, replace
from enum import Enum
from typing import Any, Optional, Tuple


class Outcome(str, Enum):
    APPLIED = "APPLIED"
    REFUSED = "REFUSED"
    HOLD = "HOLD"
    UNRESOLVED = "UNRESOLVED"


class Posture(str, Enum):
    APPLY = "APPLY"
    REFUSE = "REFUSE"
    HOLD = "HOLD"


class EffectType(str, Enum):
    SET_VALUE = "SET_VALUE"
    SET_HOLD = "SET_HOLD"
    RELEASE_HOLD = "RELEASE_HOLD"


class Reason(str, Enum):
    APPLIED = "APPLIED"
    EXPLICIT_REFUSAL = "EXPLICIT_REFUSAL"
    EXPLICIT_HOLD = "EXPLICIT_HOLD"
    CURRENT_STATE_INVALID = "CURRENT_STATE_INVALID"
    REQUIRED_ACT_MATTER_MISSING = "REQUIRED_ACT_MATTER_MISSING"
    BASIS_NOT_EXACT = "BASIS_NOT_EXACT"
    BODY_NOT_FORMED = "BODY_NOT_FORMED"
    BODY_ID_MISMATCH = "BODY_ID_MISMATCH"
    ACT_ID_ALREADY_RECORDED = "ACT_ID_ALREADY_RECORDED"
    REVISION_MISMATCH = "REVISION_MISMATCH"
    AUTHORITY_NOT_FOUND = "AUTHORITY_NOT_FOUND"
    AUTHORITY_AMBIGUOUS = "AUTHORITY_AMBIGUOUS"
    SCOPE_NOT_AUTHORIZED = "SCOPE_NOT_AUTHORIZED"
    EFFECT_NOT_AUTHORIZED = "EFFECT_NOT_AUTHORIZED"
    PROTECTED_TARGET = "PROTECTED_TARGET"
    TARGET_NOT_PERMITTED = "TARGET_NOT_PERMITTED"
    POSTURE_EFFECT_MISMATCH = "POSTURE_EFFECT_MISMATCH"
    TARGET_VALUE_MISMATCH = "TARGET_VALUE_MISMATCH"
    TARGET_ALREADY_HELD = "TARGET_ALREADY_HELD"
    TARGET_HELD = "TARGET_HELD"
    HOLD_NOT_FOUND = "HOLD_NOT_FOUND"


PROTECTED_TARGETS = frozenset(
    {
        "body.identity",
        "body.formation",
        "body.locus",
        "body.law_floor",
        "body.authority_relations",
    }
)


@dataclass(frozen=True)
class AuthorityBinding:
    bearer: str
    function: str
    scopes: Tuple[str, ...]
    effects: Tuple[EffectType, ...]


@dataclass(frozen=True)
class StateEntry:
    target: str
    value_ref: str


@dataclass(frozen=True)
class HeldTarget:
    target: str
    set_by_act_id: str
    basis_refs: Tuple[str, ...]


@dataclass(frozen=True)
class BodyState:
    body_id: str
    status: str
    revision: int
    locus_ref: str
    law_floor_ref: str
    law_floor_source_sha256: str
    authorities: Tuple[AuthorityBinding, ...]
    permitted_targets: Tuple[str, ...]
    values: Tuple[StateEntry, ...] = ()
    holds: Tuple[HeldTarget, ...] = ()


@dataclass(frozen=True)
class ProposedAct:
    act_id: str
    body_id: str
    bearer: str
    function: str
    scope: str
    target: str
    effect_type: EffectType
    effect_ref: str
    basis_refs: Tuple[str, ...]
    uncertainty_ref: str
    originated_at: str
    expected_revision: int
    expected_target_ref: Optional[str]
    posture: Posture = Posture.APPLY


@dataclass(frozen=True)
class AttemptRecord:
    record_id: str
    sequence: int
    previous_record_sha256: Optional[str]
    act: ProposedAct
    outcome: Outcome
    reasons: Tuple[Reason, ...]
    expected_revision: int
    before_revision: int
    after_revision: int
    before_state_sha256: str
    after_state_sha256: str
    state_changed: bool


@dataclass(frozen=True)
class HostSnapshot:
    state: BodyState
    records: Tuple[AttemptRecord, ...] = ()


@dataclass(frozen=True)
class Evaluation:
    outcome: Outcome
    reasons: Tuple[Reason, ...]
    state_changed: bool
    record: AttemptRecord
    snapshot: HostSnapshot


def _canonical_value(value: Any) -> Any:
    if isinstance(value, Enum):
        return value.value
    if is_dataclass(value):
        return {
            item.name: _canonical_value(getattr(value, item.name))
            for item in fields(value)
        }
    if isinstance(value, dict):
        return {str(key): _canonical_value(item) for key, item in value.items()}
    if isinstance(value, (tuple, list)):
        return [_canonical_value(item) for item in value]
    return value


def canonical_sha256(value: Any) -> str:
    encoded = json.dumps(
        _canonical_value(value),
        ensure_ascii=False,
        separators=(",", ":"),
        sort_keys=True,
    ).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def _exact_text(value: Any) -> bool:
    return isinstance(value, str) and bool(value) and value == value.strip()


def _sha256_text(value: Any) -> bool:
    return (
        _exact_text(value)
        and len(value) == 64
        and all(character in "0123456789abcdef" for character in value)
    )


def _state_valid(state: BodyState) -> bool:
    authority_keys = tuple(
        (binding.bearer, binding.function) for binding in state.authorities
    )
    value_targets = tuple(entry.target for entry in state.values)
    hold_targets = tuple(hold.target for hold in state.holds)
    return (
        _exact_text(state.body_id)
        and _exact_text(state.status)
        and isinstance(state.revision, int)
        and state.revision >= 0
        and _exact_text(state.locus_ref)
        and _exact_text(state.law_floor_ref)
        and _sha256_text(state.law_floor_source_sha256)
        and len(authority_keys) == len(set(authority_keys))
        and all(
            _exact_text(binding.bearer)
            and _exact_text(binding.function)
            and bool(binding.scopes)
            and all(_exact_text(scope) for scope in binding.scopes)
            and len(binding.scopes) == len(set(binding.scopes))
            and bool(binding.effects)
            and all(isinstance(effect, EffectType) for effect in binding.effects)
            and len(binding.effects) == len(set(binding.effects))
            for binding in state.authorities
        )
        and len(state.permitted_targets) == len(set(state.permitted_targets))
        and all(_exact_text(target) for target in state.permitted_targets)
        and len(value_targets) == len(set(value_targets))
        and all(
            _exact_text(entry.target) and _exact_text(entry.value_ref)
            for entry in state.values
        )
        and len(hold_targets) == len(set(hold_targets))
        and all(
            _exact_text(hold.target)
            and _exact_text(hold.set_by_act_id)
            and bool(hold.basis_refs)
            for hold in state.holds
        )
    )


def _act_matter_exact(act: ProposedAct) -> bool:
    required = (
        act.act_id,
        act.body_id,
        act.bearer,
        act.function,
        act.scope,
        act.target,
        act.effect_ref,
        act.uncertainty_ref,
        act.originated_at,
    )
    return (
        all(_exact_text(value) for value in required)
        and isinstance(act.expected_revision, int)
        and act.expected_revision >= 0
        and (
            act.expected_target_ref is None
            or _exact_text(act.expected_target_ref)
        )
    )


def _basis_exact(basis_refs: Tuple[str, ...]) -> bool:
    return (
        bool(basis_refs)
        and all(_exact_text(item) for item in basis_refs)
        and len(basis_refs) == len(set(basis_refs))
    )


def _value_ref(state: BodyState, target: str) -> Optional[str]:
    for entry in state.values:
        if entry.target == target:
            return entry.value_ref
    return None


def _hold_for(state: BodyState, target: str) -> Optional[HeldTarget]:
    for hold in state.holds:
        if hold.target == target:
            return hold
    return None


def _replace_value(state: BodyState, target: str, value_ref: str) -> BodyState:
    retained = tuple(entry for entry in state.values if entry.target != target)
    values = tuple(
        sorted(retained + (StateEntry(target, value_ref),), key=lambda item: item.target)
    )
    return replace(state, revision=state.revision + 1, values=values)


def _set_hold(state: BodyState, act: ProposedAct) -> BodyState:
    holds = tuple(
        sorted(
            state.holds
            + (HeldTarget(act.target, act.act_id, act.basis_refs),),
            key=lambda item: item.target,
        )
    )
    return replace(state, revision=state.revision + 1, holds=holds)


def _release_hold(state: BodyState, target: str) -> BodyState:
    holds = tuple(hold for hold in state.holds if hold.target != target)
    return replace(state, revision=state.revision + 1, holds=holds)


class BodyIntegrityHostV0_1:
    """Pure transition guard over an immutable caller-supplied snapshot."""

    def evaluate(self, snapshot: HostSnapshot, act: ProposedAct) -> Evaluation:
        state = snapshot.state

        if not _state_valid(state):
            return self._finish(
                snapshot, act, state, Outcome.UNRESOLVED, (Reason.CURRENT_STATE_INVALID,)
            )
        if not _act_matter_exact(act):
            return self._finish(
                snapshot,
                act,
                state,
                Outcome.UNRESOLVED,
                (Reason.REQUIRED_ACT_MATTER_MISSING,),
            )
        if not _basis_exact(act.basis_refs):
            return self._finish(
                snapshot, act, state, Outcome.UNRESOLVED, (Reason.BASIS_NOT_EXACT,)
            )
        if state.status != "FORMED":
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.BODY_NOT_FORMED,)
            )
        if act.body_id != state.body_id:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.BODY_ID_MISMATCH,)
            )
        if any(record.act.act_id == act.act_id for record in snapshot.records):
            return self._finish(
                snapshot,
                act,
                state,
                Outcome.REFUSED,
                (Reason.ACT_ID_ALREADY_RECORDED,),
            )
        if act.expected_revision != state.revision:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.REVISION_MISMATCH,)
            )

        bindings = tuple(
            binding
            for binding in state.authorities
            if binding.bearer == act.bearer and binding.function == act.function
        )
        if not bindings:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.AUTHORITY_NOT_FOUND,)
            )
        if len(bindings) != 1:
            return self._finish(
                snapshot, act, state, Outcome.UNRESOLVED, (Reason.AUTHORITY_AMBIGUOUS,)
            )

        binding = bindings[0]
        if act.scope not in binding.scopes:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.SCOPE_NOT_AUTHORIZED,)
            )
        if act.effect_type not in binding.effects:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.EFFECT_NOT_AUTHORIZED,)
            )
        if act.target in PROTECTED_TARGETS:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.PROTECTED_TARGET,)
            )
        if act.target not in state.permitted_targets:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.TARGET_NOT_PERMITTED,)
            )

        if act.posture is Posture.REFUSE:
            return self._finish(
                snapshot, act, state, Outcome.REFUSED, (Reason.EXPLICIT_REFUSAL,)
            )
        if act.posture is Posture.HOLD:
            if act.effect_type is not EffectType.SET_HOLD:
                return self._finish(
                    snapshot,
                    act,
                    state,
                    Outcome.UNRESOLVED,
                    (Reason.POSTURE_EFFECT_MISMATCH,),
                )
            if _hold_for(state, act.target) is not None:
                return self._finish(
                    snapshot,
                    act,
                    state,
                    Outcome.HOLD,
                    (Reason.TARGET_ALREADY_HELD,),
                )
            return self._finish(
                snapshot,
                act,
                _set_hold(state, act),
                Outcome.HOLD,
                (Reason.EXPLICIT_HOLD,),
            )
        if act.effect_type is EffectType.SET_HOLD:
            return self._finish(
                snapshot,
                act,
                state,
                Outcome.UNRESOLVED,
                (Reason.POSTURE_EFFECT_MISMATCH,),
            )
        if act.effect_type is EffectType.RELEASE_HOLD:
            if _hold_for(state, act.target) is None:
                return self._finish(
                    snapshot, act, state, Outcome.REFUSED, (Reason.HOLD_NOT_FOUND,)
                )
            return self._finish(
                snapshot,
                act,
                _release_hold(state, act.target),
                Outcome.APPLIED,
                (Reason.APPLIED,),
            )

        if _hold_for(state, act.target) is not None:
            return self._finish(
                snapshot, act, state, Outcome.HOLD, (Reason.TARGET_HELD,)
            )
        if _value_ref(state, act.target) != act.expected_target_ref:
            return self._finish(
                snapshot,
                act,
                state,
                Outcome.REFUSED,
                (Reason.TARGET_VALUE_MISMATCH,),
            )
        return self._finish(
            snapshot,
            act,
            _replace_value(state, act.target, act.effect_ref),
            Outcome.APPLIED,
            (Reason.APPLIED,),
        )

    def _finish(
        self,
        snapshot: HostSnapshot,
        act: ProposedAct,
        after_state: BodyState,
        outcome: Outcome,
        reasons: Tuple[Reason, ...],
    ) -> Evaluation:
        before_state = snapshot.state
        before_hash = canonical_sha256(before_state)
        after_hash = canonical_sha256(after_state)
        previous_hash = (
            canonical_sha256(snapshot.records[-1]) if snapshot.records else None
        )
        sequence = len(snapshot.records) + 1
        record_basis = {
            "sequence": sequence,
            "previous_record_sha256": previous_hash,
            "act": act,
            "outcome": outcome,
            "reasons": reasons,
            "before_state_sha256": before_hash,
            "after_state_sha256": after_hash,
        }
        record_id = f"attempt-{canonical_sha256(record_basis)[:24]}"
        state_changed = before_hash != after_hash
        record = AttemptRecord(
            record_id=record_id,
            sequence=sequence,
            previous_record_sha256=previous_hash,
            act=act,
            outcome=outcome,
            reasons=reasons,
            expected_revision=act.expected_revision,
            before_revision=before_state.revision,
            after_revision=after_state.revision,
            before_state_sha256=before_hash,
            after_state_sha256=after_hash,
            state_changed=state_changed,
        )
        next_snapshot = HostSnapshot(after_state, snapshot.records + (record,))
        return Evaluation(outcome, reasons, state_changed, record, next_snapshot)

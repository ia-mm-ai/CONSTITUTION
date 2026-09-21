#!/usr/bin/env python3
"""Pressure-test the non-operative Body Integrity Host candidate.

All states and Acts in this verifier are mechanically supplied representations.
Running it creates no Body Act, Authority exercise, state change, admission,
relation, publication, or effect in RM-LOCAL-REALITY-BODY-001.
"""

from __future__ import annotations

import json
from dataclasses import replace

from body_integrity_host import (
    AuthorityBinding,
    BodyIntegrityHostV0_1,
    BodyState,
    EffectType,
    HostSnapshot,
    Outcome,
    Posture,
    ProposedAct,
    Reason,
    StateEntry,
    canonical_sha256,
)


BODY_ID = "RM-LOCAL-REALITY-BODY-001"
MARKO = "Marko Markota"
CANDIDATE_SHA256 = (
    "f61959c964b768110708d7ca98b28f87b16ea1e496ffd926cfa177b9dff3310d"
)


def initial_snapshot() -> HostSnapshot:
    state = BodyState(
        body_id=BODY_ID,
        status="FORMED",
        revision=0,
        locus_ref="bounded Reality-Model undertaking",
        law_floor_ref=(
            "FORMATION/FIRST_LOCAL_REALITY_BODY_CANDIDATE_001.md"
            "#4-candidate-body-native-law-floor"
        ),
        law_floor_source_sha256=CANDIDATE_SHA256,
        authorities=(
            AuthorityBinding(
                bearer=MARKO,
                function="LOCAL_GOVERNING_AUTHORITY",
                scopes=("BODY_LOCAL_ORIENTATION",),
                effects=(EffectType.SET_VALUE, EffectType.SET_HOLD),
            ),
            AuthorityBinding(
                bearer=MARKO,
                function="ADAPTATION_AUTHORITY",
                scopes=("NON_CONSTITUTIVE_OPERATION",),
                effects=(
                    EffectType.SET_VALUE,
                    EffectType.SET_HOLD,
                    EffectType.RELEASE_HOLD,
                ),
            ),
        ),
        permitted_targets=("carrier_pointer", "working_orientation"),
        values=(
            StateEntry("carrier_pointer", "carrier://initial"),
            StateEntry("working_orientation", "orientation://initial"),
        ),
    )
    return HostSnapshot(state)


def act(
    act_id: str,
    *,
    target: str = "working_orientation",
    effect_type: EffectType = EffectType.SET_VALUE,
    effect_ref: str = "orientation://next",
    expected_revision: int = 0,
    expected_target_ref: str | None = "orientation://initial",
    bearer: str = MARKO,
    function: str = "LOCAL_GOVERNING_AUTHORITY",
    scope: str = "BODY_LOCAL_ORIENTATION",
    body_id: str = BODY_ID,
    basis_refs: tuple[str, ...] = ("basis://deliberate-local-act",),
    uncertainty_ref: str = "uncertainty://none-known",
    posture: Posture = Posture.APPLY,
) -> ProposedAct:
    return ProposedAct(
        act_id=act_id,
        body_id=body_id,
        bearer=bearer,
        function=function,
        scope=scope,
        target=target,
        effect_type=effect_type,
        effect_ref=effect_ref,
        basis_refs=basis_refs,
        uncertainty_ref=uncertainty_ref,
        originated_at="2026-09-01T20:00:00Z",
        expected_revision=expected_revision,
        expected_target_ref=expected_target_ref,
        posture=posture,
    )


def value(snapshot: HostSnapshot, target: str) -> str | None:
    for entry in snapshot.state.values:
        if entry.target == target:
            return entry.value_ref
    return None


def main() -> int:
    host = BodyIntegrityHostV0_1()
    checks: dict[str, bool] = {}

    base = initial_snapshot()
    base_hash = canonical_sha256(base.state)

    applied = host.evaluate(base, act("act-apply-001"))
    checks["exact_authorized_act_applies"] = applied.outcome is Outcome.APPLIED
    checks["applied_act_advances_revision"] = applied.snapshot.state.revision == 1
    checks["applied_act_changes_exact_target"] = (
        value(applied.snapshot, "working_orientation") == "orientation://next"
    )
    checks["predecessor_snapshot_remains_unchanged"] = (
        canonical_sha256(base.state) == base_hash
        and value(base, "working_orientation") == "orientation://initial"
    )
    checks["applied_attempt_has_before_after_lineage"] = (
        applied.record.before_state_sha256 == base_hash
        and applied.record.after_state_sha256
        == canonical_sha256(applied.snapshot.state)
        and applied.record.before_state_sha256 != applied.record.after_state_sha256
    )

    deterministic = host.evaluate(base, act("act-apply-001"))
    checks["same_input_is_deterministic"] = deterministic == applied

    stale = host.evaluate(base, act("act-stale", expected_revision=9))
    checks["stale_revision_is_refused"] = (
        stale.outcome is Outcome.REFUSED
        and stale.reasons == (Reason.REVISION_MISMATCH,)
        and not stale.state_changed
    )

    wrong_bearer = host.evaluate(base, act("act-wrong-bearer", bearer="Someone Else"))
    checks["unknown_bearer_cannot_acquire_authority"] = (
        wrong_bearer.reasons == (Reason.AUTHORITY_NOT_FOUND,)
    )

    wrong_function = host.evaluate(
        base,
        act(
            "act-function-merge",
            effect_type=EffectType.RELEASE_HOLD,
            effect_ref="release://requested",
        ),
    )
    checks["same_bearer_functions_do_not_merge"] = (
        wrong_function.reasons == (Reason.EFFECT_NOT_AUTHORIZED,)
    )

    foreign_body = host.evaluate(base, act("act-foreign-body", body_id="OTHER-BODY"))
    checks["body_boundary_is_exact"] = (
        foreign_body.reasons == (Reason.BODY_ID_MISMATCH,)
    )

    missing_basis = host.evaluate(base, act("act-no-basis", basis_refs=()))
    checks["missing_basis_remains_unresolved"] = (
        missing_basis.outcome is Outcome.UNRESOLVED
        and missing_basis.reasons == (Reason.BASIS_NOT_EXACT,)
    )

    protected = host.evaluate(
        base,
        act(
            "act-protected-target",
            target="body.law_floor",
            expected_target_ref=None,
            effect_ref="law://replacement",
        ),
    )
    checks["host_cannot_change_law_floor"] = (
        protected.reasons == (Reason.PROTECTED_TARGET,)
    )

    explicit_refusal = host.evaluate(
        base, act("act-explicit-refusal", posture=Posture.REFUSE)
    )
    checks["explicit_refusal_is_recorded_without_target_change"] = (
        explicit_refusal.outcome is Outcome.REFUSED
        and explicit_refusal.reasons == (Reason.EXPLICIT_REFUSAL,)
        and not explicit_refusal.state_changed
        and len(explicit_refusal.snapshot.records) == 1
    )

    hold_act = act(
        "act-hold",
        effect_type=EffectType.SET_HOLD,
        effect_ref="hold://orientation-review",
        expected_target_ref="orientation://initial",
        posture=Posture.HOLD,
    )
    held = host.evaluate(base, hold_act)
    checks["hold_is_target_specific_and_changes_only_host_posture"] = (
        held.outcome is Outcome.HOLD
        and held.state_changed
        and held.snapshot.state.revision == 1
        and tuple(item.target for item in held.snapshot.state.holds)
        == ("working_orientation",)
        and value(held.snapshot, "working_orientation") == "orientation://initial"
    )

    blocked_by_hold = host.evaluate(
        held.snapshot,
        act(
            "act-held-mutation",
            expected_revision=1,
            expected_target_ref="orientation://initial",
        ),
    )
    checks["held_target_does_not_progress"] = (
        blocked_by_hold.outcome is Outcome.HOLD
        and blocked_by_hold.reasons == (Reason.TARGET_HELD,)
        and not blocked_by_hold.state_changed
    )

    other_target = host.evaluate(
        held.snapshot,
        act(
            "act-other-target",
            target="carrier_pointer",
            effect_ref="carrier://next",
            expected_revision=1,
            expected_target_ref="carrier://initial",
            function="ADAPTATION_AUTHORITY",
            scope="NON_CONSTITUTIVE_OPERATION",
        ),
    )
    checks["hold_on_one_target_does_not_freeze_body"] = (
        other_target.outcome is Outcome.APPLIED
        and value(other_target.snapshot, "carrier_pointer") == "carrier://next"
        and value(other_target.snapshot, "working_orientation")
        == "orientation://initial"
    )

    released = host.evaluate(
        other_target.snapshot,
        act(
            "act-release-hold",
            effect_type=EffectType.RELEASE_HOLD,
            effect_ref="release://orientation-review-complete",
            expected_revision=2,
            expected_target_ref="orientation://initial",
            function="ADAPTATION_AUTHORITY",
            scope="NON_CONSTITUTIVE_OPERATION",
        ),
    )
    checks["hold_release_requires_separate_authorized_act"] = (
        released.outcome is Outcome.APPLIED
        and released.snapshot.state.revision == 3
        and not released.snapshot.state.holds
    )

    first_use = host.evaluate(base, act("act-replay"))
    replay = host.evaluate(
        first_use.snapshot,
        act(
            "act-replay",
            expected_revision=1,
            expected_target_ref="orientation://next",
            effect_ref="orientation://later",
        ),
    )
    checks["act_identity_cannot_be_replayed"] = (
        replay.reasons == (Reason.ACT_ID_ALREADY_RECORDED,)
    )
    checks["refused_replay_is_still_recorded"] = (
        len(replay.snapshot.records) == 2
        and replay.record.previous_record_sha256
        == canonical_sha256(first_use.record)
    )

    invalid_state = replace(base.state, law_floor_source_sha256="not-a-hash")
    invalid = host.evaluate(HostSnapshot(invalid_state), act("act-invalid-state"))
    checks["invalid_current_state_cannot_be_patched_by_act"] = (
        invalid.outcome is Outcome.UNRESOLVED
        and invalid.reasons == (Reason.CURRENT_STATE_INVALID,)
    )

    checks["identity_floor_survives_all_applied_host_motion"] = all(
        snapshot.state.body_id == BODY_ID
        and snapshot.state.law_floor_source_sha256 == CANDIDATE_SHA256
        and snapshot.state.law_floor_ref == base.state.law_floor_ref
        and snapshot.state.authorities == base.state.authorities
        for snapshot in (
            applied.snapshot,
            held.snapshot,
            other_target.snapshot,
            released.snapshot,
        )
    )
    checks["every_evaluation_appends_exactly_one_record"] = all(
        len(result.snapshot.records) == 1
        for result in (
            applied,
            stale,
            wrong_bearer,
            wrong_function,
            foreign_body,
            missing_basis,
            protected,
            explicit_refusal,
            held,
            invalid,
        )
    )

    passed = all(checks.values())
    result = {
        "test_id": "BODY-INTEGRITY-HOST-001",
        "status": "PASS" if passed else "FAIL",
        "check_count": len(checks),
        "checks": checks,
        "observed": {
            "outcomes_exercised": sorted(
                {
                    item.outcome.value
                    for item in (
                        applied,
                        stale,
                        missing_basis,
                        held,
                        blocked_by_hold,
                    )
                }
            ),
            "base_state_sha256": base_hash,
            "applied_state_sha256": canonical_sha256(applied.snapshot.state),
            "final_hold_chain_revision": released.snapshot.state.revision,
            "record_chain_length": len(replay.snapshot.records),
        },
        "effect": "NONE_OUTSIDE_SYNTHETIC_PRESSURE",
        "non_effects": [
            "NO_BODY_ACT",
            "NO_BODY_STATE_CHANGE",
            "NO_AUTHORITY_EXERCISE",
            "NO_CONSTITUTIONAL_CHANGE",
            "NO_BODY_ROOT_WRITE",
            "NO_RELATION_OR_EVENT",
            "NO_EXTERNAL_PUBLICATION",
        ],
    }
    print(json.dumps(result, indent=2, sort_keys=True))
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())

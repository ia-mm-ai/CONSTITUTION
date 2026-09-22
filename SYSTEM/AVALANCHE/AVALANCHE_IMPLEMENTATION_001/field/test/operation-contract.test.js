import test from "node:test";
import assert from "node:assert/strict";
import {
  OPERATION_CONTRACT,
  VM_EFFECTS,
  VM_SCOPES,
  operationDefinition
} from "../src/constants.js";
import { assertOperationScope } from "../src/scope.js";

test("FIELD consumes the exhaustive shared operation/effect/scope inventory", () => {
  assert.equal(OPERATION_CONTRACT.length, 21);
  assert.equal(new Set(OPERATION_CONTRACT.map(({ operation }) => operation)).size, 21);
  for (const { operation, effect, scope } of OPERATION_CONTRACT) {
    assert.equal(VM_EFFECTS[operation], effect);
    assert.equal(VM_SCOPES[operation], scope);
    assert.deepEqual(operationDefinition(operation), { operation, effect, scope });
  }
  assert.throws(() => operationDefinition("UNKNOWN_OPERATION"), /unsupported operation/);
});

test("every operation scope fails closed on mismatched loci", () => {
  const target = "a".repeat(64);
  const state = {
    active_locus_id: "LOCUS-CURRENT",
    body: { posture: "ACTIVE_P1" },
    events: { [target]: { locus_id: "LOCUS-TARGET" } }
  };
  for (const operation of OPERATION_CONTRACT.filter(({ scope }) => scope === "ACTIVE_LOCUS").map(({ operation }) => operation)) {
    assert.throws(() => assertOperationScope({ operation, locusId: "", payload: {}, state }), /ACTIVE_LOCUS/);
    assert.throws(() => assertOperationScope({ operation, locusId: "LOCUS-OTHER", payload: {}, state }), /ACTIVE_LOCUS/);
  }
  for (const operation of OPERATION_CONTRACT.filter(({ scope }) => scope === "BODY_LOCAL").map(({ operation }) => operation)) {
    assert.throws(() => assertOperationScope({ operation, locusId: "LOCUS-CURRENT", payload: {}, state }), /BODY_LOCAL/);
  }
  for (const operation of OPERATION_CONTRACT.filter(({ scope }) => scope === "DORMANT_BODY").map(({ operation }) => operation)) {
    assert.throws(() => assertOperationScope({ operation, locusId: "", payload: {}, state }), /DORMANT_BODY/);
  }
  assert.doesNotThrow(() => assertOperationScope({
    operation: "CORRECT",
    locusId: "LOCUS-TARGET",
    payload: { target_transition_id: target },
    state
  }));
  assert.throws(() => assertOperationScope({
    operation: "CORRECT",
    locusId: "LOCUS-CURRENT",
    payload: { target_transition_id: target },
    state
  }), /TARGET_SCOPED/);
});

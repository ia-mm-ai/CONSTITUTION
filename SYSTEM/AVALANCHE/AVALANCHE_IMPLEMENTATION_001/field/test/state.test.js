import test from "node:test";
import assert from "node:assert/strict";
import { deriveActorContext } from "../src/state.js";
import { keyMaterial, presentState } from "./helpers.js";

test("entered actor is PRESENT and checkpoint immediately narrows phase", () => {
  const keys = keyMaterial();
  const state = presentState(keys.actorId, keys.publicKeyHex);
  assert.equal(deriveActorContext(state, keys.actorId, 1000).phase, "PRESENT");
  const checkpointId = "f".repeat(64);
  state.departure_checkpoints[checkpointId] = {
    checkpoint_id: checkpointId,
    participant_id: keys.actorId,
    status: "OPEN"
  };
  state.loci[state.active_locus_id].entries[keys.actorId].departure_checkpoint_id = checkpointId;
  assert.equal(deriveActorContext(state, keys.actorId, 1000).phase, "DEPARTURE_CHECKPOINTED");
});

test("capacity deficit overrides an otherwise present phase", () => {
  const keys = keyMaterial();
  const state = presentState(keys.actorId, keys.publicKeyHex);
  state.body.posture = "HOLD_CAPACITY_DEFICIT";
  assert.equal(deriveActorContext(state, keys.actorId, 1000).phase, "HOLD_CAPACITY_DEFICIT");
});

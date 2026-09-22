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

test("an exited entry cannot override a fresh presentation and reentry offer in the same open locus", () => {
  const keys = keyMaterial();
  const state = presentState(keys.actorId, keys.publicKeyHex);
  const locus = state.loci[state.active_locus_id];
  const entry = locus.entries[keys.actorId];
  entry.status = "EXITED";
  entry.departure_checkpoint_id = "f".repeat(64);
  state.departure_checkpoints[entry.departure_checkpoint_id] = {
    checkpoint_id: entry.departure_checkpoint_id, participant_id: keys.actorId,
    status: "SEALED", consumed_by_entry_transition_id: ""
  };
  assert.equal(deriveActorContext(state, keys.actorId, 1000).phase, "ENDED");
  const fresh = { ...locus.presentations[keys.actorId], presentation_id: "9".repeat(64) };
  locus.presentations[keys.actorId] = fresh;
  assert.equal(deriveActorContext(state, keys.actorId, 1000).phase, "PRESENTED");
  locus.gates[keys.actorId] = {
    presentation_id: fresh.presentation_id, disposition: "ADMIT", offer_status: "OPEN", offer_expires_at: 2000
  };
  const context = deriveActorContext(state, keys.actorId, 1000);
  assert.equal(context.phase, "REENTRY_AVAILABLE");
  assert.equal(context.entry, null);
  assert.equal(context.reentry_checkpoint.checkpoint_id, entry.departure_checkpoint_id);
});

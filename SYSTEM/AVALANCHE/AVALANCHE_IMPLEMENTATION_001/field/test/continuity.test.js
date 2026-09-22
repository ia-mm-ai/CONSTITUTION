import test from "node:test";
import assert from "node:assert/strict";
import { buildPassage, passageCommitment, successorStateCommitment } from "../src/continuity.js";

test("continuity commitment matches independent frozen vector", () => {
  const participant = `PRESENCE-AVALANCHE-ACTOR-${"a".repeat(40)}`;
  const from = "0".repeat(64);
  const delta = "1".repeat(64);
  const successor = successorStateCommitment(participant, from, delta, 1);
  assert.equal(successor, "3d28f3687952061ace74340738a6be829b57fa87f01ba30b22e4f6a2bc1ad3f5");
  const built = buildPassage(participant, from, [delta]);
  assert.equal(built.result_state_commitment, successor);
  assert.equal(
    passageCommitment(participant, from, built.passage, successor),
    "521eee0165e800c4465d9744a42e97d40ff1c525d3364d8b7ebca64d0401491f"
  );
});

test("continuity passage refuses a broken successor edge", () => {
  assert.throws(() => passageCommitment(
    "PARTICIPANT-001",
    "0".repeat(64),
    [{ delta_sha256: "1".repeat(64), successor_state_commitment: "2".repeat(64) }],
    "2".repeat(64)
  ), /does not succeed/);
});

import test from "node:test";
import assert from "node:assert/strict";
import { buildPassage, passageCommitment, successorStateCommitment } from "../src/continuity.js";

test("continuity commitment matches independent frozen vector", () => {
  const participant = `PRESENCE-ACTOR-${"a".repeat(40)}`;
  const from = "0".repeat(64);
  const delta = "1".repeat(64);
  const successor = successorStateCommitment(participant, from, delta, 1);
  assert.equal(successor, "b684073fe2b00e1b9f37ac7979411bd7712b0269f27d182b262c0c21a110ca66");
  const built = buildPassage(participant, from, [delta]);
  assert.equal(built.result_state_commitment, successor);
  assert.equal(
    passageCommitment(participant, from, built.passage, successor),
    "fc643669ea14c3a6e4a75fad23068f3c65d160897a2a08d7b0b565956867d26b"
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

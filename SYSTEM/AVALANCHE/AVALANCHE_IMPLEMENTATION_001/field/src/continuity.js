import { createHash } from "node:crypto";
import { requireDigest } from "./canonical.js";

export const CONTINUITY_STEP_SCHEMA = "PRESENCE_AVALANCHE_PARTICIPANT_STATE_SUCCESSION_001";
export const CONTINUITY_PASSAGE_SCHEMA = "PRESENCE_AVALANCHE_PARTICIPANT_STATE_PASSAGE_001";

function goJSON(value) {
  return Buffer.from(JSON.stringify(value), "utf8");
}

function digest(value) {
  return createHash("sha256").update(goJSON(value)).digest("hex");
}

export function successorStateCommitment(participantId, predecessorStateCommitment, deltaSha256, ordinal) {
  if (!participantId || typeof participantId !== "string") throw new Error("participant_id is required");
  requireDigest(predecessorStateCommitment, "predecessor_state_commitment");
  requireDigest(deltaSha256, "delta_sha256");
  if (!Number.isInteger(ordinal) || ordinal < 1 || ordinal > 128) throw new Error("ordinal must be 1..128");
  return digest({
    schema: CONTINUITY_STEP_SCHEMA,
    participant_id: participantId,
    ordinal,
    predecessor_state_commitment: predecessorStateCommitment,
    delta_sha256: deltaSha256
  });
}

export function buildPassage(participantId, fromStateCommitment, deltaDigests) {
  requireDigest(fromStateCommitment, "from_state_commitment");
  if (!Array.isArray(deltaDigests) || deltaDigests.length > 128) throw new Error("delta digests must be an array of at most 128 items");
  let current = fromStateCommitment;
  const passage = deltaDigests.map((deltaSha256, index) => {
    requireDigest(deltaSha256, `delta_sha256[${index}]`);
    current = successorStateCommitment(participantId, current, deltaSha256, index + 1);
    return { delta_sha256: deltaSha256, successor_state_commitment: current };
  });
  return { passage, result_state_commitment: current };
}

export function passageCommitment(participantId, fromStateCommitment, passage, expectedResult) {
  requireDigest(fromStateCommitment, "from_state_commitment");
  requireDigest(expectedResult, "result_state_commitment");
  if (!Array.isArray(passage) || passage.length > 128) throw new Error("passage must be an array of at most 128 items");
  let current = fromStateCommitment;
  for (let index = 0; index < passage.length; index += 1) {
    const step = passage[index];
    const expected = successorStateCommitment(participantId, current, step.delta_sha256, index + 1);
    if (step.successor_state_commitment !== expected) throw new Error(`passage step ${index} does not succeed its predecessor`);
    current = expected;
  }
  if (current !== expectedResult) throw new Error("passage result does not equal required state commitment");
  return digest({
    schema: CONTINUITY_PASSAGE_SCHEMA,
    participant_id: participantId,
    from_state_commitment: fromStateCommitment,
    passage,
    result_state_commitment: expectedResult
  });
}

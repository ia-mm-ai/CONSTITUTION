import { createPublicKey, verify } from "node:crypto";
import { canonicalize, sha256Hex } from "./canonical.js";
import { VM_EFFECTS, VM_TRANSITION_SCHEMA } from "./constants.js";

const ED25519_SPKI_PREFIX = Buffer.from("302a300506032b6570032100", "hex");
const UNSIGNED_FIELDS = [
  "schema", "operation", "revision", "previous_state_commitment", "actor_id",
  "actor_public_key", "nonce", "locus_id", "observed_at", "effect", "payload"
];

function exactFields(value, fields, label) {
  const actual = Object.keys(value).sort();
  const expected = [...fields].sort();
  if (actual.length !== expected.length || actual.some((field, index) => field !== expected[index])) {
    throw new Error(`${label} fields do not match the frozen contract`);
  }
}

export function signingBytes(unsigned) {
  // PRESENCE_AVALANCHE_VM_001 signs Go encoding/json output of the unsigned
  // struct in declared field order. The signature is produced OUTSIDE the VM,
  // so the FIELD must reproduce those exact bytes.
  const ordered = {
    schema: unsigned.schema,
    operation: unsigned.operation,
    revision: unsigned.revision,
    previous_state_commitment: unsigned.previous_state_commitment,
    actor_id: unsigned.actor_id,
    actor_public_key: unsigned.actor_public_key,
    nonce: unsigned.nonce,
    locus_id: unsigned.locus_id,
    observed_at: unsigned.observed_at,
    effect: unsigned.effect,
    payload: unsigned.payload
  };
  return Buffer.from(JSON.stringify(ordered), "utf8");
}

export function validateUnsigned(unsigned) {
  if (!unsigned || typeof unsigned !== "object" || Array.isArray(unsigned)) throw new Error("unsigned transition must be an object");
  exactFields(unsigned, UNSIGNED_FIELDS, "unsigned transition");
  if (unsigned.schema !== VM_TRANSITION_SCHEMA) throw new Error("unsigned transition schema mismatch");
  const effect = VM_EFFECTS[unsigned.operation];
  if (!effect || unsigned.effect !== effect) throw new Error("unsigned transition effect mismatch");
  if (!Number.isSafeInteger(unsigned.revision) || unsigned.revision < 1) throw new Error("unsigned revision is invalid");
  if (!Number.isSafeInteger(unsigned.nonce) || unsigned.nonce < 0) throw new Error("unsigned nonce is invalid");
  if (!Number.isSafeInteger(unsigned.observed_at) || unsigned.observed_at < 0) throw new Error("unsigned observed_at is invalid");
  if (!/^[0-9a-f]{64}$/.test(unsigned.previous_state_commitment)) throw new Error("unsigned previous state commitment is invalid");
  if (!/^[0-9a-f]{64}$/.test(unsigned.actor_public_key)) throw new Error("unsigned public key is invalid");
  if (typeof unsigned.actor_id !== "string" || !unsigned.actor_id) throw new Error("unsigned actor_id is invalid");
  if (typeof unsigned.locus_id !== "string") throw new Error("unsigned locus_id is invalid");
  if (!unsigned.payload || typeof unsigned.payload !== "object" || Array.isArray(unsigned.payload)) throw new Error("unsigned payload must be an object");
  return unsigned;
}

export function verifyTransitionSignature(transition) {
  if (!transition || typeof transition !== "object") throw new Error("transition must be an object");
  exactFields(transition, ["unsigned", "signature"], "transition");
  const unsigned = validateUnsigned(transition.unsigned);
  if (!/^[0-9a-f]{128}$/.test(transition.signature ?? "")) throw new Error("signature must be 64-byte lowercase hex");
  const publicKey = createPublicKey({
    key: Buffer.concat([ED25519_SPKI_PREFIX, Buffer.from(unsigned.actor_public_key, "hex")]),
    format: "der",
    type: "spki"
  });
  if (!verify(null, signingBytes(unsigned), publicKey, Buffer.from(transition.signature, "hex"))) {
    throw new Error("invalid Ed25519 transition signature");
  }
  return true;
}

export function normalizedUnsigned(unsigned) {
  validateUnsigned(unsigned);
  return {
    schema: unsigned.schema,
    operation: unsigned.operation,
    revision: unsigned.revision,
    previous_state_commitment: unsigned.previous_state_commitment,
    actor_id: unsigned.actor_id,
    actor_public_key: unsigned.actor_public_key,
    nonce: unsigned.nonce,
    locus_id: unsigned.locus_id,
    observed_at: unsigned.observed_at,
    effect: unsigned.effect,
    payload: unsigned.payload
  };
}

export function normalizedTransition(unsigned, signature) {
  const normalized = normalizedUnsigned(unsigned);
  if (!/^[0-9a-f]{128}$/.test(signature ?? "")) throw new Error("signature must be 64-byte lowercase hex");
  return { unsigned: normalized, signature };
}

export function transitionID(transition) {
  verifyTransitionSignature(transition);
  return sha256Hex(Buffer.from(JSON.stringify(normalizedTransition(transition.unsigned, transition.signature)), "utf8"));
}

export function draftCommitment(unsigned) {
  validateUnsigned(unsigned);
  return sha256Hex(canonicalize(unsigned));
}

// signUnsigned produces the external ED25519 signature over the exact signing
// bytes. It is used by the coupled path and by tests. The private key is
// supplied by the caller (an external protected signer); it is never stored.
export function signUnsigned(unsigned, sign, privateKey) {
  validateUnsigned(unsigned);
  return sign(null, signingBytes(unsigned), privateKey).toString("hex");
}

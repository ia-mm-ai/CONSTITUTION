import test from "node:test";
import assert from "node:assert/strict";
import { sign } from "node:crypto";
import { goJSONStringify } from "../src/go-json.js";
import { canonicalize, sha256Hex } from "../src/canonical.js";
import { signingBytes, normalizedTransition, transitionID, verifyTransitionSignature } from "../src/signature.js";
import { VM_EFFECTS, VM_TRANSITION_SCHEMA } from "../src/constants.js";
import { keyMaterial } from "./helpers.js";

test("signed wire JSON matches Go HTML and Unicode separator escaping", () => {
  const text = "A&B <>\u2028\u2029";
  assert.equal(goJSONStringify({ claim: text }), '{"claim":"A\\u0026B \\u003c\\u003e\\u2028\\u2029"}');
  assert.equal(goJSONStringify({ "<&>": ["\\u003c", text, "😀"] }), '{"\\u003c\\u0026\\u003e":["\\\\u003c","A\\u0026B \\u003c\\u003e\\u2028\\u2029","😀"]}');
  assert.deepEqual(JSON.parse(goJSONStringify({ claim: text })), { claim: text });
});

test("wire encoding does not alter local canonical commitment serialization", () => {
  const payload = { claim: "A&B <>\u2028\u2029" };
  assert.equal(canonicalize(payload), JSON.stringify(payload));
  assert.notEqual(canonicalize(payload), goJSONStringify(payload));
});

test("signing bytes and transition ID consistently encode Go-compatible envelopes", () => {
  const keys = keyMaterial();
  const unsigned = {
    schema: VM_TRANSITION_SCHEMA, operation: "OBSERVE_CROSSING", revision: 1,
    previous_state_commitment: "a".repeat(64), actor_id: keys.actorId,
    actor_public_key: keys.publicKeyHex, nonce: 0, locus_id: "LOCUS-001", observed_at: 1001,
    effect: VM_EFFECTS.OBSERVE_CROSSING,
    payload: { claim: "A&B <>\u2028\u2029", content_sha256: "b".repeat(64), matter_id: "MATTER-001", media_type: "text/plain" }
  };
  const bytes = signingBytes(unsigned);
  assert.ok(bytes.toString().includes('"claim":"A\\u0026B \\u003c\\u003e\\u2028\\u2029"'));
  assert.doesNotMatch(bytes.toString(), /[<>&\u2028\u2029]/);
  const signature = sign(null, bytes, keys.privateKey).toString("hex");
  const transition = normalizedTransition(unsigned, signature);
  assert.equal(verifyTransitionSignature(transition), true);
  assert.equal(transitionID(transition), sha256Hex(Buffer.from(goJSONStringify(transition))));
  assert.notEqual(transitionID(transition), sha256Hex(Buffer.from(JSON.stringify(transition))));
  const legacySignature = sign(null, Buffer.from(JSON.stringify(unsigned)), keys.privateKey).toString("hex");
  assert.throws(() => verifyTransitionSignature({ unsigned, signature: legacySignature }), /invalid Ed25519/);
});

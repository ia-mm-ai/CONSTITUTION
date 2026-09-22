import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, readFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { sign } from "node:crypto";
import { AppendOnlyJournal } from "../src/journal.js";
import { MediumEngine } from "../src/engine.js";
import { loadProfile } from "../src/profile.js";
import { signingBytes, transitionID } from "../src/signature.js";
import { VM_ID, VM_RPCCHAINVM_PROTOCOL, VM_TRANSITION_SCHEMA, VM_VERSION } from "../src/constants.js";
import { CapabilityBroker } from "../src/broker.js";
import { keyMaterial, presentState, statusFor } from "./helpers.js";

async function fixture() {
  const keys = keyMaterial();
  let state = presentState(keys.actorId, keys.publicKeyHex);
  let status = statusFor(state);
  let submitted;
  const client = {
    status: async () => structuredClone(status),
    state: async () => structuredClone(state),
    draft: async (request) => ({
      unsigned: {
        schema: VM_TRANSITION_SCHEMA,
        operation: request.operation,
        revision: state.revision + 1,
        previous_state_commitment: state.state_commitment,
        actor_id: request.actor_id,
        actor_public_key: request.actor_public_key,
        nonce: state.next_nonces[keys.actorId],
        locus_id: request.locus_id,
        observed_at: request.observed_at,
        effect: "RECORDS_CROSSING_ONLY",
        payload: request.payload
      }
    }),
    submit: async (transition) => {
      submitted = transition;
      return { transition_id: transitionID(transition), status: "PENDING_CONSENSUS", effect: "NO_ACCEPTANCE_UNTIL_BLOCK_ACCEPTED" };
    }
  };
  const directory = await mkdtemp(join(tmpdir(), "presence-avalanche-field-engine-"));
  const journal = await new AppendOnlyJournal(join(directory, "events.ndjson")).initialize();
  const profile = await loadProfile(resolve("profiles/AI_MEDIUM_001.json"));
  const config = {
    expected_locality_id: state.host_locality_id,
    expected_vm_id: VM_ID,
    expected_vm_version: VM_VERSION,
    expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
    actor_id: keys.actorId,
    actor_public_key: keys.publicKeyHex,
    receipt_timeout_ms: 1000
  };
  const engine = new MediumEngine({ config, profile, client, journal });
  return {
    keys, client, engine, journal,
    get state() { return state; },
    setState(value) { state = value; status = statusFor(state); },
    submitted: () => submitted,
    directory
  };
}

test("broker grants a controlled AI tool only while accepted entry is PRESENT", async () => {
  const f = await fixture();
  const broker = new CapabilityBroker(f.engine).register("AI_BOUNDED_TOOL_BROKER", async (input) => ({ echoed: input }));
  assert.deepEqual(await broker.invoke("AI_BOUNDED_TOOL_BROKER", "bounded"), { echoed: "bounded" });
  const ended = structuredClone(f.state);
  ended.loci[ended.active_locus_id].entries[f.keys.actorId].status = "EXITED";
  ended.entry_history["d".repeat(64)].status = "EXITED";
  ended.revision += 1;
  ended.state_commitment = "b".repeat(64);
  f.setState(ended);
  await assert.rejects(() => broker.invoke("AI_BOUNDED_TOOL_BROKER", "after-exit"), /denied/);
});

test("any pending consensus transition suspends substantive broker grants", async () => {
  const f = await fixture();
  f.setState(f.state);
  f.client.status = async () => statusFor(f.state, { pending_transitions: 1 });
  const broker = new CapabilityBroker(f.engine).register("AI_BOUNDED_TOOL_BROKER", async () => "should-not-run");
  await assert.rejects(() => broker.invoke("AI_BOUNDED_TOOL_BROKER", null), /PENDING_TRANSITION/);
  assert.equal((await f.engine.decide("READ_PUBLIC_STATE")).allowed, true);
});

test("draft is exact, externally signed and submitted without exposing private key", async () => {
  const f = await fixture();
  const draft = await f.engine.prepareDraft({
    operation: "OBSERVE_CROSSING",
    locus_id: f.state.active_locus_id,
    observed_at: 1001,
    payload: { matter_id: "MATTER-001", content_sha256: "4".repeat(64), media_type: "application/octet-stream", claim: "COMMITMENT_ONLY" }
  });
  const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
  const response = await f.engine.submitTransition({ unsigned: draft.unsigned, signature });
  assert.equal(response.status, "PENDING_CONSENSUS");
  assert.equal(f.submitted().signature, signature);
  const journal = await readFile(join(f.directory, "events.ndjson"), "utf8");
  assert.doesNotMatch(journal, /private_key/i);
  assert.doesNotMatch(journal, new RegExp(signature));
});

test("state change between draft and signature submission revokes authorization", async () => {
  const f = await fixture();
  const draft = await f.engine.prepareDraft({
    operation: "OBSERVE_CROSSING",
    locus_id: f.state.active_locus_id,
    observed_at: 1001,
    payload: { matter_id: "MATTER-002", content_sha256: "5".repeat(64), media_type: "text/plain", claim: "COMMITMENT_ONLY" }
  });
  const changed = structuredClone(f.state);
  changed.revision += 1;
  changed.state_commitment = "b".repeat(64);
  f.setState(changed);
  const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
  await assert.rejects(
    () => f.engine.submitTransition({ unsigned: draft.unsigned, signature }),
    /EXPECTED_STATE_COMMITMENT_IS_STALE/
  );
});

test("altered signature is refused before network submission", async () => {
  const f = await fixture();
  const draft = await f.engine.prepareDraft({
    operation: "OBSERVE_CROSSING",
    locus_id: f.state.active_locus_id,
    observed_at: 1001,
    payload: { matter_id: "MATTER-003", content_sha256: "6".repeat(64), media_type: "text/plain", claim: "COMMITMENT_ONLY" }
  });
  const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
  const altered = `${signature.slice(0, -2)}${signature.endsWith("00") ? "01" : "00"}`;
  await assert.rejects(() => f.engine.submitTransition({ unsigned: draft.unsigned, signature: altered }), /invalid Ed25519/);
});

test("accepted receipt is checked against the current accepted commitment", async () => {
  const f = await fixture();
  const id = "9".repeat(64);
  f.client.receipt = async () => ({
    schema: "PRESENCE_RECEIPT_001",
    transition_id: id,
    revision: f.state.revision,
    state_commitment: f.state.state_commitment,
    operation: "OBSERVE_CROSSING",
    actor_id: f.keys.actorId,
    locus_id: f.state.active_locus_id,
    effect: "RECORDS_CROSSING_ONLY",
    posture: f.state.body.posture,
    presence_count: 1,
    capacity: {},
    non_effects: []
  });
  assert.equal((await f.engine.waitReceipt(id)).transition_id, id);
});

test("incoherent status and state snapshots fail closed", async () => {
  const f = await fixture();
  f.client.status = async () => statusFor(f.state, { state_commitment: "8".repeat(64) });
  await assert.rejects(() => f.engine.observe(), /incoherent/);
});

import assert from "node:assert/strict";
import { createHash, createPrivateKey, sign } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { MediumEngine } from "../field/src/engine.js";
import { VMClient } from "../field/src/vm-client.js";
import { AppendOnlyJournal } from "../field/src/journal.js";
import { loadProfile } from "../field/src/profile.js";
import { actorIdFromPublicKey, validateConfig } from "../field/src/config.js";
import { buildPassage } from "../field/src/continuity.js";
import { signingBytes } from "../field/src/signature.js";
import {
  MEDIUM_CONFIG_SCHEMA, REQUIRED_EFFECT_CEILING, VM_FORM_ID, VM_FORM_SHA256,
  VM_ID, VM_REQUIRED_MEDIUM_CAPABILITIES, VM_VERSION, VM_RPCCHAINVM_PROTOCOL
} from "../field/src/constants.js";

const [mode, summaryPath, workspace] = process.argv.slice(2);
assert.ok(["enact", "verify"].includes(mode), "enact or verify is required");
const root = dirname(dirname(fileURLToPath(import.meta.url)));
const summary = JSON.parse(await readFile(summaryPath, "utf8"));
assert.equal(summary.nodes.length, 3);
assert.equal(summary.vm_id, VM_ID);
assert.notEqual(summary.network_id, 1);
assert.notEqual(summary.network_id, 5);
const endpoints = summary.nodes.map(({ uri }) => {
  const url = new URL(uri);
  assert.equal(url.hostname, "127.0.0.1", "no public-network endpoint is permitted");
  assert.equal(url.protocol, "http:");
  return `${url.origin}/ext/bc/${summary.blockchain_id}`;
});
const options = { request_timeout_ms: 10000, max_response_bytes: 8 * 1024 * 1024 };
const clients = endpoints.map(endpoint => new VMClient({ endpoint, ...options }));
const resultPath = join(workspace, "result.json");
const hash = text => createHash("sha256").update(text).digest("hex");

async function agree(expected) {
  for (const client of clients) {
    const deadline = Date.now() + 60000;
    let state;
    do {
      state = await client.state();
      if (state.state_commitment === expected.state_commitment) break;
      await new Promise(resolve => setTimeout(resolve, 250));
    } while (Date.now() < deadline);
    assert.equal(state.state_commitment, expected.state_commitment);
    assert.equal(state.revision, expected.revision);
  }
}

if (mode === "verify") {
  const result = JSON.parse(await readFile(resultPath, "utf8"));
  await agree(result.final_state);
  for (const client of clients) {
    for (const expected of result.receipts) {
      const receipt = await client.receipt(expected.transition_id);
      assert.equal(receipt.state_commitment, expected.state_commitment);
    }
  }
  console.log("all three validators retain accepted state and receipts");
} else {
  async function signer(directory) {
    const authority = JSON.parse(await readFile(join(workspace, `${directory}/locality-authority.private.json`), "utf8"));
    const secret = Buffer.from(authority.private_key, "hex");
    const der = Buffer.concat([Buffer.from("302e020100300506032b657004220420", "hex"), secret.subarray(0, 32)]);
    const key = createPrivateKey({ key: der, format: "der", type: "pkcs8" });
    secret.fill(0);
    der.fill(0);
    delete authority.private_key;
    return { publicKey: authority.public_key, sign: unsigned => sign(null, signingBytes(unsigned), key).toString("hex") };
  }
  async function field(name, actorID, keyMaterial, profileName) {
    const config = validateConfig({
      ...options,
      schema: MEDIUM_CONFIG_SCHEMA, medium_id: `LOCAL-NETWORK-${name}`,
      endpoint: endpoints[0], expected_blockchain_id: summary.blockchain_id,
      expected_locality_id: "PRESENCE-LOCAL-QUALIFICATION",
      expected_vm_id: VM_ID, expected_vm_version: VM_VERSION,
      expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
      actor_id: actorID, actor_public_key: keyMaterial.publicKey,
      profile_path: join(root, "field/profiles", profileName),
      journal_path: join(workspace, `${name.toLowerCase()}-field.ndjson`), receipt_timeout_ms: 120000,
    });
    const journal = await new AppendOnlyJournal(config.journal_path).initialize();
    const profile = await loadProfile(config.profile_path);
    return { engine: new MediumEngine({ config, profile, journal, client: clients[0] }), keyMaterial };
  }
  const hostKey = await signer("authority");
  const participantKey = await signer("participant");
  const host = await field("HOST", "PRESENCE-LOCAL-QUALIFICATION", hostKey, "LOCALITY_FIELD_001.json");
  const participantID = actorIdFromPublicKey(participantKey.publicKey);
  const participant = await field("PARTICIPANT", participantID, participantKey, "AI_FIELD_001.json");
  const receipts = [];
  async function enact(actor, operation, locus_id, payload, observed_at = Math.floor(Date.now() / 1000)) {
    const before = await actor.engine.observe();
    const request = { operation, locus_id, payload, observed_at: Math.floor(Date.now() / 1000) };
    request.observed_at = observed_at;
    const draft = await actor.engine.prepareDraft(request);
    const transition = { unsigned: draft.unsigned, signature: actor.keyMaterial.sign(draft.unsigned) };
    const submitted = await actor.engine.submitTransition(transition);
    const receipt = await actor.engine.waitReceipt(submitted.transition_id);
    const renewed = await actor.engine.observe();
    assert.equal(receipt.revision, before.state.revision + 1);
    assert.equal(renewed.state.state_commitment, receipt.state_commitment);
    receipts.push(receipt);
    await agree(renewed.state);
    return { state: renewed.state, id: receipt.transition_id };
  }
  const first = "PRESENCE-LOCAL-LOCUS-001";
  const second = "PRESENCE-LOCAL-LOCUS-002";
  const boundPayload = locus => ({
    locus_id: locus, purpose_sha256: hash("local test"), closure_condition_sha256: hash("close after test"),
    capacity_ceiling_units: 500
  });
  const capacity = { actual_units: 900, resource_commitment_sha256: hash("local capacity"), basis_sha256: hash("test only") };
  await enact(host, "BOUND", first, boundPayload(first));
  await assert.rejects(() => host.engine.prepareDraft({
    operation: "DECLARE_CAPACITY", locus_id: first, payload: capacity, observed_at: Math.floor(Date.now() / 1000),
  }), /scope|locus/i);
  await enact(host, "DECLARE_CAPACITY", "", capacity);
  const pulseAt = Math.floor(Date.now() / 1000);
  const carrier = hash("local network carrier");
  const currentness = await clients[0].request(`/currentness?observed_at=${pulseAt}&carrier_set_sha256=${carrier}`);
  await enact(host, "PULSE", "", {
    currentness_commitment_sha256: currentness.next_pulse.currentness_commitment_sha256,
    carrier_set_sha256: carrier
  }, pulseAt);
  const initialState = hash("participant initial state");
  const presentation = state_commitment => ({
    locality_reference: "PRESENCE-LOCAL-QUALIFICATION", source_reference: "LOCAL-NETWORK-QUALIFICATION",
    source_sha256: hash("participant source"), nucleus_version: "1.0.0",
    nucleus_sha256: hash("participant nucleus"), form_id: VM_FORM_ID, form_sha256: VM_FORM_SHA256,
    state_commitment, medium_capabilities: [...VM_REQUIRED_MEDIUM_CAPABILITIES],
    effect_ceiling: [...REQUIRED_EFFECT_CEILING]
  });
  const presented = await enact(participant, "PRESENT_FORM", first, presentation(initialState));
  const gate = presentationID => ({
    participant_id: participantID, presentation_id: presentationID, disposition: "ADMIT",
    reason_sha256: hash("bounded admission"), work_units: 10, resolution_units: 2,
    offer_expires_at: Math.floor(Date.now() / 1000) + 600
  });
  await enact(host, "GATE_DISPOSITION", first, gate(presented.id));
  const entered = await enact(participant, "ENTER", first, { presentation_id: presented.id });
  const crossing = await enact(participant, "OBSERVE_CROSSING", first, {
    matter_id: "NETWORK-MATTER-001", content_sha256: hash("matter"),
    media_type: "application/octet-stream", claim: "bounded network crossing"
  });
  await enact(participant, "MATTER_DISPOSITION", first, {
    matter_id: "NETWORK-MATTER-001", disposition: "ADMIT", reason_sha256: hash("matter admission")
  });
  await enact(host, "RECORD_EMERGENCE", first, {
    emergence_id: "NETWORK-EMERGENCE-001", contributor_ids: [participantID],
    matter_ids: ["NETWORK-MATTER-001"], kind: "LOCAL_TEST", description_sha256: hash("network emergence")
  });
  await enact(participant, "CORRECT", first, {
    target_transition_id: crossing.id, replacement_commitment: hash("corrected"),
    reason_sha256: hash("append-only correction")
  });
  const departure = buildPassage(participantID, initialState, [hash("departure delta")]);
  const checkpoint = await enact(participant, "CHECKPOINT_DEPARTURE", first, {
    entry_transition_id: entered.id, presentation_id: presented.id,
    from_state_commitment: initialState, passage: departure.passage,
    departure_state_commitment: departure.result_state_commitment
  });
  await enact(participant, "EXIT", first, {
    departure_checkpoint_id: checkpoint.id, reason_sha256: hash("network exit")
  });
  const closed = await enact(host, "CLOSE", first, { closure_basis_sha256: hash("first lifecycle complete") });
  const residue = closed.state.loci[first].residues[participantID];
  assert.ok(residue);
  await enact(host, "BOUND", second, boundPayload(second));
  const returned = buildPassage(participantID, departure.result_state_commitment, [hash("absence delta")]);
  const presentedAgain = await enact(participant, "PRESENT_FORM", second, presentation(returned.result_state_commitment));
  await enact(host, "GATE_DISPOSITION", second, gate(presentedAgain.id));
  await enact(participant, "REENTER", second, {
    presentation_id: presentedAgain.id, prior_entry_transition_id: entered.id,
    departure_checkpoint_id: checkpoint.id, residue_id: residue.residue_id, passage: returned.passage
  });
  const final = await enact(host, "INCORPORATE_ADDRESSED_RESIDUE", "", residue);
  assert.ok(final.state.incorporated_residues[residue.residue_id]);
  const result = {
    schema: "PRESENCE_AVALANCHE_LOCAL_NETWORK_QUALIFICATION_001",
    status: "PASSED_AT_DECLARED_LOCAL_SCOPE",
    validator_count: 3, network_id: summary.network_id, vm_id: VM_ID,
    operations: receipts.map(r => r.operation), receipts,
    final_state: { revision: final.state.revision, state_commitment: final.state.state_commitment },
    renewed_reads_and_three_validator_agreement: true,
    public_network_act: false,
    effect: "DISPOSABLE_IMPLEMENTATION_TEST_NOT_DEPLOYMENT_OR_CONSTITUTIONAL_EFFECT",
  };
  await writeFile(resultPath, `${JSON.stringify(result, null, 2)}\n`, { mode: 0o600 });
  console.log("complete accepted lifecycle accepted by three local validators");
}

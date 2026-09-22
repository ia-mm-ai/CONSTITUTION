import assert from "node:assert/strict";
import { createHash, createPrivateKey, generateKeyPairSync, sign } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { MediumEngine } from "../field/src/engine.js";
import { VMClient } from "../field/src/vm-client.js";
import { AppendOnlyJournal } from "../field/src/journal.js";
import { loadProfile } from "../field/src/profile.js";
import { validateConfig, actorIdFromPublicKey } from "../field/src/config.js";
import { buildPassage } from "../field/src/continuity.js";
import { signingBytes } from "../field/src/signature.js";
import {
  MEDIUM_CONFIG_SCHEMA, REQUIRED_EFFECT_CEILING, VM_FORM_ID, VM_FORM_SHA256,
  VM_ID, VM_REQUIRED_MEDIUM_CAPABILITIES, VM_RPCCHAINVM_PROTOCOL, VM_VERSION
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
const localityID = "PRESENCE-LOCAL-QUALIFICATION";

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
  const authority = JSON.parse(await readFile(join(workspace, "authority/locality-authority.private.json"), "utf8"));
  const secret = Buffer.from(authority.private_key, "hex");
  const der = Buffer.concat([Buffer.from("302e020100300506032b657004220420", "hex"), secret.subarray(0, 32)]);
  const hostKey = createPrivateKey({ key: der, format: "der", type: "pkcs8" });
  secret.fill(0);
  der.fill(0);
  delete authority.private_key;
  const participantPair = generateKeyPairSync("ed25519");
  const participantPublicKey = participantPair.publicKey
    .export({ format: "der", type: "spki" }).subarray(-32).toString("hex");
  const participantID = actorIdFromPublicKey(participantPublicKey, VM_ID);

  async function field(name, actorID, publicKey, profileName, key) {
    const config = validateConfig({
      ...options,
      schema: MEDIUM_CONFIG_SCHEMA, medium_id: `LOCAL-NETWORK-${name.toUpperCase()}`,
      endpoint: endpoints[0], expected_blockchain_id: summary.blockchain_id,
      expected_locality_id: localityID,
      expected_vm_id: VM_ID, expected_vm_version: VM_VERSION,
      expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
      actor_id: actorID, actor_public_key: publicKey,
      profile_path: join(root, "field/profiles", profileName),
      journal_path: join(workspace, `${name}.ndjson`), receipt_timeout_ms: 120000,
    });
    const journal = await new AppendOnlyJournal(config.journal_path).initialize();
    const profile = await loadProfile(config.profile_path);
    const signer = unsigned => ({ unsigned, signature: sign(null, signingBytes(unsigned), key).toString("hex") });
    return { config, signer, engine: new MediumEngine({ config, profile, journal, client: clients[0] }) };
  }

  const host = await field("host", localityID, authority.public_key, "LOCALITY_FIELD_001.json", hostKey);
  const participant = await field("participant", participantID, participantPublicKey, "AI_FIELD_001.json", participantPair.privateKey);
  const receipts = [];
  async function enact(actor, operation, locus_id, payload, observedAt = Math.floor(Date.now() / 1000)) {
    const before = await actor.engine.observe();
    const draft = await actor.engine.prepareDraft({ operation, locus_id, payload, observed_at: observedAt });
    const submitted = await actor.engine.submitTransition(actor.signer(draft.unsigned));
    const receipt = await actor.engine.waitReceipt(submitted.transition_id);
    const renewed = await actor.engine.observe();
    assert.equal(receipt.operation, operation);
    assert.equal(receipt.revision, before.state.revision + 1);
    assert.equal(renewed.state.state_commitment, receipt.state_commitment);
    receipts.push(receipt);
    await agree(renewed.state);
    return { id: submitted.transition_id, receipt, state: renewed.state };
  }

  const first = "PRESENCE-LOCAL-LOCUS-001";
  const second = "PRESENCE-LOCAL-LOCUS-002";
  const boundPayload = id => ({
    locus_id: id, purpose_sha256: hash(`${id} purpose`),
    closure_condition_sha256: hash(`${id} closure`), capacity_ceiling_units: 500,
  });
  const capacity = { actual_units: 900, resource_commitment_sha256: hash("local capacity"), basis_sha256: hash("test only") };
  const presentationPayload = (stateCommitment, version) => ({
    locality_reference: "LOCAL-NETWORK-PARTICIPANT-LOCALITY", source_reference: "INTEGRATION-LOCAL-ONLY",
    source_sha256: hash("participant source"), nucleus_version: version,
    nucleus_sha256: hash(`participant nucleus ${version}`), form_id: VM_FORM_ID, form_sha256: VM_FORM_SHA256,
    state_commitment: stateCommitment, medium_capabilities: [...VM_REQUIRED_MEDIUM_CAPABILITIES],
    effect_ceiling: [...REQUIRED_EFFECT_CEILING]
  });
  const admissionPayload = presentationID => ({
    participant_id: participantID, presentation_id: presentationID, disposition: "ADMIT",
    reason_sha256: hash("bounded admission"), work_units: 10, resolution_units: 2,
    offer_expires_at: Math.floor(Date.now() / 1000) + 600
  });

  // 1 BOUND
  await enact(host, "BOUND", first, boundPayload(first));
  // wrong-scope rejection: locality capacity must not impersonate the locus
  await assert.rejects(() => host.engine.prepareDraft({
    operation: "DECLARE_CAPACITY", locus_id: first, payload: capacity, observed_at: Math.floor(Date.now() / 1000),
  }), /scope|locus/i);
  // 2 DECLARE_CAPACITY (body-local while a locus is active)
  const active = await enact(host, "DECLARE_CAPACITY", "", capacity);
  assert.equal(active.state.active_locus_id, first);
  // 3 PULSE
  const pulseObservedAt = Math.floor(Date.now() / 1000);
  const carrier = hash("local network carrier set");
  const current = await clients[0].request(`/currentness?observed_at=${pulseObservedAt}&carrier_set_sha256=${carrier}`);
  await enact(host, "PULSE", "", {
    currentness_commitment_sha256: current.next_pulse.currentness_commitment_sha256,
    carrier_set_sha256: carrier
  }, pulseObservedAt);
  // 4 PRESENT_FORM
  const initialParticipantState = hash("network participant state commitment");
  const presented = await enact(participant, "PRESENT_FORM", first, presentationPayload(initialParticipantState, "1"));
  // 5 GATE_DISPOSITION
  await enact(host, "GATE_DISPOSITION", first, admissionPayload(presented.id));
  // 6 ENTER
  const entered = await enact(participant, "ENTER", first, { presentation_id: presented.id });
  assert.equal(entered.state.body.presence_count, 1);
  // 7 OBSERVE_CROSSING
  const crossing = await enact(participant, "OBSERVE_CROSSING", first, {
    matter_id: "NETWORK-MATTER-001", content_sha256: hash("network matter content"),
    media_type: "application/octet-stream", claim: "bounded network claim"
  });
  // 8 MATTER_DISPOSITION
  await enact(participant, "MATTER_DISPOSITION", first, {
    matter_id: "NETWORK-MATTER-001", disposition: "ADMIT", reason_sha256: hash("local matter admission")
  });
  // 9 RECORD_EMERGENCE
  await enact(host, "RECORD_EMERGENCE", first, {
    emergence_id: "NETWORK-EMERGENCE-001", contributor_ids: [participantID],
    matter_ids: ["NETWORK-MATTER-001"], kind: "LOCAL_TEST", description_sha256: hash("bounded emergence")
  });
  // 10 CORRECT (target-scoped to the exact crossing transition)
  await enact(participant, "CORRECT", first, {
    target_transition_id: crossing.id, replacement_commitment: hash("corrected commitment"),
    reason_sha256: hash("append-only correction")
  });
  // 11 CHECKPOINT_DEPARTURE
  const departure = buildPassage(participantID, initialParticipantState, [hash("departure delta")]);
  const checkpoint = await enact(participant, "CHECKPOINT_DEPARTURE", first, {
    entry_transition_id: entered.id, presentation_id: presented.id, from_state_commitment: initialParticipantState,
    passage: departure.passage, departure_state_commitment: departure.result_state_commitment
  });
  // 12 EXIT
  await enact(participant, "EXIT", first, {
    departure_checkpoint_id: checkpoint.id, reason_sha256: hash("lawful exit")
  });
  // 13 CLOSE
  const closed = await enact(host, "CLOSE", first, { closure_basis_sha256: hash("completed network lifecycle") });
  assert.equal(closed.state.active_locus_id, "");
  const residue = closed.state.loci[first].residues[participantID];
  assert.equal(residue.departure_checkpoint_id, checkpoint.id);
  // 14 BOUND (return locus)
  await enact(host, "BOUND", second, boundPayload(second));
  // 15 PRESENT_FORM (renewed)
  const returned = buildPassage(participantID, departure.result_state_commitment, [hash("delta while absent")]);
  const presentedAgain = await enact(participant, "PRESENT_FORM", second, presentationPayload(returned.result_state_commitment, "2"));
  // 16 GATE_DISPOSITION (renewed admission)
  await enact(host, "GATE_DISPOSITION", second, admissionPayload(presentedAgain.id));
  // 17 REENTER (carried state, consumes checkpoint, addresses exact residue)
  const reentered = await enact(participant, "REENTER", second, {
    presentation_id: presentedAgain.id, prior_entry_transition_id: entered.id,
    departure_checkpoint_id: checkpoint.id, residue_id: residue.residue_id, passage: returned.passage
  });
  assert.equal(reentered.state.departure_checkpoints[checkpoint.id].status, "CONSUMED");
  assert.equal(reentered.state.entry_history[reentered.id].ingress_mode, "RENEWED_ENTRY");
  // Complete the renewed presence lawfully. The LOCALITY_FIELD_001 reference
  // profile permits INCORPORATE_ADDRESSED_RESIDUE only with no active locus,
  // so the second locus is checkpointed, exited and closed first; the exact
  // eighteen-operation ordering at VM law level is separately proven by
  // vm/lifecycle_regression_test.go.
  const secondCheckpoint = await enact(participant, "CHECKPOINT_DEPARTURE", second, {
    entry_transition_id: reentered.id, presentation_id: presentedAgain.id,
    from_state_commitment: returned.result_state_commitment, passage: [],
    departure_state_commitment: returned.result_state_commitment
  });
  await enact(participant, "EXIT", second, {
    departure_checkpoint_id: secondCheckpoint.id, reason_sha256: hash("final lawful exit")
  });
  await enact(host, "CLOSE", second, { closure_basis_sha256: hash("second completed lifecycle") });
  // 18th distinct operation: INCORPORATE_ADDRESSED_RESIDUE (explicit addressed
  // incorporation with no active locus; envelope bytes and commitments exactly
  // match the Go residueCommitments)
  const envelope = {
    schema: "PRESENCE_AVALANCHE_ADDRESSED_RESIDUE_001",
    addressed_to_locality_id: localityID,
    source_locality_id: "EXTERNAL-LOCALITY-NETWORK-001",
    source_locus_id: "EXTERNAL-LOCUS-NETWORK-001",
    source_state_commitment: hash("external network source state"),
    closure_transition_id: hash("external network closure"),
    participant_id: "EXTERNAL-PARTICIPANT-NETWORK-001",
    presentation_id: hash("external network presentation"),
    entry_transition_id: hash("external network entry"),
    departure_checkpoint_id: hash("external network departure checkpoint"),
    departure_state_commitment: hash("external network departure state"),
    effect: "AVAILABLE_FOR_ADDRESSEE_DECISION_ONLY",
  };
  const envelopeBytes = Buffer.from(JSON.stringify(envelope), "utf8");
  const envelopeSHA = createHash("sha256").update(envelopeBytes).digest("hex");
  const envelopeID = createHash("sha256")
    .update(Buffer.concat([Buffer.from("PRESENCE_AVALANCHE_RESIDUE_ID_001\x00", "utf8"), envelopeBytes]))
    .digest("hex");
  const finalState = (await enact(host, "INCORPORATE_ADDRESSED_RESIDUE", "", {
    residue_id: envelopeID, residue_sha256: envelopeSHA, schema: envelope.schema,
    addressed_to_locality_id: envelope.addressed_to_locality_id, source_locality_id: envelope.source_locality_id,
    source_locus_id: envelope.source_locus_id, source_state_commitment: envelope.source_state_commitment,
    closure_transition_id: envelope.closure_transition_id, participant_id: envelope.participant_id,
    presentation_id: envelope.presentation_id, entry_transition_id: envelope.entry_transition_id,
    departure_checkpoint_id: envelope.departure_checkpoint_id, departure_state_commitment: envelope.departure_state_commitment,
    effect: envelope.effect,
  })).state;
  assert.ok(finalState.incorporated_residues[envelopeID]);
  assert.equal(finalState.revision, 21);
  const result = {
    schema: "PRESENCE_AVALANCHE_LOCAL_NETWORK_QUALIFICATION_001",
    status: "PASSED_AT_DECLARED_LOCAL_SCOPE",
    validator_count: 3, network_id: summary.network_id, vm_id: VM_ID,
    operations: receipts.map(r => r.operation), receipts,
    final_state: { revision: finalState.revision, state_commitment: finalState.state_commitment },
    renewed_reads_and_three_validator_agreement: true,
    public_network_act: false,
    effect: "DISPOSABLE_IMPLEMENTATION_TEST_NOT_DEPLOYMENT_OR_CONSTITUTIONAL_EFFECT",
  };
  await writeFile(resultPath, `${JSON.stringify(result, null, 2)}\n`, { mode: 0o600 });
  console.log("complete lifecycle including all 18 accepted operations accepted by three local validators");
}

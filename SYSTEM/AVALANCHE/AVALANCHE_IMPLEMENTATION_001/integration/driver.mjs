import assert from "node:assert/strict";
import { createHash, createPrivateKey, sign } from "node:crypto";
import { readFile, rm } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { MediumEngine } from "../field/src/engine.js";
import { VMClient } from "../field/src/vm-client.js";
import { AppendOnlyJournal } from "../field/src/journal.js";
import { CapabilityBroker } from "../field/src/broker.js";
import { loadProfile } from "../field/src/profile.js";
import { validateConfig } from "../field/src/config.js";
import { buildPassage } from "../field/src/continuity.js";
import { signingBytes } from "../field/src/signature.js";
import {
  MEDIUM_CONFIG_SCHEMA, REQUIRED_EFFECT_CEILING, VM_FORM_ID, VM_FORM_SHA256,
  VM_ID, VM_REQUIRED_MEDIUM_CAPABILITIES, VM_RPCCHAINVM_PROTOCOL, VM_VERSION
} from "../field/src/constants.js";

const directory = dirname(fileURLToPath(import.meta.url));
const chunks = [];
for await (const chunk of process.stdin) chunks.push(chunk);
const inputBytes = Buffer.concat(chunks);
const input = JSON.parse(inputBytes.toString("utf8"));
inputBytes.fill(0);
for (const chunk of chunks) chunk.fill(0);
const publicURL = new URL(input.endpoint);
assert.equal(publicURL.hostname, "127.0.0.1", "integration must not write to a public network");
assert.equal(new URL(input.control_endpoint).origin, publicURL.origin);
assert.match(input.journal_name, /^\.coupled-journal-[0-9a-f]{32}$/);
const workspace = join(directory, input.journal_name);
const hash = (text) => createHash("sha256").update(text).digest("hex");
const records = [];
const checks = [];
const seeds = [input.host_seed, input.participant_seed];

function externalSigner(seedHex) {
  const seed = Buffer.from(seedHex, "hex");
  const der = Buffer.concat([Buffer.from("302e020100300506032b657004220420", "hex"), seed]);
  const key = createPrivateKey({ key: der, format: "der", type: "pkcs8" });
  seed.fill(0);
  der.fill(0);
  return (unsigned) => ({ unsigned, signature: sign(null, signingBytes(unsigned), key).toString("hex") });
}

async function field(name, actorID, publicKey, profileName, signer) {
  const config = validateConfig({
    schema: MEDIUM_CONFIG_SCHEMA,
    medium_id: `COUPLED-${name.toUpperCase()}`,
    endpoint: input.endpoint,
    expected_blockchain_id: "LOCAL-COUPLED",
    expected_locality_id: input.host_id,
    expected_vm_id: VM_ID,
    expected_vm_version: VM_VERSION,
    expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
    actor_id: actorID,
    actor_public_key: publicKey,
    profile_path: resolve(directory, "../field/profiles", profileName),
    journal_path: join(workspace, `${name}.ndjson`),
    request_timeout_ms: 5000,
    receipt_timeout_ms: 1000,
    max_response_bytes: 8 * 1024 * 1024
  });
  const profile = await loadProfile(config.profile_path);
  const journal = await new AppendOnlyJournal(config.journal_path).initialize();
  const client = new VMClient(config);
  return { config, journal, client, signer, engine: new MediumEngine({ config, profile, client, journal }) };
}

async function control(action) {
  const response = await fetch(`${input.control_endpoint}/__test/${action}`, {
    method: "POST", headers: { "X-Test-Control": input.control_token },
    signal: AbortSignal.timeout(10000), redirect: "error"
  });
  const body = await response.json();
  assert.equal(response.status, 200, `real VM ${action}: ${JSON.stringify(body)}`);
  return body;
}

function request(operation, locusID, payload, observedAt = Math.floor(Date.now() / 1000)) {
  return { operation, locus_id: locusID, observed_at: observedAt, payload };
}

async function enact(actor, operation, locusID, payload, observedAt) {
  const before = await actor.engine.observe();
  const draft = await actor.engine.prepareDraft(request(operation, locusID, payload, observedAt));
  assert.equal(draft.effect, "DRAFT_ONLY_NO_STATE_CHANGE");
  assert.equal(draft.unsigned.locus_id, locusID);
  assert.equal((await actor.engine.observe()).state.state_commitment, before.state.state_commitment);
  const signed = actor.signer(draft.unsigned);
  if (records.length === 0) {
    const invalid = { ...signed, signature: "00".repeat(64) };
    await assert.rejects(() => actor.engine.submitTransition(invalid), /invalid Ed25519/i);
    await assert.rejects(() => actor.client.submit(invalid), /signature/i);
    checks.push("invalid_signature_rejected_by_FIELD_and_VM");
  }
  const submitted = await actor.engine.submitTransition(signed);
  assert.equal(submitted.status, "PENDING_CONSENSUS");
  assert.equal(submitted.effect, "NO_ACCEPTANCE_UNTIL_BLOCK_ACCEPTED");
  assert.equal(await actor.client.receipt(submitted.transition_id), null);
  assert.equal((await actor.engine.observe()).state.state_commitment, before.state.state_commitment);
  assert.equal((await actor.engine.decide(operation)).allowed, false);
  assert.equal((await actor.engine.decide("READ_PUBLIC_STATE")).allowed, true);
  const accepted = await control("accept");
  const receipt = await actor.engine.waitReceipt(submitted.transition_id);
  const renewed = await actor.engine.observe();
  assert.equal(receipt.operation, operation);
  assert.equal(receipt.locus_id, locusID);
  assert.equal(receipt.revision, before.state.revision + 1);
  assert.equal(receipt.state_commitment, renewed.state.state_commitment);
  assert.equal(renewed.observation.accepted_revision, receipt.revision);
  assert.equal(renewed.status.pending_transitions, 0);
  assert.notEqual(receipt.state_commitment, before.state.state_commitment);
  assert.equal((await actor.client.transition(submitted.transition_id)).status, "ACCEPTED");
  records.push({
    operation, locus_id: locusID, revision: receipt.revision,
    transition_id: receipt.transition_id, state_commitment: receipt.state_commitment,
    accepted_block_id: accepted.accepted_block_id,
    receipt_then_renewed_read: true, draft_and_submission_did_not_accept: true
  });
  return { id: submitted.transition_id, receipt, state: renewed.state };
}

async function restart(actor) {
  const before = await actor.engine.observe();
  const response = await control("restart");
  const after = await actor.engine.observe();
  assert.equal(response.previous_last_accepted, response.last_accepted);
  assert.deepEqual(after.state, before.state);
  for (const record of records) {
    const receipt = await actor.client.receipt(record.transition_id);
    assert.equal(receipt.state_commitment, record.state_commitment);
    assert.equal((await actor.client.transition(record.transition_id)).status, "ACCEPTED");
  }
}

try {
  const host = await field("host", input.host_id, input.host_public_key, "LOCALITY_FIELD_001.json", externalSigner(input.host_seed));
  const participant = await field("participant", input.participant_id, input.participant_public_key, "AI_FIELD_001.json", externalSigner(input.participant_seed));
  delete input.host_seed;
  delete input.participant_seed;
  const broker = new CapabilityBroker(participant.engine).register("AI_BOUNDED_TOOL_BROKER", async () => "BOUNDED_LOCAL_EFFECT");
  await assert.rejects(() => broker.invoke("AI_BOUNDED_TOOL_BROKER"), /denied/);
  const initial = await host.engine.observe();
  assert.equal(initial.state.revision, 0);
  assert.equal(initial.status.rpcchainvm_protocol, 46);
  assert.equal(initial.status.avalanchego_target, "v1.15.0");

  const first = "COUPLED-LOCUS-001";
  const second = "COUPLED-LOCUS-002";
  const boundPayload = (id) => ({
    locus_id: id, purpose_sha256: hash(`${id} purpose`),
    closure_condition_sha256: hash(`${id} closure`), capacity_ceiling_units: 100
  });
  const capacityPayload = (units) => ({
    actual_units: units, resource_commitment_sha256: hash(`capacity ${units}`),
    basis_sha256: hash("signed local account, not external resource proof")
  });
  const bound = await enact(host, "BOUND", first, boundPayload(first));
  assert.equal(bound.state.active_locus_id, first);

  await assert.rejects(
    () => host.engine.prepareDraft(request("DECLARE_CAPACITY", first, capacityPayload(90))),
    /scope|locus/i
  );
  await assert.rejects(
    () => host.client.draft({
      ...request("DECLARE_CAPACITY", first, capacityPayload(90)),
      actor_id: input.host_id, actor_public_key: input.host_public_key
    }), /scope|locus/i
  );
  const stale = await host.engine.prepareDraft(request("DECLARE_CAPACITY", "", capacityPayload(89)));
  const capacity = await enact(host, "DECLARE_CAPACITY", "", capacityPayload(90));
  assert.equal(capacity.state.active_locus_id, first);
  assert.equal(capacity.state.capacity.declared_actual_units, 90);
  await assert.rejects(() => host.engine.submitTransition(host.signer(stale.unsigned)), /STALE/);
  checks.push("BOUND_then_empty_locus_DECLARE_CAPACITY_while_active", "wrong_global_scope_rejected_by_FIELD_and_VM", "stale_draft_rejected");
  await restart(host);
  checks.push("VM_reload_while_locus_active_preserves_state_and_receipts");

  const initialParticipantState = hash("private participant state commitment only");
  const presentationPayload = (stateCommitment, version) => ({
    locality_reference: "COUPLED-PARTICIPANT-LOCALITY", source_reference: "INTEGRATION-LOCAL-ONLY",
    source_sha256: hash("participant source"), nucleus_version: version,
    nucleus_sha256: hash(`participant nucleus ${version}`), form_id: VM_FORM_ID, form_sha256: VM_FORM_SHA256,
    state_commitment: stateCommitment, medium_capabilities: [...VM_REQUIRED_MEDIUM_CAPABILITIES],
    effect_ceiling: [...REQUIRED_EFFECT_CEILING]
  });
  const admissionPayload = (presentationID) => ({
    participant_id: input.participant_id, presentation_id: presentationID, disposition: "ADMIT",
    reason_sha256: hash("bounded admission"), work_units: 10, resolution_units: 2,
    offer_expires_at: Math.floor(Date.now() / 1000) + 600
  });
  const presented = await enact(participant, "PRESENT_FORM", first, presentationPayload(initialParticipantState, "1"));
  assert.equal(presented.state.loci[first].entries[input.participant_id], undefined);
  await assert.rejects(() => participant.engine.prepareDraft(request("ENTER", first, { presentation_id: presented.id })), /denied/);
  const admitted = await enact(host, "GATE_DISPOSITION", first, admissionPayload(presented.id));
  assert.equal(admitted.state.loci[first].entries[input.participant_id], undefined);
  const entered = await enact(participant, "ENTER", first, { presentation_id: presented.id });
  assert.equal(entered.state.body.presence_count, 1);
  assert.equal(await broker.invoke("AI_BOUNDED_TOOL_BROKER"), "BOUNDED_LOCAL_EFFECT");
  checks.push("presentation_and_admission_do_not_enter", "broker_requires_current_accepted_presence");

  const crossing = await enact(participant, "OBSERVE_CROSSING", first, {
    matter_id: "COUPLED-MATTER-001", content_sha256: hash("content commitment"),
    media_type: "application/octet-stream", claim: "A&B <bounded> \u2028line\u2029paragraph"
  });
  assert.equal(Object.keys(crossing.state.loci[first].matter["COUPLED-MATTER-001"].dispositions).length, 0);
  await enact(participant, "MATTER_DISPOSITION", first, {
    matter_id: "COUPLED-MATTER-001", disposition: "ADMIT", reason_sha256: hash("local matter admission")
  });
  await enact(host, "RECORD_EMERGENCE", first, {
    emergence_id: "COUPLED-EMERGENCE-001", contributor_ids: [input.participant_id],
    matter_ids: ["COUPLED-MATTER-001"], kind: "LOCAL_TEST", description_sha256: hash("bounded emergence")
  });
  const deficit = await enact(host, "DECLARE_CAPACITY", "", capacityPayload(1));
  assert.equal(deficit.state.body.posture, "HOLD_CAPACITY_DEFICIT");
  await assert.rejects(() => broker.invoke("AI_BOUNDED_TOOL_BROKER"), /denied/);
  await enact(participant, "CORRECT", first, {
    target_transition_id: crossing.id, replacement_commitment: hash("corrected commitment"), reason_sha256: hash("append-only correction")
  });
  const departure = buildPassage(input.participant_id, initialParticipantState, [hash("departure delta")]);
  const checkpoint = await enact(participant, "CHECKPOINT_DEPARTURE", first, {
    entry_transition_id: entered.id, presentation_id: presented.id, from_state_commitment: initialParticipantState,
    passage: departure.passage, departure_state_commitment: departure.result_state_commitment
  });
  assert.equal(checkpoint.state.loci[first].entries[input.participant_id].status, "PRESENT");
  assert.equal(checkpoint.state.departure_checkpoints[checkpoint.id].status, "OPEN");
  const exited = await enact(participant, "EXIT", first, {
    departure_checkpoint_id: checkpoint.id, reason_sha256: hash("lawful deficit egress")
  });
  assert.equal(exited.state.departure_checkpoints[checkpoint.id].status, "SEALED");
  assert.equal(exited.state.body.presence_count, 0);
  await assert.rejects(() => broker.invoke("AI_BOUNDED_TOOL_BROKER"), /denied/);
  await enact(host, "DECLARE_CAPACITY", "", capacityPayload(100));
  const sameLocusState = buildPassage(input.participant_id, departure.result_state_commitment, [hash("same open locus absence delta")]);
  const sameLocusPresentation = await enact(participant, "PRESENT_FORM", first, presentationPayload(sameLocusState.result_state_commitment, "2"));
  await enact(host, "GATE_DISPOSITION", first, admissionPayload(sameLocusPresentation.id));
  assert.equal((await participant.engine.observe()).observation.phase, "REENTRY_AVAILABLE");
  await assert.rejects(() => participant.engine.prepareDraft(request("ENTER", first, { presentation_id: sameLocusPresentation.id })), /denied/);
  const sameLocusEntry = await enact(participant, "REENTER", first, {
    presentation_id: sameLocusPresentation.id, prior_entry_transition_id: entered.id,
    departure_checkpoint_id: checkpoint.id, residue_id: "", passage: sameLocusState.passage
  });
  assert.equal(sameLocusEntry.state.loci[first].phase, "OPEN");
  assert.equal(sameLocusEntry.state.departure_checkpoints[checkpoint.id].status, "CONSUMED");
  assert.equal(sameLocusEntry.state.entry_history[entered.id].status, "EXITED");
  assert.equal(sameLocusEntry.state.entry_history[sameLocusEntry.id].ingress_mode, "RENEWED_ENTRY");
  assert.equal(await broker.invoke("AI_BOUNDED_TOOL_BROKER"), "BOUNDED_LOCAL_EFFECT");
  const sameLocusCheckpoint = await enact(participant, "CHECKPOINT_DEPARTURE", first, {
    entry_transition_id: sameLocusEntry.id, presentation_id: sameLocusPresentation.id,
    from_state_commitment: sameLocusState.result_state_commitment, passage: [],
    departure_state_commitment: sameLocusState.result_state_commitment
  });
  await enact(participant, "EXIT", first, {
    departure_checkpoint_id: sameLocusCheckpoint.id, reason_sha256: hash("same open locus lawful exit")
  });
  checks.push("Go_and_JavaScript_signing_ID_and_submission_agree_for_HTML_and_Unicode_separators", "same_open_locus_fresh_admission_overrides_retained_EXITED_entry_for_reentry");
  const closed = await enact(host, "CLOSE", first, { closure_basis_sha256: hash("first completed lifecycle") });
  assert.equal(closed.state.active_locus_id, "");
  assert.equal(closed.state.loci[first].phase, "CLOSED");
  assert.ok(closed.state.events[crossing.id]);
  const residue = closed.state.loci[first].residues[input.participant_id];
  assert.equal(residue.departure_checkpoint_id, sameLocusCheckpoint.id);
  assert.equal(closed.state.incorporated_residues[residue.residue_id], undefined);
  checks.push("crossing_does_not_admit_matter", "HOLD_denies_work_but_preserves_correction_and_egress", "checkpoint_does_not_exit", "closure_does_not_incorporate_residue");

  async function pulse() {
    const observedAt = Math.floor(Date.now() / 1000);
    const carrier = hash("local controlled test carrier");
    const current = await host.client.request(`/currentness?observed_at=${observedAt}&carrier_set_sha256=${carrier}`);
    const result = await enact(host, "PULSE", "", {
      currentness_commitment_sha256: current.next_pulse.currentness_commitment_sha256,
      carrier_set_sha256: carrier
    }, observedAt);
    assert.equal(result.state.body.presence_count, 0);
  }
  await pulse();
  await enact(host, "BOUND", second, boundPayload(second));
  await enact(host, "CORRECT", first, {
    target_transition_id: bound.id, replacement_commitment: hash("historical bound annotation"),
    reason_sha256: hash("historical explicit locus remains distinct from current locus")
  });
  checks.push("historical_CORRECT_uses_existing_target_locus_not_active_locus");
  const returnedState = buildPassage(input.participant_id, sameLocusState.result_state_commitment, [hash("delta while absent")]);
  const presentedAgain = await enact(participant, "PRESENT_FORM", second, presentationPayload(returnedState.result_state_commitment, "3"));
  await enact(host, "GATE_DISPOSITION", second, admissionPayload(presentedAgain.id));
  await assert.rejects(() => participant.engine.prepareDraft(request("ENTER", second, { presentation_id: presentedAgain.id })), /denied/);
  const reentered = await enact(participant, "REENTER", second, {
    presentation_id: presentedAgain.id, prior_entry_transition_id: sameLocusEntry.id,
    departure_checkpoint_id: sameLocusCheckpoint.id, residue_id: residue.residue_id, passage: returnedState.passage
  });
  assert.equal(reentered.state.departure_checkpoints[sameLocusCheckpoint.id].status, "CONSUMED");
  assert.equal(reentered.state.entry_history[reentered.id].ingress_mode, "RENEWED_ENTRY");
  assert.equal(reentered.state.loci[first].phase, "CLOSED");
  assert.equal(reentered.state.loci[first].entries[input.participant_id].status, "EXITED");
  assert.equal(await broker.invoke("AI_BOUNDED_TOOL_BROKER"), "BOUNDED_LOCAL_EFFECT");
  const finalCheckpoint = await enact(participant, "CHECKPOINT_DEPARTURE", second, {
    entry_transition_id: reentered.id, presentation_id: presentedAgain.id,
    from_state_commitment: returnedState.result_state_commitment, passage: [],
    departure_state_commitment: returnedState.result_state_commitment
  });
  await enact(participant, "EXIT", second, { departure_checkpoint_id: finalCheckpoint.id, reason_sha256: hash("final lawful exit") });
  await enact(host, "CLOSE", second, { closure_basis_sha256: hash("second completed lifecycle") });
  await pulse();
  await restart(host);
  await assert.rejects(() => broker.invoke("AI_BOUNDED_TOOL_BROKER"), /denied/);
  checks.push("renewed_entry_consumes_checkpoint_without_restoring_prior_locus", "final_VM_reload_preserves_all_receipts", "historical_receipts_do_not_grant_presence");

  for (const actor of [host, participant]) {
    const text = await readFile(actor.config.journal_path, "utf8");
    for (const seed of seeds) assert.equal(text.includes(seed), false, "signer seed leaked into journal");
    assert.doesNotMatch(text, /"private_key"|"private_state"|"signature"/);
    const reopened = await new AppendOnlyJournal(actor.config.journal_path).initialize();
    assert.equal(reopened.sequence, actor.journal.sequence);
  }
  const final = await host.engine.observe();
  assert.equal(final.state.active_locus_id, "");
  assert.equal(final.state.body.presence_count, 0);
  assert.equal(final.state.body.posture, "DORMANT_P0");
  assert.equal(final.state.revision, records.length);
  process.stdout.write(`${JSON.stringify({
    schema: "PRESENCE_AVALANCHE_COUPLED_QUALIFICATION_001", result: "PASS",
    mode: "REAL_VM_HTTP_AND_REAL_FIELD_EXPLICIT_BLOCK_ACCEPTANCE",
    observed_at: new Date().toISOString(), go_version: input.go_version, node_version: process.version,
    avalanchego: "v1.15.0", avalanchego_commit: "70bd6d063b7343fd2cd8217200aaf77b57f19f68",
    rpcchainvm_protocol: 46, accepted_transitions: records.length,
    final_revision: final.state.revision, final_posture: final.state.body.posture,
    final_state_commitment: final.state.state_commitment, vm_database_reloads: 2,
    validator_network: "NOT_RUN", validator_process_restart: "NOT_RUN",
    private_credentials: "EPHEMERAL_MEMORY_AND_STDIN_ONLY",
    effect: "LOCAL_IMPLEMENTATION_TEST_EVIDENCE_ONLY_NO_CONSTITUTIONAL_EFFECT",
    checks, transitions: records
  })}\n`);
} finally {
  await rm(workspace, { recursive: true, force: true });
}

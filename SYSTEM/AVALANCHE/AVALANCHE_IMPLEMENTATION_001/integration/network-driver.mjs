import assert from "node:assert/strict";
import { createHash, createPrivateKey, sign } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { MediumEngine } from "../field/src/engine.js";
import { VMClient } from "../field/src/vm-client.js";
import { AppendOnlyJournal } from "../field/src/journal.js";
import { loadProfile } from "../field/src/profile.js";
import { validateConfig } from "../field/src/config.js";
import { signingBytes } from "../field/src/signature.js";
import { MEDIUM_CONFIG_SCHEMA, VM_ID, VM_VERSION, VM_RPCCHAINVM_PROTOCOL } from "../field/src/constants.js";

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
  const authority = JSON.parse(await readFile(join(workspace, "authority/locality-authority.private.json"), "utf8"));
  const secret = Buffer.from(authority.private_key, "hex");
  const der = Buffer.concat([Buffer.from("302e020100300506032b657004220420", "hex"), secret.subarray(0, 32)]);
  const key = createPrivateKey({ key: der, format: "der", type: "pkcs8" });
  secret.fill(0);
  der.fill(0);
  delete authority.private_key;
  const config = validateConfig({
    ...options,
    schema: MEDIUM_CONFIG_SCHEMA, medium_id: "LOCAL-NETWORK-QUALIFICATION",
    endpoint: endpoints[0], expected_blockchain_id: summary.blockchain_id,
    expected_locality_id: "PRESENCE-LOCAL-QUALIFICATION",
    expected_vm_id: VM_ID, expected_vm_version: VM_VERSION,
    expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
    actor_id: "PRESENCE-LOCAL-QUALIFICATION", actor_public_key: authority.public_key,
    profile_path: join(root, "field/profiles/LOCALITY_FIELD_001.json"),
    journal_path: join(workspace, "field.ndjson"), receipt_timeout_ms: 120000,
  });
  const journal = await new AppendOnlyJournal(config.journal_path).initialize();
  const profile = await loadProfile(config.profile_path);
  const engine = new MediumEngine({ config, profile, journal, client: clients[0] });
  const receipts = [];
  async function enact(operation, locus_id, payload) {
    const before = await engine.observe();
    const request = { operation, locus_id, payload, observed_at: Math.floor(Date.now() / 1000) };
    const draft = await engine.prepareDraft(request);
    const transition = { unsigned: draft.unsigned, signature: sign(null, signingBytes(draft.unsigned), key).toString("hex") };
    const submitted = await engine.submitTransition(transition);
    const receipt = await engine.waitReceipt(submitted.transition_id);
    const renewed = await engine.observe();
    assert.equal(receipt.revision, before.state.revision + 1);
    assert.equal(renewed.state.state_commitment, receipt.state_commitment);
    receipts.push(receipt);
    await agree(renewed.state);
    return renewed.state;
  }
  const locus = "PRESENCE-LOCAL-LOCUS";
  await enact("BOUND", locus, {
    locus_id: locus, purpose_sha256: hash("local test"), closure_condition_sha256: hash("close after test"),
    capacity_ceiling_units: 500,
  });
  const capacity = { actual_units: 900, resource_commitment_sha256: hash("local capacity"), basis_sha256: hash("test only") };
  await assert.rejects(() => engine.prepareDraft({
    operation: "DECLARE_CAPACITY", locus_id: locus, payload: capacity, observed_at: Math.floor(Date.now() / 1000),
  }), /scope|locus/i);
  const active = await enact("DECLARE_CAPACITY", "", capacity);
  assert.equal(active.active_locus_id, locus);
  const finalState = await enact("CLOSE", locus, { closure_basis_sha256: hash("completed local scope regression") });
  assert.equal(finalState.active_locus_id, "");
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
  console.log("BOUND -> DECLARE_CAPACITY -> renewed read -> CLOSE accepted by three local validators");
}

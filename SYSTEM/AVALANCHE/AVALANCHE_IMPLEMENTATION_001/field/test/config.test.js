import test from "node:test";
import assert from "node:assert/strict";
import { actorIdFromPublicKey, validateConfig } from "../src/config.js";
import { ROOT_VM_ID, VM_ID, VM_RPCCHAINVM_PROTOCOL, VM_VERSION } from "../src/constants.js";

function base() {
  const publicKey = "12".repeat(32);
  return {
    schema: "PRESENCE_AVALANCHE_FIELD_CONFIG_001",
    medium_id: "AI-FIELD-TEST-001",
    endpoint: "http://127.0.0.1:9650/ext/bc/BLOCKCHAIN-001/rpc",
    expected_blockchain_id: "BLOCKCHAIN-001",
    expected_locality_id: "SUCCESSOR-LOCALITY-001",
    expected_vm_id: VM_ID,
    expected_vm_version: VM_VERSION,
    expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
    actor_id: actorIdFromPublicKey(publicKey),
    actor_public_key: publicKey,
    profile_path: "./profile.json",
    journal_path: "./journal.ndjson",
    request_timeout_ms: 5000,
    receipt_timeout_ms: 10000,
    max_response_bytes: 1048576
  };
}

test("config binds chain route, frozen VM and derived actor ID", () => {
  const value = validateConfig(base());
  assert.equal(value.listen_host, "127.0.0.1");
  assert.equal(value.listen_port, 8787);
});

test("config binds each VM ID to its own actor namespace", () => {
  const config = base();
  config.expected_vm_id = ROOT_VM_ID;
  assert.throws(() => validateConfig(config), /actor_id/);
  config.actor_id = actorIdFromPublicKey(config.actor_public_key, ROOT_VM_ID);
  assert.match(validateConfig(config).actor_id, /^LOCALITY-ACTOR-/);
  assert.throws(() => validateConfig({ ...config, expected_vm_id: VM_ID }), /actor_id/);
  for (const id of ["unknown", "toString", undefined]) {
    assert.throws(() => validateConfig({ ...config, expected_vm_id: id }), /expected_vm_id/);
  }
});

test("config refuses non-loopback plaintext and unknown fields", () => {
  const remote = base();
  remote.endpoint = "http://example.com/ext/bc/BLOCKCHAIN-001/rpc";
  assert.throws(() => validateConfig(remote), /HTTPS/);
  const extra = base();
  extra.private_key = "never";
  assert.throws(() => validateConfig(extra), /unknown field/);
});

test("only the exact expected host locality may use a non-derived actor ID", () => {
  const host = base();
  host.actor_id = host.expected_locality_id;
  assert.equal(validateConfig(host).actor_id, host.expected_locality_id);
  assert.throws(() => validateConfig({ ...host, actor_id: "OTHER-HOST" }), /actor_id/);
  assert.throws(() => validateConfig({ ...host, actor_id: "unsafe host", expected_locality_id: "unsafe host" }), /actor_id/);
});

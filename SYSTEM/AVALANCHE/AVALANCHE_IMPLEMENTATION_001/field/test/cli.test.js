import test from "node:test";
import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { resolve } from "node:path";
import { actorIdFromPublicKey } from "../src/config.js";
import { VM_OPERATION_CONTRACT_SHA256 } from "../src/constants.js";

const run = promisify(execFile);
const cli = resolve("bin/presence-avalanche-field.js");

test("CLI reports frozen version identity and derives the VM-compatible actor ID", async () => {
  const version = JSON.parse((await run(process.execPath, [cli, "--version-json"])).stdout);
  assert.equal(version.protocol, "PRESENCE_AVALANCHE_FIELD_001");
  assert.equal(version.implementation, "AVALANCHE_IMPLEMENTATION_001");
  assert.equal(version.version, "1.0.0");
  assert.equal(version.operation_contract_sha256, VM_OPERATION_CONTRACT_SHA256);
  const publicKey = "42".repeat(32);
  const actor = JSON.parse((await run(process.execPath, [cli, "actor-id", "--public-key", publicKey])).stdout);
  assert.equal(actor.actor_id, actorIdFromPublicKey(publicKey));
  assert.match(actor.actor_id, /^PRESENCE-AVALANCHE-ACTOR-/);
  await assert.rejects(
    () => run(process.execPath, [cli, "actor-id", "--public-key", publicKey, "--vm-id", "NOT-THE-CANONICAL-VM-ID"]),
    /unsupported expected_vm_id/
  );
});

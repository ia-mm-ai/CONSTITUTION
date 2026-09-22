import test from "node:test";
import assert from "node:assert/strict";
import { createServer } from "node:http";
import { once } from "node:events";
import { sign } from "node:crypto";
import { VMClient, VMHTTPError } from "../src/vm-client.js";
import { signingBytes, transitionID } from "../src/signature.js";
import { sha256Hex } from "../src/canonical.js";
import { VM_EFFECTS, VM_TRANSITION_SCHEMA } from "../src/constants.js";
import { keyMaterial } from "./helpers.js";

async function withVM(handler) {
  const server = createServer((request, response) => {
    if (request.url.endsWith("/status")) {
      response.writeHead(200, { "content-type": "application/json" });
      response.end('{"healthy":true}');
    } else if (request.url.includes("/receipts/")) {
      response.writeHead(404, { "content-type": "application/json" });
      response.end('{"error":"not found"}');
    } else if (request.url.endsWith("/state")) {
      response.writeHead(500, { "content-type": "application/json" });
      response.end('{"error":"fixture failure"}');
    } else {
      response.writeHead(404, { "content-type": "application/json" });
      response.end('{}');
    }
  });
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  const client = new VMClient({
    endpoint: `http://127.0.0.1:${server.address().port}/ext/bc/BLOCKCHAIN-001`,
    request_timeout_ms: 1000,
    max_response_bytes: 1024
  });
  try {
    await handler(client);
  } finally {
    server.close();
    await once(server, "close");
  }
}

test("VM client accepts JSON, treats absent receipt as pending, and preserves HTTP errors", async () => {
  await withVM(async (client) => {
    assert.equal((await client.status()).healthy, true);
    assert.equal(await client.receipt("a".repeat(64)), null);
    await assert.rejects(
      () => client.state(),
      (error) => error instanceof VMHTTPError && error.status === 500 && /fixture failure/.test(error.message)
    );
  });
});

test("VM submission sends canonical envelope bytes matching the Go-compatible transition ID", async () => {
  const keys = keyMaterial();
  const unsigned = {
    schema: VM_TRANSITION_SCHEMA, operation: "OBSERVE_CROSSING", revision: 1,
    previous_state_commitment: "a".repeat(64), actor_id: keys.actorId,
    actor_public_key: keys.publicKeyHex, nonce: 0, locus_id: "LOCUS-001", observed_at: 1001,
    effect: VM_EFFECTS.OBSERVE_CROSSING,
    payload: { claim: "A&B <>\u2028\u2029", content_sha256: "b".repeat(64), matter_id: "MATTER-001", media_type: "text/plain" }
  };
  const signature = sign(null, signingBytes(unsigned), keys.privateKey).toString("hex");
  let received;
  const server = createServer(async (request, response) => {
    const chunks = [];
    for await (const chunk of request) chunks.push(chunk);
    received = Buffer.concat(chunks);
    response.writeHead(200, { "content-type": "application/json" });
    response.end(JSON.stringify({ transition_id: sha256Hex(received), status: "PENDING_CONSENSUS" }));
  });
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  try {
    const client = new VMClient({
      endpoint: `http://127.0.0.1:${server.address().port}`, request_timeout_ms: 1000, max_response_bytes: 1024
    });
    const reordered = { signature, unsigned: Object.fromEntries(Object.entries(unsigned).reverse()) };
    const result = await client.submit(reordered);
    assert.equal(result.transition_id, transitionID(reordered));
    assert.ok(received.toString().startsWith('{"unsigned":{"schema":'));
    assert.ok(received.toString().includes('"claim":"A\\u0026B \\u003c\\u003e\\u2028\\u2029"'));
    assert.doesNotMatch(received.toString(), /[<>&\u2028\u2029]/);
    assert.equal(JSON.parse(received).unsigned.payload.claim, unsigned.payload.claim);
  } finally {
    server.close();
    await once(server, "close");
  }
});

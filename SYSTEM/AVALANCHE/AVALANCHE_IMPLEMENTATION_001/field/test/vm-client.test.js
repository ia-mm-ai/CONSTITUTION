import test from "node:test";
import assert from "node:assert/strict";
import { createServer } from "node:http";
import { once } from "node:events";
import { VMClient, VMHTTPError } from "../src/vm-client.js";

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

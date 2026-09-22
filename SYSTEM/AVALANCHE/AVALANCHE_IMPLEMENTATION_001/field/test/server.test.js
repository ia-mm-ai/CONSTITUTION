import test from "node:test";
import assert from "node:assert/strict";
import { once } from "node:events";
import { createMediumServer } from "../src/server.js";

async function withServer(handler) {
  const runtime = {
    engine: {
      observe: async () => ({
        observation: { phase: "PRESENT", accepted_state_commitment: "a".repeat(64) },
        status: { capacity: { available_units: 1 }, pending_transitions: 0, successor_boundary: null }
      }),
      decide: async (capability) => ({
        schema: "PRESENCE_FIELD_DECISION_001",
        decision_id: "b".repeat(64),
        allowed: capability === "READ_PUBLIC_STATE",
        capability,
        observation: {},
        state: {}
      })
    }
  };
  const server = createMediumServer(runtime);
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  try {
    await handler(`http://127.0.0.1:${server.address().port}`);
  } finally {
    server.close();
    await once(server, "close");
  }
}

test("loopback API exposes public context and strict authorization requests", async () => {
  await withServer(async (base) => {
    const health = await fetch(`${base}/health`);
    assert.equal(health.status, 200);
    assert.equal((await health.json()).observation.phase, "PRESENT");

    const authorized = await fetch(`${base}/v1/authorize`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ capability: "READ_PUBLIC_STATE" })
    });
    assert.equal(authorized.status, 200);

    const extra = await fetch(`${base}/v1/authorize`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ capability: "READ_PUBLIC_STATE", hidden: true })
    });
    assert.equal(extra.status, 400);
    assert.match((await extra.json()).message, /unknown request field/);
  });
});

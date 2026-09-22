import { createServer } from "node:http";
import { isDigest } from "./canonical.js";

const MAX_REQUEST_BYTES = 1 << 20;

function send(response, status, value) {
  const encoded = Buffer.from(`${JSON.stringify(value)}\n`, "utf8");
  response.writeHead(status, {
    "content-type": "application/json; charset=utf-8",
    "content-length": encoded.length,
    "cache-control": "no-store",
    "x-content-type-options": "nosniff",
    "content-security-policy": "default-src 'none'",
    "referrer-policy": "no-referrer"
  });
  response.end(encoded);
}

async function readJSON(request) {
  const chunks = [];
  let size = 0;
  for await (const chunk of request) {
    size += chunk.length;
    if (size > MAX_REQUEST_BYTES) throw Object.assign(new Error("request exceeds 1 MiB"), { status: 413 });
    chunks.push(chunk);
  }
  if (size === 0) throw Object.assign(new Error("JSON request body is required"), { status: 400 });
  try {
    const value = JSON.parse(Buffer.concat(chunks).toString("utf8"));
    if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("body must be an object");
    return value;
  } catch (error) {
    throw Object.assign(new Error(`invalid JSON: ${error.message}`), { status: 400 });
  }
}

function pathSegments(url) {
  const parsed = new URL(url, "http://local.invalid");
  if (parsed.search) throw Object.assign(new Error("query parameters are not accepted"), { status: 400 });
  return parsed.pathname.split("/").filter(Boolean);
}

function exactFields(value, required, optional = []) {
  const allowed = new Set([...required, ...optional]);
  for (const key of Object.keys(value)) {
    if (!allowed.has(key)) throw Object.assign(new Error(`unknown request field ${key}`), { status: 400 });
  }
  for (const key of required) {
    if (!Object.hasOwn(value, key)) throw Object.assign(new Error(`missing request field ${key}`), { status: 400 });
  }
}

function observationView(result) {
  return {
    observation: result.observation,
    capacity: result.status.capacity,
    pending_transitions: result.status.pending_transitions,
    successor_boundary: result.status.successor_boundary,
    effect: "PUBLIC_ACCEPTED_STATE_VIEW_ONLY"
  };
}

export function createMediumServer(runtime) {
  return createServer(async (request, response) => {
    try {
      const segments = pathSegments(request.url);
      if (request.method === "GET" && segments.length === 1 && segments[0] === "health") {
        const observed = await runtime.engine.observe();
        send(response, 200, { healthy: true, ...observationView(observed) });
        return;
      }
      if (request.method === "GET" && segments.join("/") === "v1/context") {
        send(response, 200, observationView(await runtime.engine.observe()));
        return;
      }
      if (request.method === "POST" && segments.join("/") === "v1/authorize") {
        const body = await readJSON(request);
        exactFields(body, ["capability"], ["expected_state_commitment"]);
        const decision = await runtime.engine.decide(body.capability, {
          expectedStateCommitment: body.expected_state_commitment
        });
        const { observation: _observation, state: _state, ...publicDecision } = decision;
        send(response, decision.allowed ? 200 : 403, publicDecision);
        return;
      }
      if (request.method === "POST" && segments.join("/") === "v1/drafts") {
        send(response, 200, await runtime.engine.prepareDraft(await readJSON(request)));
        return;
      }
      if (request.method === "POST" && segments.join("/") === "v1/transitions") {
        send(response, 202, await runtime.engine.submitTransition(await readJSON(request)));
        return;
      }
      if (request.method === "GET" && segments.length === 3 && segments[0] === "v1" && segments[1] === "receipts") {
        if (!isDigest(segments[2])) throw Object.assign(new Error("receipt transition ID is invalid"), { status: 400 });
        send(response, 200, await runtime.engine.waitReceipt(segments[2]));
        return;
      }
      send(response, 404, { error: "NOT_FOUND" });
    } catch (error) {
      const status = error.status ?? (error.code === "CAPABILITY_DENIED" ? 403 : 400);
      send(response, status, {
        error: error.code ?? "PRESENCE_AVALANCHE_FIELD_ERROR",
        message: error.message,
        ...(error.decision ? { decision: error.decision } : {})
      });
    }
  });
}

export async function listen(runtime) {
  const server = createMediumServer(runtime);
  await new Promise((resolve, reject) => {
    server.once("error", reject);
    server.listen(runtime.config.listen_port, runtime.config.listen_host, resolve);
  });
  return server;
}

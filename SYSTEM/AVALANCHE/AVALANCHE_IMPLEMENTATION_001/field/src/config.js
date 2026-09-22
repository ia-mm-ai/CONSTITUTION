import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import { dirname, isAbsolute, resolve } from "node:path";
import {
  MEDIUM_CONFIG_SCHEMA,
  VM_ID,
  VM_RPCCHAINVM_PROTOCOL,
  VM_VERSION
} from "./constants.js";
import { assertPlainObject } from "./canonical.js";

const ACTOR_PATTERN = /^PRESENCE-AVALANCHE-ACTOR-[0-9a-f]{40}$/;
const HOST_ACTOR_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/;
const HEX_KEY_PATTERN = /^[0-9a-f]{64}$/;
const CONFIG_FIELDS = new Set([
  "schema", "medium_id", "endpoint", "expected_blockchain_id",
  "expected_locality_id", "expected_vm_id", "expected_vm_version",
  "expected_rpcchainvm_protocol", "actor_id", "actor_public_key",
  "profile_path", "journal_path", "request_timeout_ms",
  "receipt_timeout_ms", "max_response_bytes", "listen_host", "listen_port"
]);

export function actorIdFromPublicKey(publicKeyHex) {
  if (!HEX_KEY_PATTERN.test(publicKeyHex)) throw new Error("actor_public_key must be 32-byte lowercase hex");
  const digest = createHash("sha256").update(Buffer.from(publicKeyHex, "hex")).digest("hex");
  return `PRESENCE-AVALANCHE-ACTOR-${digest.slice(0, 40)}`;
}

function validateEndpoint(endpoint, expectedBlockchainId) {
  const url = new URL(endpoint);
  if (url.username || url.password || url.search || url.hash) throw new Error("endpoint must not contain credentials, query, or fragment");
  const loopback = url.hostname === "127.0.0.1" || url.hostname === "[::1]" || url.hostname === "::1";
  if (url.protocol !== "https:" && !(url.protocol === "http:" && loopback)) {
    throw new Error("endpoint must use HTTPS, except HTTP is allowed on loopback");
  }
  if (!url.pathname.split("/").includes(expectedBlockchainId)) {
    throw new Error("endpoint path must contain expected_blockchain_id exactly");
  }
  return url.toString().replace(/\/$/, "");
}

export function validateConfig(input) {
  assertPlainObject(input, "config");
  for (const field of Object.keys(input)) {
    if (!CONFIG_FIELDS.has(field)) throw new Error(`config contains unknown field ${field}`);
  }
  if (input.schema !== MEDIUM_CONFIG_SCHEMA) throw new Error(`config.schema must be ${MEDIUM_CONFIG_SCHEMA}`);
  if (typeof input.medium_id !== "string" || !/^[A-Z0-9][A-Z0-9_-]{2,95}$/.test(input.medium_id)) throw new Error("medium_id is invalid");
  if (typeof input.expected_blockchain_id !== "string" || input.expected_blockchain_id.length < 8) throw new Error("expected_blockchain_id is required");
  if (typeof input.expected_locality_id !== "string" || input.expected_locality_id.length < 3) throw new Error("expected_locality_id is required");
  if (input.expected_vm_id !== VM_ID) throw new Error(`expected_vm_id must be ${VM_ID}`);
  if (input.expected_vm_version !== VM_VERSION) throw new Error(`expected_vm_version must be ${VM_VERSION}`);
  if (input.expected_rpcchainvm_protocol !== VM_RPCCHAINVM_PROTOCOL) throw new Error(`expected_rpcchainvm_protocol must be ${VM_RPCCHAINVM_PROTOCOL}`);
  if (!HEX_KEY_PATTERN.test(input.actor_public_key ?? "")) throw new Error("actor_public_key must be 32-byte lowercase hex");
  const hostActor = input.actor_id === input.expected_locality_id && HOST_ACTOR_PATTERN.test(input.actor_id ?? "");
  if (!ACTOR_PATTERN.test(input.actor_id ?? "") && !hostActor) throw new Error("actor_id is invalid");
  if (actorIdFromPublicKey(input.actor_public_key) !== input.actor_id && !hostActor) {
    throw new Error("actor_id must derive from actor_public_key or name the expected host locality");
  }
  if (typeof input.profile_path !== "string" || !input.profile_path) throw new Error("profile_path is required");
  if (typeof input.journal_path !== "string" || !input.journal_path) throw new Error("journal_path is required");
  if (!Number.isSafeInteger(input.request_timeout_ms) || input.request_timeout_ms < 250 || input.request_timeout_ms > 60000) throw new Error("request_timeout_ms must be 250..60000");
  if (!Number.isSafeInteger(input.receipt_timeout_ms) || input.receipt_timeout_ms < 1000 || input.receipt_timeout_ms > 600000) throw new Error("receipt_timeout_ms must be 1000..600000");
  if (!Number.isSafeInteger(input.max_response_bytes) || input.max_response_bytes < 1024 || input.max_response_bytes > 16 * 1024 * 1024) throw new Error("max_response_bytes is outside bounds");
  const listenHost = input.listen_host ?? "127.0.0.1";
  if (!["127.0.0.1", "::1"].includes(listenHost)) throw new Error("listen_host must be an explicit loopback address");
  const listenPort = input.listen_port ?? 8787;
  if (!Number.isSafeInteger(listenPort) || listenPort < 1024 || listenPort > 65535) throw new Error("listen_port must be 1024..65535");
  const endpoint = validateEndpoint(input.endpoint, input.expected_blockchain_id);
  return Object.freeze({ ...input, endpoint, listen_host: listenHost, listen_port: listenPort });
}

export async function loadConfig(path) {
  const parsed = JSON.parse(await readFile(path, "utf8"));
  const base = dirname(resolve(path));
  if (typeof parsed.profile_path === "string" && !isAbsolute(parsed.profile_path)) parsed.profile_path = resolve(base, parsed.profile_path);
  if (typeof parsed.journal_path === "string" && !isAbsolute(parsed.journal_path)) parsed.journal_path = resolve(base, parsed.journal_path);
  return validateConfig(parsed);
}

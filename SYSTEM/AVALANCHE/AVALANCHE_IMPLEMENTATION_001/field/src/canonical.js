import { createHash } from "node:crypto";

export function assertPlainObject(value, label = "value") {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new TypeError(`${label} must be an object`);
  }
}

export function canonicalize(value) {
  if (value === null || typeof value === "boolean" || typeof value === "string") {
    return JSON.stringify(value);
  }
  if (typeof value === "number") {
    if (!Number.isSafeInteger(value)) throw new TypeError("canonical JSON numbers must be safe integers");
    return JSON.stringify(value);
  }
  if (Array.isArray(value)) return `[${value.map(canonicalize).join(",")}]`;
  assertPlainObject(value);
  const keys = Object.keys(value).sort();
  return `{${keys.map((key) => `${JSON.stringify(key)}:${canonicalize(value[key])}`).join(",")}}`;
}

export function canonicalBytes(value) {
  return Buffer.from(canonicalize(value), "utf8");
}

export function sha256Hex(value) {
  const input = Buffer.isBuffer(value) || value instanceof Uint8Array
    ? value
    : typeof value === "string"
      ? Buffer.from(value, "utf8")
      : canonicalBytes(value);
  return createHash("sha256").update(input).digest("hex");
}

export function isDigest(value) {
  return typeof value === "string" && /^[0-9a-f]{64}$/.test(value);
}

export function requireDigest(value, label) {
  if (!isDigest(value)) throw new Error(`${label} must be lowercase SHA-256 hex`);
  return value;
}

export function deepFreeze(value) {
  if (value && typeof value === "object" && !Object.isFrozen(value)) {
    Object.freeze(value);
    for (const child of Object.values(value)) deepFreeze(child);
  }
  return value;
}

import { assertPlainObject } from "./canonical.js";

export class VMHTTPError extends Error {
  constructor(status, message, body = null) {
    super(`VM HTTP ${status}: ${message}`);
    this.name = "VMHTTPError";
    this.status = status;
    this.body = body;
  }
}

// VMClient speaks the PRESENCE_AVALANCHE_VM_001 read/write HTTP surface. It only
// reads accepted material and posts already-signed canonical transitions. It
// never signs and never interprets; interpretation is the engine's job.
export class VMClient {
  constructor({ endpoint, request_timeout_ms, max_response_bytes }) {
    this.endpoint = endpoint.replace(/\/$/, "");
    this.timeout = request_timeout_ms;
    this.maxBytes = max_response_bytes;
  }

  async request(path, { method = "GET", body, rawBody, allowNotFound = false } = {}) {
    const headers = { accept: "application/json" };
    let encoded;
    if (rawBody !== undefined) {
      headers["content-type"] = "application/json";
      encoded = rawBody;
      if (Buffer.byteLength(encoded) > 1 << 20) throw new Error("request exceeds VM 1 MiB limit");
    } else if (body !== undefined) {
      headers["content-type"] = "application/json";
      encoded = JSON.stringify(body);
      if (Buffer.byteLength(encoded) > 1 << 20) throw new Error("request exceeds VM 1 MiB limit");
    }
    const response = await fetch(`${this.endpoint}${path}`, {
      method,
      headers,
      body: encoded,
      signal: AbortSignal.timeout(this.timeout),
      redirect: "error"
    });
    if (allowNotFound && response.status === 404) {
      await response.arrayBuffer();
      return null;
    }
    const length = Number(response.headers.get("content-length") ?? 0);
    if (length > this.maxBytes) throw new Error("VM response exceeds configured limit");
    const bytes = Buffer.from(await response.arrayBuffer());
    if (bytes.length > this.maxBytes) throw new Error("VM response exceeds configured limit");
    let parsed;
    try {
      parsed = JSON.parse(bytes.toString("utf8"));
    } catch {
      throw new VMHTTPError(response.status, "response is not JSON");
    }
    if (!response.ok) throw new VMHTTPError(response.status, parsed?.error ?? response.statusText, parsed);
    assertPlainObject(parsed, "VM response");
    return parsed;
  }

  status() { return this.request("/status"); }
  state() { return this.request("/state"); }
  capacity() { return this.request("/capacity"); }
  currentness() { return this.request("/currentness"); }
  genesis() { return this.request("/genesis"); }
  draft(request) { return this.request("/drafts", { method: "POST", body: request }); }
  submit(transition) { return this.request("/transitions", { method: "POST", body: transition }); }
  // submitCanonical posts the exact canonical bytes whose SHA-256 is the
  // transition id, so the VM's byte-equal canonical decode cannot diverge from
  // what the FIELD signed and hashed.
  submitCanonical(canonicalBytes) { return this.request("/transitions", { method: "POST", rawBody: canonicalBytes }); }
  receipt(id) { return this.request(`/receipts/${encodeURIComponent(id)}`, { allowNotFound: true }); }
  transition(id) { return this.request(`/transitions/${encodeURIComponent(id)}`, { allowNotFound: true }); }
}

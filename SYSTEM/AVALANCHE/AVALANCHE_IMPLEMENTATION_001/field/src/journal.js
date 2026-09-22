import { open, readFile, stat } from "node:fs/promises";
import { dirname } from "node:path";
import { mkdir } from "node:fs/promises";
import { canonicalize, sha256Hex } from "./canonical.js";
import { MEDIUM_EVENT_SCHEMA } from "./constants.js";

const ZERO = "0".repeat(64);
const RESERVED = new Set(["schema", "sequence", "previous_event_sha256", "recorded_at", "type", "event_sha256"]);
const FORBIDDEN_CONTENT_FIELDS = new Set(["payload", "unsigned", "signature", "private_key", "private_state", "content", "secret"]);

function committedEvent(event) {
  const without = { ...event };
  delete without.event_sha256;
  return without;
}

export class AppendOnlyJournal {
  constructor(path) {
    this.path = path;
    this.sequence = 0;
    this.previous = ZERO;
    this.tail = Promise.resolve();
  }

  async initialize() {
    await mkdir(dirname(this.path), { recursive: true, mode: 0o700 });
    try {
      const info = await stat(this.path);
      if ((info.mode & 0o077) !== 0) throw new Error("journal permissions are broader than 0600");
      const contents = await readFile(this.path, "utf8");
      let expectedSequence = 1;
      let previous = ZERO;
      for (const line of contents.split("\n")) {
        if (!line) continue;
        const event = JSON.parse(line);
        if (event.schema !== MEDIUM_EVENT_SCHEMA || event.sequence !== expectedSequence || event.previous_event_sha256 !== previous) {
          throw new Error("journal chain is malformed");
        }
        const expected = sha256Hex(canonicalize(committedEvent(event)));
        if (event.event_sha256 !== expected) throw new Error("journal event digest mismatch");
        previous = expected;
        expectedSequence += 1;
      }
      this.sequence = expectedSequence - 1;
      this.previous = previous;
    } catch (error) {
      if (error.code !== "ENOENT") throw error;
      const handle = await open(this.path, "wx", 0o600);
      await handle.close();
    }
    return this;
  }

  append(type, fields = {}) {
    const operation = this.tail.then(() => this.appendSerialized(type, fields));
    this.tail = operation.catch(() => {});
    return operation;
  }

  async appendSerialized(type, fields = {}) {
    if (typeof type !== "string" || !/^[A-Z][A-Z0-9_]{1,95}$/.test(type)) throw new Error("journal event type is invalid");
    for (const key of Object.keys(fields)) {
      if (RESERVED.has(key)) throw new Error(`journal field ${key} is reserved`);
    }
    const inspect = (value) => {
      if (!value || typeof value !== "object") return;
      for (const [key, child] of Object.entries(value)) {
        if (FORBIDDEN_CONTENT_FIELDS.has(key.toLowerCase())) throw new Error(`journal field ${key} may contain non-public content`);
        inspect(child);
      }
    };
    inspect(fields);
    const event = {
      schema: MEDIUM_EVENT_SCHEMA,
      sequence: this.sequence + 1,
      previous_event_sha256: this.previous,
      recorded_at: Math.floor(Date.now() / 1000),
      type,
      ...fields
    };
    event.event_sha256 = sha256Hex(canonicalize(committedEvent(event)));
    const handle = await open(this.path, "a", 0o600);
    try {
      await handle.write(`${canonicalize(event)}\n`);
      await handle.sync();
    } finally {
      await handle.close();
    }
    this.sequence = event.sequence;
    this.previous = event.event_sha256;
    return event;
  }
}

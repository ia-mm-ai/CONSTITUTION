import test from "node:test";
import assert from "node:assert/strict";
import { readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { AppendOnlyJournal } from "../src/journal.js";
import { testDirectory } from "./helpers.js";

test("journal verifies its append-only hash chain and detects mutation", async () => {
  const directory = await testDirectory();
  const path = join(directory, "events.ndjson");
  const journal = await new AppendOnlyJournal(path).initialize();
  await journal.append("TEST", { state_commitment: "a".repeat(64) });
  const reopened = await new AppendOnlyJournal(path).initialize();
  assert.equal(reopened.sequence, 1);
  const contents = await readFile(path, "utf8");
  await writeFile(path, contents.replace(`"${"a".repeat(64)}"`, `"${"b".repeat(64)}"`), { mode: 0o600 });
  await assert.rejects(() => new AppendOnlyJournal(path).initialize(), /digest mismatch/);
});

test("concurrent journal calls serialize and reserved or content-bearing fields are refused", async () => {
  const directory = await testDirectory();
  const path = join(directory, "events.ndjson");
  const journal = await new AppendOnlyJournal(path).initialize();
  await Promise.all(Array.from({ length: 32 }, (_, index) => journal.append("DECISION", { index })));
  const reopened = await new AppendOnlyJournal(path).initialize();
  assert.equal(reopened.sequence, 32);
  await assert.rejects(() => journal.append("BAD", { sequence: 100 }), /reserved/);
  await assert.rejects(() => journal.append("BAD", { payload: { raw: true } }), /non-public content/);
});

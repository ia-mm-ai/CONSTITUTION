#!/usr/bin/env node
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import process from "node:process";
import { sha256Hex } from "../src/canonical.js";
import { buildPassage, passageCommitment } from "../src/continuity.js";
import { MEDIUM_PROTOCOL, MEDIUM_VERSION, VM_ID, VM_VERSION } from "../src/constants.js";
import { actorIdFromPublicKey } from "../src/config.js";
import { AppendOnlyJournal } from "../src/journal.js";
import { createRuntime } from "../src/runtime.js";
import { listen } from "../src/server.js";

function fail(message) {
  process.stderr.write(`presence-avalanche-field: ${message}\n`);
  process.exitCode = 1;
}

function valueAfter(args, name) {
  const index = args.indexOf(name);
  if (index < 0 || index + 1 >= args.length) throw new Error(`${name} is required`);
  return args[index + 1];
}

function usage() {
  return [
    "Usage:",
    "  presence-avalanche-field --version-json",
    "  presence-avalanche-field inspect --config <config.json>",
    "  presence-avalanche-field serve --config <config.json>",
    "  presence-avalanche-field actor-id --public-key <64-lowercase-hex>",
    "  presence-avalanche-field commit <file>",
    "  presence-avalanche-field successor --participant <id> --from <sha256> --delta <sha256> [--delta <sha256> ...]",
    "  presence-avalanche-field verify-journal <journal.ndjson>"
  ].join("\n");
}

async function main() {
  const args = process.argv.slice(2);
  if (args.length === 1 && args[0] === "--version-json") {
    process.stdout.write(`${JSON.stringify({ protocol: MEDIUM_PROTOCOL, version: MEDIUM_VERSION, vm_id: VM_ID, vm_version: VM_VERSION })}\n`);
    return;
  }
  const command = args[0];
  if (command === "inspect") {
    const runtime = await createRuntime(resolve(valueAfter(args, "--config")));
    const { observation, status } = await runtime.engine.observe();
    process.stdout.write(`${JSON.stringify({ observation, capacity: status.capacity, pending_transitions: status.pending_transitions }, null, 2)}\n`);
    return;
  }
  if (command === "actor-id") {
    const publicKey = valueAfter(args, "--public-key");
    process.stdout.write(`${JSON.stringify({ actor_id: actorIdFromPublicKey(publicKey), actor_public_key: publicKey })}\n`);
    return;
  }
  if (command === "serve") {
    const runtime = await createRuntime(resolve(valueAfter(args, "--config")));
    const server = await listen(runtime);
    const address = server.address();
    process.stdout.write(`${JSON.stringify({ status: "LISTENING", host: address.address, port: address.port, medium_id: runtime.config.medium_id })}\n`);
    const close = () => server.close(() => process.exit(0));
    process.once("SIGINT", close);
    process.once("SIGTERM", close);
    return;
  }
  if (command === "commit" && args.length === 2) {
    const bytes = await readFile(resolve(args[1]));
    process.stdout.write(`${JSON.stringify({ sha256: sha256Hex(bytes), bytes: bytes.length, effect: "COMMITMENT_ONLY_CONTENT_NOT_DISCLOSED" })}\n`);
    return;
  }
  if (command === "successor") {
    const participant = valueAfter(args, "--participant");
    const from = valueAfter(args, "--from");
    const deltas = [];
    for (let index = 0; index < args.length; index += 1) {
      if (args[index] === "--delta") deltas.push(args[index + 1]);
    }
    const built = buildPassage(participant, from, deltas);
    const commitment = passageCommitment(participant, from, built.passage, built.result_state_commitment);
    process.stdout.write(`${JSON.stringify({ participant_id: participant, from_state_commitment: from, ...built, passage_commitment: commitment }, null, 2)}\n`);
    return;
  }
  if (command === "verify-journal" && args.length === 2) {
    const journal = await new AppendOnlyJournal(resolve(args[1])).initialize();
    process.stdout.write(`${JSON.stringify({ valid: true, events: journal.sequence, head: journal.previous })}\n`);
    return;
  }
  throw new Error(usage());
}

main().catch((error) => fail(error.message));

// Deterministic FIELD qualification runner for AVALANCHE_IMPLEMENTATION_001.
//
// Runs the local, deterministic qualification steps and writes a machine
// readable report. It performs no network access and no release act. Steps a
// runner cannot perform locally (multi-validator restart/rejoin, external
// advisory queries) are qualification boundaries recorded elsewhere in
// records/QUALIFICATION_STATUS.json — this script never claims them.

import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import process from "node:process";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const artifacts = resolve(root, "artifacts");
await mkdir(artifacts, { recursive: true });

function run(command, args) {
  const result = spawnSync(command, args, { cwd: root, encoding: "utf8" });
  return {
    command: [command, ...args].join(" "),
    status: result.status,
    stdout: result.stdout,
    stderr: result.stderr
  };
}

const verify = run(process.execPath, ["./scripts/verify.mjs"]);
const tests = run(process.execPath, ["--test", "test/cli.test.js", "test/config.test.js", "test/continuity.test.js", "test/engine.test.js", "test/journal.test.js", "test/profile.test.js", "test/scope-contract.test.js", "test/server.test.js", "test/state.test.js", "test/vm-client.test.js"]);
const kernel = run("uname", ["-srm"]);

const scopeContractBytes = await readFile(resolve(root, "../contract/OPERATION_SCOPE_CONTRACT_001.json"));

const report = {
  schema: "PRESENCE_FIELD_QUALIFICATION_001",
  implementation: "AVALANCHE_IMPLEMENTATION_001",
  version: "1.0.0",
  kernel: kernel.stdout.trim(),
  node: process.version,
  operation_scope_contract_sha256: createHash("sha256").update(scopeContractBytes).digest("hex"),
  steps: {
    verify: { command: verify.command, status: verify.status },
    tests: {
      command: tests.command,
      status: tests.status,
      passed: Number(/^# pass (\d+)$/m.exec(tests.stdout)?.[1] ?? 0),
      failed: Number(/^# fail (\d+)$/m.exec(tests.stdout)?.[1] ?? -1)
    }
  },
  qualified: verify.status === 0 && tests.status === 0
};

const reportPath = resolve(artifacts, "FIELD_QUALIFICATION_REPORT.json");
await writeFile(reportPath, `${JSON.stringify(report, null, 2)}\n`);
process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
if (!report.qualified) process.exit(1);

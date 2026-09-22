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

function sha256(bytes) {
  return createHash("sha256").update(bytes).digest("hex");
}

const syntaxAndContract = run(process.execPath, ["./scripts/verify.mjs"]);
const tests = run(process.execPath, ["--test", "test/cli.test.js", "test/config.test.js", "test/continuity.test.js", "test/engine.test.js", "test/journal.test.js", "test/profile.test.js", "test/server.test.js", "test/state.test.js", "test/vm-client.test.js"]);
const coverage = run(process.execPath, ["--test", "--experimental-test-coverage", "test/cli.test.js", "test/config.test.js", "test/continuity.test.js", "test/engine.test.js", "test/journal.test.js", "test/profile.test.js", "test/server.test.js", "test/state.test.js", "test/vm-client.test.js"]);
const kernel = run("uname", ["-srm"]);
const publicOrigin = run("python3", ["./scripts/verify_public_origin.py", "--report", "./artifacts/PRESENCE_AVALANCHE_FIELD_001-v2.0.0-public-origin.json"]);

let predecessor = { supplied: false, expected_sha256: "96007f0ffe089490d937e6dce18a225debedb521f86bfd2b323f6f2cc09bcdf2" };
if (process.env.LOCALITY_MEDIUM_PREDECESSOR_PACK) {
  const path = resolve(process.env.LOCALITY_MEDIUM_PREDECESSOR_PACK);
  const digest = sha256(await readFile(path));
  predecessor = {
    supplied: true,
    filename: path.split("/").at(-1),
    expected_sha256: predecessor.expected_sha256,
    observed_sha256: digest,
    exact: digest === predecessor.expected_sha256
  };
}

let consensusPredecessor = { supplied: false, expected_sha256: "5597eab61802b1df92636b6df88fea46b88f460c4e5f2c51e393c5c6ce7a214f" };
if (process.env.LOCALITY_MEDIUM_VM003_PACK) {
  const path = resolve(process.env.LOCALITY_MEDIUM_VM003_PACK);
  const digest = sha256(await readFile(path));
  const stateSource = run("unzip", ["-p", path, "state.go"]);
  const operationsSource = run("unzip", ["-p", path, "operations.go"]);
  const wireSchemaExact = stateSource.status === 0 && operationsSource.status === 0 &&
    stateSource.stdout.includes('Schema: "PRESENCE_RECEIPT_001"') &&
    operationsSource.stdout.includes('continuityStepSchema    = "PRESENCE_PARTICIPANT_STATE_SUCCESSION_001"') &&
    operationsSource.stdout.includes('continuityPassageSchema = "PRESENCE_PARTICIPANT_STATE_PASSAGE_001"');
  consensusPredecessor = {
    supplied: true,
    filename: path.split("/").at(-1),
    expected_sha256: consensusPredecessor.expected_sha256,
    observed_sha256: digest,
    exact: digest === consensusPredecessor.expected_sha256,
    wire_schema_exact: wireSchemaExact
  };
}

let vm003Evidence = { supplied: false, compatible: false };
if (process.env.LOCALITY_MEDIUM_VM003_EVIDENCE) {
  const path = resolve(process.env.LOCALITY_MEDIUM_VM003_EVIDENCE);
  const output = resolve(artifacts, "PRESENCE_AVALANCHE_FIELD_001-v2.0.0-vm003-evidence-compatibility.json");
  const verification = run(process.execPath, ["./scripts/verify-vm003-evidence.mjs", path, output]);
  vm003Evidence = {
    supplied: true,
    filename: path.split("/").at(-1),
    sha256: sha256(await readFile(path)),
    compatible: verification.status === 0,
    stdout: verification.stdout,
    stderr: verification.stderr
  };
}

const passed = syntaxAndContract.status === 0 && tests.status === 0 && coverage.status === 0 && publicOrigin.status === 0 && predecessor.supplied && predecessor.exact && consensusPredecessor.supplied && consensusPredecessor.exact && consensusPredecessor.wire_schema_exact && vm003Evidence.supplied && vm003Evidence.compatible;
const report = {
  schema: "PRESENCE_FIELD_QUALIFICATION_001",
  protocol: "PRESENCE_AVALANCHE_FIELD_001",
  version: "2.0.0",
  qualified: passed,
  environment: {
    node: process.version,
    platform: process.platform,
    architecture: process.arch,
    kernel: kernel.stdout.trim()
  },
  direct_predecessor: predecessor,
  consensus_predecessor: consensusPredecessor,
  vm003_evidence_compatibility: vm003Evidence,
  public_origin: {
    commit: "4cf5a926d5fece4e3cccfdfcab40e16431dd332b",
    exact_binding: publicOrigin.status === 0
  },
  gates: {
    syntax_contract_profiles_manifest: syntaxAndContract.status === 0 ? "PASS" : "FAIL",
    adversarial_tests: tests.status === 0 ? "PASS" : "FAIL",
    coverage_run: coverage.status === 0 ? "PASS" : "FAIL",
    public_origin_exact_binding: publicOrigin.status === 0 ? "PASS" : "FAIL",
    published_predecessor_identity: predecessor.supplied && predecessor.exact ? "PASS" : "FAIL",
    vm003_identity: consensusPredecessor.supplied && consensusPredecessor.exact ? "PASS" : "FAIL",
    vm003_wire_schemas: consensusPredecessor.wire_schema_exact ? "PASS" : "FAIL",
    qualified_vm003_state_evidence_compatibility: vm003Evidence.supplied && vm003Evidence.compatible ? "PASS" : "FAIL",
    live_successor_l1_coupling: "NOT_RUN_DEPLOYMENT_SPECIFIC"
  },
  claim_ceiling: "REFERENCE_IMPLEMENTATION_QUALIFIED_LIVE_SUCCESSOR_COUPLING_REQUIRED"
};

await writeFile(resolve(artifacts, "PRESENCE_AVALANCHE_FIELD_001-v2.0.0-qualification.json"), `${JSON.stringify(report, null, 2)}\n`, { mode: 0o644 });
await writeFile(resolve(artifacts, "PRESENCE_AVALANCHE_FIELD_001-v2.0.0-tests.txt"), `${tests.stdout}${tests.stderr}`, { mode: 0o644 });
await writeFile(resolve(artifacts, "PRESENCE_AVALANCHE_FIELD_001-v2.0.0-coverage.txt"), `${coverage.stdout}${coverage.stderr}`, { mode: 0o644 });
await writeFile(resolve(artifacts, "PRESENCE_AVALANCHE_FIELD_001-v2.0.0-verification.txt"), `${syntaxAndContract.stdout}${syntaxAndContract.stderr}${publicOrigin.stdout}${publicOrigin.stderr}`, { mode: 0o644 });

process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
if (!passed) process.exit(1);

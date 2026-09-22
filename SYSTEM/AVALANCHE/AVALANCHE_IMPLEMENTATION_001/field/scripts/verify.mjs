import { createHash } from "node:crypto";
import { lstat, readFile, readdir } from "node:fs/promises";
import { dirname, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import { validateProfile } from "../src/profile.js";
import { CONTINUITY_PASSAGE_SCHEMA, CONTINUITY_STEP_SCHEMA } from "../src/continuity.js";
import { OPERATION_SCOPES, OPERATION_SCOPE_CONTRACT } from "../src/scope-contract.js";
import {
  MEDIUM_CONTRACT_SCHEMA,
  MEDIUM_PROTOCOL,
  MEDIUM_VERSION,
  VM_FORM_ID,
  VM_FORM_SHA256,
  VM_ID,
  VM_PROTOCOL,
  VM_RECEIPT_SCHEMA,
  VM003_REPOSITORY_PACK_SHA256,
  VM_RPCCHAINVM_PROTOCOL,
  VM_STATE_SCHEMA,
  VM_TRANSITION_SCHEMA,
  VM_VERSION
} from "../src/constants.js";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");

function sha256(bytes) {
  return createHash("sha256").update(bytes).digest("hex");
}

async function walk(directory) {
  const result = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    if (entry.name === ".git" || entry.name === "node_modules" || entry.name === "artifacts") continue;
    const path = resolve(directory, entry.name);
    const info = await lstat(path);
    if (info.isSymbolicLink()) throw new Error(`symbolic link is forbidden in implementation source: ${relative(root, path)}`);
    if (info.isDirectory()) result.push(...await walk(path));
    else if (info.isFile()) result.push(path);
  }
  return result.sort();
}

function requireEqual(actual, expected, label) {
  if (actual !== expected) throw new Error(`${label}: expected ${expected}, got ${actual}`);
}

function runNode(args, label) {
  const result = spawnSync(process.execPath, args, { cwd: root, encoding: "utf8" });
  if (result.status !== 0) throw new Error(`${label} failed\n${result.stdout}${result.stderr}`);
  return result.stdout;
}

const files = await walk(root);
const forbiddenNames = new Set(["staker.key", "signer.key", "private.key", "id_rsa", "id_ed25519"]);
for (const path of files) {
  const name = path.split("/").at(-1).toLowerCase();
  if (forbiddenNames.has(name) || /\.(pem|p12|pfx|key)$/.test(name)) {
    throw new Error(`secret-bearing filename is forbidden: ${relative(root, path)}`);
  }
  if (/\.(js|mjs)$/.test(path)) runNode(["--check", path], `syntax ${relative(root, path)}`);
  if (path.endsWith(".json")) {
    const value = JSON.parse(await readFile(path, "utf8"));
    const searchKeys = (node) => {
      if (!node || typeof node !== "object") return;
      for (const [key, child] of Object.entries(node)) {
        if (["private_key", "mnemonic", "seed_phrase"].includes(key.toLowerCase())) {
          throw new Error(`forbidden secret field in ${relative(root, path)}`);
        }
        searchKeys(child);
      }
    };
    searchKeys(value);
  }
}

const packageDocument = JSON.parse(await readFile(resolve(root, "package.json"), "utf8"));
requireEqual(packageDocument.version, MEDIUM_VERSION, "package version");
if (packageDocument.dependencies || packageDocument.devDependencies) throw new Error("third-party package dependencies are not permitted");

const contract = JSON.parse(await readFile(resolve(root, "contract/PRESENCE_FIELD_CONTRACT_001.json"), "utf8"));
requireEqual(contract.schema, MEDIUM_CONTRACT_SCHEMA, "contract schema");
requireEqual(contract.version, MEDIUM_VERSION, "contract version");
requireEqual(contract.medium_protocol, MEDIUM_PROTOCOL, "contract protocol");
requireEqual(contract.coupled_vm.protocol, VM_PROTOCOL, "coupled VM protocol");
requireEqual(contract.coupled_vm.version, VM_VERSION, "coupled VM version");
requireEqual(contract.coupled_vm.vm_id, VM_ID, "coupled VM ID");
requireEqual(contract.coupled_vm.rpcchainvm_protocol, VM_RPCCHAINVM_PROTOCOL, "coupled VM RPC protocol");
requireEqual(contract.coupled_vm.state_schema, VM_STATE_SCHEMA, "coupled VM state schema");
requireEqual(contract.coupled_vm.transition_schema, VM_TRANSITION_SCHEMA, "coupled VM transition schema");
requireEqual(contract.coupled_vm.receipt_schema, VM_RECEIPT_SCHEMA, "coupled VM receipt schema");
requireEqual(contract.coupled_vm.continuity_step_schema, CONTINUITY_STEP_SCHEMA, "coupled VM continuity-step schema");
requireEqual(contract.coupled_vm.continuity_passage_schema, CONTINUITY_PASSAGE_SCHEMA, "coupled VM continuity-passage schema");
requireEqual(contract.coupled_vm.form_id, VM_FORM_ID, "coupled VM form ID");
requireEqual(contract.coupled_vm.form_sha256, VM_FORM_SHA256, "coupled VM form digest");
requireEqual(contract.coupled_vm.implementation, "AVALANCHE_IMPLEMENTATION_001", "coupled VM implementation");

// One shared operation-scope contract: the exact file the VM embeds.
const scopeContractPath = resolve(root, contract.operation_scope_contract.path);
const scopeContractBytes = await readFile(scopeContractPath);
requireEqual(sha256(scopeContractBytes), contract.operation_scope_contract.file_sha256, "operation-scope contract digest");
const scopeContract = JSON.parse(scopeContractBytes.toString("utf8"));
requireEqual(scopeContract.schema, contract.operation_scope_contract.schema, "operation-scope contract schema");
requireEqual(scopeContract.schema, OPERATION_SCOPE_CONTRACT.schema, "runtime operation-scope schema");
requireEqual(Object.keys(scopeContract.operations).length, Object.keys(OPERATION_SCOPES).length, "operation-scope inventory size");

// Predecessor archives must remain byte-identical in STATE/.
for (const predecessor of contract.predecessors) {
  const archivePath = resolve(root, predecessor.state_archive);
  const digest = sha256(await readFile(archivePath));
  requireEqual(digest, predecessor.state_archive_sha256, `${predecessor.identity} state archive digest`);
  requireEqual(predecessor.disposition, "PRESERVED_BYTE_IDENTICAL", `${predecessor.identity} disposition`);
}
const vmPredecessor = contract.predecessors.find((entry) => entry.identity === "LOCALITY_VM_003");
requireEqual(vmPredecessor.repository_pack_sha256, VM003_REPOSITORY_PACK_SHA256, "VM003 repository pack digest");

const kinds = [];
for (const reference of contract.reference_profiles) {
  const path = resolve(root, reference.path);
  const bytes = await readFile(path);
  requireEqual(sha256(bytes), reference.file_sha256, `${reference.profile_id} digest`);
  const profile = validateProfile(JSON.parse(bytes.toString("utf8")));
  requireEqual(profile.profile_id, reference.profile_id, "profile ID");
  requireEqual(profile.embodiment_kind, reference.embodiment_kind, "profile kind");
  kinds.push(profile.embodiment_kind);
}
requireEqual([...new Set(kinds)].sort().join(","), "AI,DEVICE,HUMAN,LOCALITY", "reference profile kinds");

const testFiles = files.filter((path) => path.endsWith(".test.js"));
const testOutput = runNode(["--test", ...testFiles], "test suite");

process.stdout.write(`${JSON.stringify({
  verified: true,
  protocol: MEDIUM_PROTOCOL,
  version: MEDIUM_VERSION,
  source_files_checked: files.length,
  reference_profiles: kinds.length,
  operation_scope_contract_sha256: contract.operation_scope_contract.file_sha256,
  tests: (testOutput.match(/^(✔|ok \d)/gm) ?? []).length,
  predecessor_archives_byte_identical: true
}, null, 2)}\n`);

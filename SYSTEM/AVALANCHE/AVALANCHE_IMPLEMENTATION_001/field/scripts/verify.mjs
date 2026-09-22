import { createHash } from "node:crypto";
import { lstat, readFile, readdir } from "node:fs/promises";
import { dirname, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import { validateProfile } from "../src/profile.js";
import { CONTINUITY_PASSAGE_SCHEMA, CONTINUITY_STEP_SCHEMA } from "../src/continuity.js";
import {
  MEDIUM_CONTRACT_SCHEMA,
  MEDIUM_PROTOCOL,
  MEDIUM_VERSION,
  VM_FORM_ID,
  VM_FORM_SHA256,
  VM_ID,
  VM_PROTOCOL,
  VM_RECEIPT_SCHEMA,
  VM_REPOSITORY_SHA256,
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
    if (info.isSymbolicLink()) throw new Error(`symbolic link is forbidden in release source: ${relative(root, path)}`);
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

function runCommand(command, args, label) {
  const result = spawnSync(command, args, { cwd: root, encoding: "utf8" });
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
requireEqual(contract.predecessor.protocol, VM_PROTOCOL, "consensus protocol");
requireEqual(contract.predecessor.version, VM_VERSION, "predecessor version");
requireEqual(contract.predecessor.vm_id, VM_ID, "predecessor VM ID");
requireEqual(contract.predecessor.rpcchainvm_protocol, VM_RPCCHAINVM_PROTOCOL, "predecessor RPC protocol");
requireEqual(contract.predecessor.state_schema, VM_STATE_SCHEMA, "predecessor state schema");
requireEqual(contract.predecessor.transition_schema, VM_TRANSITION_SCHEMA, "predecessor transition schema");
requireEqual(contract.predecessor.receipt_schema, VM_RECEIPT_SCHEMA, "predecessor receipt schema");
requireEqual(contract.predecessor.continuity_step_schema, CONTINUITY_STEP_SCHEMA, "predecessor continuity-step schema");
requireEqual(contract.predecessor.continuity_passage_schema, CONTINUITY_PASSAGE_SCHEMA, "predecessor continuity-passage schema");
requireEqual(contract.predecessor.form_id, VM_FORM_ID, "predecessor form ID");
requireEqual(contract.predecessor.form_sha256, VM_FORM_SHA256, "predecessor form digest");
requireEqual(contract.predecessor.repository_pack_sha256, VM_REPOSITORY_SHA256, "predecessor pack digest");
requireEqual(contract.predecessor.mutation, "NONE", "predecessor mutation");
requireEqual(contract.direct_predecessor.protocol, MEDIUM_PROTOCOL, "direct predecessor protocol");
requireEqual(contract.direct_predecessor.version, "1.0.0", "direct predecessor version");
requireEqual(contract.direct_predecessor.repository_pack_sha256, "96007f0ffe089490d937e6dce18a225debedb521f86bfd2b323f6f2cc09bcdf2", "direct predecessor pack digest");
requireEqual(contract.direct_predecessor.public_repository_commit, "0bd963f46d447d0e41d494c6a782c1c2db9246e8", "direct predecessor publication commit");
requireEqual(contract.direct_predecessor.disposition, "PRESERVED_BYTE_IDENTICAL", "direct predecessor disposition");
requireEqual(contract.public_origin.commit, "4cf5a926d5fece4e3cccfdfcab40e16431dd332b", "public origin commit");

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

const predecessor = JSON.parse(await readFile(resolve(root, "lineage/PREDECESSOR_LOCK.json"), "utf8"));
requireEqual(predecessor.schema, "LOCALITY_MEDIUM_SUCCESSION_LOCK_002", "lineage schema");
requireEqual(predecessor.direct_predecessor.artifacts.repository_pack_sha256, "96007f0ffe089490d937e6dce18a225debedb521f86bfd2b323f6f2cc09bcdf2", "lineage direct predecessor pack digest");
requireEqual(predecessor.direct_predecessor.disposition, "PRESERVED_BYTE_IDENTICAL", "lineage direct predecessor disposition");
requireEqual(predecessor.consensus_predecessor.repository_pack_sha256, VM_REPOSITORY_SHA256, "lineage VM003 pack digest");
requireEqual(predecessor.consensus_predecessor.vm_id, VM_ID, "lineage VM003 ID");
requireEqual(predecessor.consensus_predecessor.mutation, "NONE", "lineage VM003 mutation");
requireEqual(predecessor.public_origin.commit, "4cf5a926d5fece4e3cccfdfcab40e16431dd332b", "lineage public origin commit");
requireEqual(predecessor.successor.version, MEDIUM_VERSION, "lineage successor version");

const release = JSON.parse(await readFile(resolve(root, "RELEASE_MANIFEST.json"), "utf8"));
requireEqual(release.release.protocol, MEDIUM_PROTOCOL, "release protocol");
requireEqual(release.release.version, MEDIUM_VERSION, "release version");
requireEqual(release.direct_predecessor.disposition, "PRESERVED_BYTE_IDENTICAL", "release direct predecessor disposition");
requireEqual(release.consensus_predecessor.repository_pack_sha256, VM_REPOSITORY_SHA256, "release VM003 pack digest");
requireEqual(release.consensus_predecessor.mutation, "NONE", "release VM003 mutation");
requireEqual(release.public_origin.commit, "4cf5a926d5fece4e3cccfdfcab40e16431dd332b", "release public origin commit");
for (const [name, artifact] of Object.entries(release.bound_artifacts)) {
  requireEqual(sha256(await readFile(resolve(root, artifact.path))), artifact.sha256, `release artifact ${name}`);
}

runCommand("python3", ["./scripts/verify_public_origin.py"], "public origin verification");

const testFiles = files.filter((path) => path.endsWith(".test.js"));
const testOutput = runNode(["--test", ...testFiles], "test suite");

const manifestPath = resolve(root, "MANIFEST.sha256");
if (files.includes(manifestPath)) {
  for (const line of (await readFile(manifestPath, "utf8")).trim().split("\n")) {
    const match = /^([0-9a-f]{64})  (.+)$/.exec(line);
    if (!match) throw new Error("MANIFEST.sha256 contains a malformed line");
    const path = resolve(root, match[2]);
    requireEqual(sha256(await readFile(path)), match[1], `manifest ${match[2]}`);
  }
}

process.stdout.write(`${JSON.stringify({
  verified: true,
  protocol: MEDIUM_PROTOCOL,
  version: MEDIUM_VERSION,
  source_files_checked: files.length,
  reference_profiles: kinds.length,
  tests: (testOutput.match(/^✔/gm) ?? []).length,
  direct_predecessor_mutated: false,
  consensus_predecessor_mutated: false,
  public_origin_verified: true
}, null, 2)}\n`);

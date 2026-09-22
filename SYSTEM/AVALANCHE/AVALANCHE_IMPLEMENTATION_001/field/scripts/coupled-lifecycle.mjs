#!/usr/bin/env node
// Coupled FIELD ⇄ VM lifecycle rehearsal for AVALANCHE_IMPLEMENTATION_001.
//
// Drives the REAL MediumEngine (FIELD) against the REAL VM served in-process
// by `presence-avalanche-vm --serve-rehearsal` on loopback. It reruns the
// exact predecessor coupling failure beyond its former breaking point:
//
//   read → BOUND → receipt → renewed read → DECLARE_CAPACITY (empty locus)
//        → receipt → renewed read → PULSE → CORRECT (target-scoped) → read
//
// All key material is disposable and confined to a caller-supplied temporary
// directory. Nothing here touches Mainnet, Fuji, or any public network.
//
// Usage: node scripts/coupled-lifecycle.mjs VM_BINARY WORK_DIR [PORT]

import { spawn, execFileSync } from "node:child_process";
import { generateKeyPairSync, sign } from "node:crypto";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join, resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const fieldRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const implementationRoot = resolve(fieldRoot, "..");

const { MediumEngine } = await import(join(fieldRoot, "src/engine.js"));
const { VMClient } = await import(join(fieldRoot, "src/vm-client.js"));
const { AppendOnlyJournal } = await import(join(fieldRoot, "src/journal.js"));
const { loadProfile } = await import(join(fieldRoot, "src/profile.js"));
const { actorIdFromPublicKey } = await import(join(fieldRoot, "src/config.js"));
const { signingBytes } = await import(join(fieldRoot, "src/signature.js"));
const { VM_ID, VM_VERSION, VM_RPCCHAINVM_PROTOCOL, VM_EFFECTS } = await import(join(fieldRoot, "src/constants.js"));

const [vmBinary, workDir, portArg] = process.argv.slice(2);
if (!vmBinary || !workDir) {
  console.error("usage: node scripts/coupled-lifecycle.mjs VM_BINARY WORK_DIR [PORT]");
  process.exit(2);
}
const port = Number(portArg ?? 19650);
const endpoint = `http://127.0.0.1:${port}`;

import { createHash } from "node:crypto";
const digest = (text) => createHash("sha256").update(text).digest("hex");

mkdirSync(workDir, { recursive: true, mode: 0o700 });
const transcript = [];
const record = (step, detail) => {
  transcript.push({ step, ...detail });
  console.log(`[coupled] ${step}: ${JSON.stringify(detail)}`);
};

// 1. Disposable locality authority + genesis materialization via the VM CLI.
execFileSync(vmBinary, ["--generate-authority", workDir], { stdio: "inherit" });
const descriptorPath = join(workDir, "descriptor.json");
writeFileSync(descriptorPath, JSON.stringify({
  schema: "PRESENCE_DEPLOYMENT_DESCRIPTOR_001",
  genesis_id: "PRESENCE-COUPLED-REHEARSAL-GENESIS-001",
  locality_id: "PRESENCE-LOCALITY-001",
  purpose: "Disposable coupled FIELD-VM lifecycle rehearsal for AVALANCHE_IMPLEMENTATION_001. No public network.",
  source_reference: "CONTINUITY",
  source_sha256: digest("presence coupled rehearsal continuity source")
}, null, 2));
const genesisPath = join(workDir, "genesis.json");
execFileSync(vmBinary, [
  "--materialize-genesis",
  join(implementationRoot, "genesis/PRESENCE_RUNTIME_GENESIS_TEMPLATE_001.json"),
  descriptorPath,
  join(workDir, "locality-authority.public.json"),
  genesisPath
], { stdio: "inherit" });

// 2. Serve the real VM in-process on loopback.
const server = spawn(vmBinary, ["--serve-rehearsal", genesisPath, `127.0.0.1:${port}`], { stdio: "inherit" });
const stopServer = () => { if (!server.killed) server.kill("SIGTERM"); };
process.on("exit", stopServer);

const client = new VMClient({ endpoint, request_timeout_ms: 5000, max_response_bytes: 4 * 1024 * 1024 });
const waitFor = async (fn, what, attempts = 50) => {
  for (let i = 0; i < attempts; i += 1) {
    try { return await fn(); } catch { await new Promise((r) => setTimeout(r, 200)); }
  }
  throw new Error(`timed out waiting for ${what}`);
};
const status0 = await waitFor(() => client.status(), "VM /status");
record("VM_SERVING", { vm_version: status0.vm_version, revision: status0.revision, locality: status0.locality_id });

try {
  // 3. Disposable FIELD medium actor; register it as a continuity authority
  //    through the host administrative path (draft → external sign → submit).
  const pair = generateKeyPairSync("ed25519");
  const publicKeyHex = pair.publicKey.export({ format: "der", type: "spki" }).subarray(-32).toString("hex");
  const actorId = actorIdFromPublicKey(publicKeyHex);
  const authority = JSON.parse(readFileSync(join(workDir, "locality-authority.public.json"), "utf8"));

  const registerDraft = await client.draft({
    operation: "REGISTER_CONTINUITY_AUTHORITY",
    actor_id: "PRESENCE-LOCALITY-001",
    actor_public_key: authority.public_key,
    locus_id: "",
    observed_at: 1,
    payload: {
      key_id: "COUPLED-FIELD-MEDIUM-001",
      public_key: publicKeyHex,
      capabilities: [
        "OPEN_LOCUS", "REGULATE_GATE", "RECORD_EMERGENCE", "CLOSE_LOCUS",
        "INCORPORATE_RESIDUE", "DECLARE_CAPACITY", "PULSE",
        "RECLAIM_ADMISSION_OFFER", "PROPOSE_SUCCESSOR"
      ],
      mandate_sha256: digest("coupled rehearsal continuity mandate")
    }
  });
  const unsignedPath = join(workDir, "register.unsigned.json");
  writeFileSync(unsignedPath, JSON.stringify(registerDraft.unsigned, null, 2));
  const signedPath = join(workDir, "register.signed.json");
  execFileSync(vmBinary, ["--sign-unsigned", unsignedPath, join(workDir, "locality-authority.private.json"), signedPath], { stdio: "inherit" });
  const signed = JSON.parse(readFileSync(signedPath, "utf8"));
  const registered = await client.submit(signed);
  const registerReceipt = await waitFor(async () => {
    const receipt = await client.receipt(registered.transition_id);
    if (!receipt) throw new Error("pending");
    return receipt;
  }, "REGISTER_CONTINUITY_AUTHORITY receipt");
  record("REGISTER_CONTINUITY_AUTHORITY_ACCEPTED", { transition_id: registerReceipt.transition_id, revision: registerReceipt.revision });

  // 4. Real FIELD engine for the registered medium actor (host profile).
  const journal = await new AppendOnlyJournal(join(workDir, "medium-events.ndjson")).initialize();
  const profile = await loadProfile(join(fieldRoot, "profiles/PRESENCE_FIELD_HOST_001.json"));
  const engine = new MediumEngine({
    config: {
      expected_locality_id: "PRESENCE-LOCALITY-001",
      expected_vm_id: VM_ID,
      expected_vm_version: VM_VERSION,
      expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
      actor_id: actorId,
      actor_public_key: publicKeyHex,
      receipt_timeout_ms: 30000
    },
    profile,
    client,
    journal
  });
  const signDraft = (draft) => sign(null, signingBytes(draft.unsigned), pair.privateKey).toString("hex");
  const runStep = async (operation, locusId, payload) => {
    const observation = await engine.observe();
    const draft = await engine.prepareDraft({
      operation,
      locus_id: locusId,
      observed_at: observation.state.last_observed_at + 1,
      payload
    });
    const submitted = await engine.submitTransition({ unsigned: draft.unsigned, signature: signDraft(draft) });
    const receipt = await engine.waitReceipt(submitted.transition_id);
    record(`${operation}_ACCEPTED`, { transition_id: receipt.transition_id, revision: receipt.revision, effect: receipt.effect, locus_id: draft.unsigned.locus_id });
    return receipt;
  };

  // 5. Coupled lifecycle: read → BOUND → renewed read.
  const readBefore = await engine.observe();
  record("READ", { revision: readBefore.state.revision, active_locus: readBefore.state.active_locus_id ?? "" });
  const locusId = "LOCUS-COUPLED-001";
  await runStep("BOUND", locusId, {
    locus_id: locusId,
    purpose_sha256: digest("coupled rehearsal locus purpose"),
    closure_condition_sha256: digest("coupled rehearsal closure condition"),
    capacity_ceiling_units: 500
  });
  const readAfterBound = await engine.observe();
  record("RENEWED_READ_AFTER_BOUND", { revision: readAfterBound.state.revision, active_locus: readAfterBound.state.active_locus_id });

  // 6. Beyond the former failure: DECLARE_CAPACITY with empty locus while the
  //    locus is active — the exact predecessor coupling defect, repaired.
  const declareReceipt = await runStep("DECLARE_CAPACITY", "", {
    actual_units: 800,
    resource_commitment_sha256: digest("coupled rehearsal resources"),
    basis_sha256: digest("coupled rehearsal capacity basis")
  });
  const renewedRead = await engine.observe();
  record("RENEWED_READ_AFTER_DECLARE_CAPACITY", {
    revision: renewedRead.state.revision,
    active_locus: renewedRead.state.active_locus_id,
    declared_actual_units: renewedRead.state.capacity.declared_actual_units
  });
  if (renewedRead.state.active_locus_id !== locusId) throw new Error("active locus lost after body-local capacity declaration");

  // 7. Continue the coupled lifecycle as far as this runner supports.
  const currentness = await client.request(`/currentness?observed_at=${renewedRead.state.last_observed_at + 1}&carrier_set_sha256=${digest("coupled carriers")}`);
  await runStep("PULSE", "", {
    currentness_commitment_sha256: currentness.next_pulse.currentness_commitment_sha256,
    carrier_set_sha256: digest("coupled carriers")
  });

  // TARGET_SCOPED: correct the body-local capacity declaration with an empty
  // locus even though LOCUS-COUPLED-001 is active.
  await runStep("CORRECT", "", {
    target_transition_id: declareReceipt.transition_id,
    replacement_commitment: digest("corrected coupled capacity basis"),
    reason_sha256: digest("coupled correction reason")
  });

  const readFinal = await engine.observe();
  record("FINAL_READ", {
    revision: readFinal.state.revision,
    active_locus: readFinal.state.active_locus_id,
    posture: readFinal.state.body.posture,
    events: Object.keys(readFinal.state.events).length
  });

  writeFileSync(join(workDir, "coupled-lifecycle-transcript.json"), JSON.stringify({
    schema: "PRESENCE_COUPLED_LIFECYCLE_TRANSCRIPT_001",
    endpoint,
    effects: VM_EFFECTS,
    steps: transcript
  }, null, 2));
  console.log("[coupled] lifecycle complete; transcript written");
} finally {
  stopServer();
}

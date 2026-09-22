import test from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { mkdtemp } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";

import {
  OPERATION_EFFECTS,
  OPERATION_SCOPES,
  OPERATION_SCOPE_CONTRACT,
  OPERATION_SCOPE_CONTRACT_PATH,
  scopeOf,
  validateDraftLocus
} from "../src/scope-contract.js";
import { VM_EFFECTS, VM_ID, VM_VERSION, VM_RPCCHAINVM_PROTOCOL } from "../src/constants.js";
import { MediumEngine } from "../src/engine.js";
import { loadProfile } from "../src/profile.js";
import { AppendOnlyJournal } from "../src/journal.js";
import { keyMaterial, presentState, statusFor, DIGEST_A } from "./helpers.js";

// The FIELD's native operation inventory. Requirement 9 asserts this agrees
// exactly with the shared contract, which the VM embeds byte-identically.
const FIELD_OPERATION_INVENTORY = [
  "BOUND",
  "PRESENT_FORM",
  "GATE_DISPOSITION",
  "ENTER",
  "CHECKPOINT_DEPARTURE",
  "REENTER",
  "OBSERVE_CROSSING",
  "MATTER_DISPOSITION",
  "RECORD_EMERGENCE",
  "CORRECT",
  "EXIT",
  "CLOSE",
  "INCORPORATE_ADDRESSED_RESIDUE",
  "DECLARE_CAPACITY",
  "PULSE",
  "RECLAIM_ADMISSION_OFFER",
  "REGISTER_CONTINUITY_AUTHORITY",
  "EXHAUST_FORMATION_AUTHORITY",
  "PROPOSE_SUCCESSOR",
  "ATTEST_SUCCESSOR",
  "ACTIVATE_SUCCESSOR"
];

const ACTIVE = "LOCUS-001";

function stateWithActiveLocus(extra = {}) {
  const keys = keyMaterial();
  const state = presentState(keys.actorId, keys.publicKeyHex, extra);
  return { keys, state };
}

// Requirement 1: BOUND succeeds with a new locus ID.
test("scope: BOUND accepts a new locus ID and rejects impersonating the active locus", () => {
  const { state } = stateWithActiveLocus();
  assert.doesNotThrow(() => validateDraftLocus("BOUND", "LOCUS-NEW-001", state, {}));
  assert.throws(() => validateDraftLocus("BOUND", ACTIVE, state, {}), /already active locus/);
  assert.throws(() => validateDraftLocus("BOUND", "", state, {}), /proposed new locus/);
});

// Requirement 2: DECLARE_CAPACITY succeeds with empty locus_id while a locus
// is active. This was the exact predecessor coupling failure.
test("scope: DECLARE_CAPACITY accepts an empty locus while a locus is active", () => {
  const { state } = stateWithActiveLocus();
  assert.equal(state.active_locus_id, ACTIVE);
  assert.doesNotThrow(() => validateDraftLocus("DECLARE_CAPACITY", "", state, {}));
});

// Requirement 3: DECLARE_CAPACITY with the active locus ID is rejected before
// signing/submission (prepareDraft throws before any VM draft or signature).
test("scope: DECLARE_CAPACITY with the active locus is rejected before signing", async () => {
  const { keys, state } = stateWithActiveLocus();
  assert.throws(() => validateDraftLocus("DECLARE_CAPACITY", ACTIVE, state, {}), /empty locus_id/);

  let vmDraftCalls = 0;
  const client = {
    status: async () => statusFor(state),
    state: async () => state,
    draft: async (request) => {
      vmDraftCalls += 1;
      return {
        unsigned: {
          schema: "PRESENCE_TRANSITION_001",
          operation: request.operation,
          revision: state.revision + 1,
          previous_state_commitment: state.state_commitment,
          actor_id: request.actor_id,
          actor_public_key: request.actor_public_key,
          nonce: state.next_nonces[keys.actorId] ?? 0,
          locus_id: request.locus_id,
          observed_at: request.observed_at,
          effect: VM_EFFECTS[request.operation],
          payload: request.payload
        }
      };
    },
    submit: async () => { throw new Error("must not submit"); }
  };
  const directory = await mkdtemp(join(tmpdir(), "presence-field-scope-"));
  const journal = await new AppendOnlyJournal(join(directory, "events.ndjson")).initialize();
  const profile = await loadProfile(resolve("profiles/PRESENCE_FIELD_HOST_001.json"));
  const engine = new MediumEngine({
    config: {
      expected_locality_id: state.host_locality_id,
      expected_vm_id: VM_ID,
      expected_vm_version: VM_VERSION,
      expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
      actor_id: keys.actorId,
      actor_public_key: keys.publicKeyHex,
      receipt_timeout_ms: 1000
    },
    profile,
    client,
    journal
  });

  await assert.rejects(
    () => engine.prepareDraft({
      operation: "DECLARE_CAPACITY",
      locus_id: ACTIVE,
      observed_at: 1001,
      payload: { actual_units: 90, resource_commitment_sha256: DIGEST_A, basis_sha256: DIGEST_A }
    }),
    /empty locus_id/
  );
  assert.equal(vmDraftCalls, 0, "rejected draft must never reach the VM");

  const draft = await engine.prepareDraft({
    operation: "DECLARE_CAPACITY",
    locus_id: "",
    observed_at: 1001,
    payload: { actual_units: 90, resource_commitment_sha256: DIGEST_A, basis_sha256: DIGEST_A }
  });
  assert.equal(draft.unsigned.operation, "DECLARE_CAPACITY");
  assert.equal(draft.unsigned.locus_id, "");
  assert.equal(vmDraftCalls, 1);
});

// Requirement 4: every ACTIVE_LOCUS operation rejects an empty or different
// locus.
test("scope: every ACTIVE_LOCUS operation rejects empty and foreign loci", () => {
  const { state } = stateWithActiveLocus();
  for (const [operation, scope] of Object.entries(OPERATION_SCOPES)) {
    if (scope !== "ACTIVE_LOCUS") continue;
    assert.throws(() => validateDraftLocus(operation, "", state, {}), /exact accepted active locus|requires an active locus/, operation);
    assert.throws(() => validateDraftLocus(operation, "LOCUS-FOREIGN-001", state, {}), /exact accepted active locus/, operation);
    assert.doesNotThrow(() => validateDraftLocus(operation, ACTIVE, state, {}), operation);
  }
  const dormant = { ...state, active_locus_id: "" };
  for (const [operation, scope] of Object.entries(OPERATION_SCOPES)) {
    if (scope !== "ACTIVE_LOCUS") continue;
    assert.throws(() => validateDraftLocus(operation, ACTIVE, dormant, {}), /requires an active locus/, operation);
  }
});

// Requirement 5: every BODY_LOCAL operation rejects a nonempty locus.
test("scope: every BODY_LOCAL operation rejects a nonempty locus", () => {
  const { state } = stateWithActiveLocus();
  for (const [operation, scope] of Object.entries(OPERATION_SCOPES)) {
    if (scope !== "BODY_LOCAL") continue;
    assert.throws(() => validateDraftLocus(operation, ACTIVE, state, {}), /empty locus_id/, operation);
    assert.doesNotThrow(() => validateDraftLocus(operation, "", state, {}), operation);
  }
});

// Requirement 6: dormant-body operations reject an active locus.
test("scope: DORMANT_BODY operations reject an active locus and nonempty loci", () => {
  const { state } = stateWithActiveLocus();
  const dormant = { ...state, active_locus_id: "" };
  for (const [operation, scope] of Object.entries(OPERATION_SCOPES)) {
    if (scope !== "DORMANT_BODY") continue;
    assert.throws(() => validateDraftLocus(operation, "", state, {}), /requires no active locus/, operation);
    assert.throws(() => validateDraftLocus(operation, ACTIVE, dormant, {}), /empty locus_id/, operation);
    assert.doesNotThrow(() => validateDraftLocus(operation, "", dormant, {}), operation);
  }
});

// Requirement 7: CORRECT binds its target transition's scope rather than
// incidental current activity.
test("scope: CORRECT binds the exact scope of its target transition", () => {
  const { state } = stateWithActiveLocus();
  const bodyLocalTarget = "f".repeat(64);
  const locusTarget = "9".repeat(64);
  state.events[bodyLocalTarget] = { operation: "DECLARE_CAPACITY", locus_id: "", actor_id: "X" };
  state.events[locusTarget] = { operation: "OBSERVE_CROSSING", locus_id: ACTIVE, actor_id: "X" };

  // Body-local target: empty locus is required even though a locus is active.
  assert.doesNotThrow(() => validateDraftLocus("CORRECT", "", state, { target_transition_id: bodyLocalTarget }));
  assert.throws(
    () => validateDraftLocus("CORRECT", ACTIVE, state, { target_transition_id: bodyLocalTarget }),
    /exact scope of its target/
  );

  // Locus-scoped target: the exact target locus is required.
  assert.doesNotThrow(() => validateDraftLocus("CORRECT", ACTIVE, state, { target_transition_id: locusTarget }));
  assert.throws(
    () => validateDraftLocus("CORRECT", "", state, { target_transition_id: locusTarget }),
    /exact scope of its target/
  );

  // Unknown target fails closed.
  assert.throws(
    () => validateDraftLocus("CORRECT", "", state, { target_transition_id: "0".repeat(64) }),
    /failing closed/
  );
  assert.throws(() => validateDraftLocus("CORRECT", "", state, {}), /target_transition_id/);
});

// Requirement 8: every declared operation appears exactly once in the scope
// contract, at the byte level of the shared file.
test("scope: every operation appears exactly once in the contract file", () => {
  const raw = readFileSync(OPERATION_SCOPE_CONTRACT_PATH, "utf8");
  const start = raw.indexOf('"operations"');
  assert.ok(start > 0, "contract file declares an operations table");
  const operationsBlock = raw.slice(start);
  for (const operation of FIELD_OPERATION_INVENTORY) {
    const occurrences = operationsBlock.split(`"${operation}":`).length - 1;
    assert.equal(occurrences, 1, `${operation} must appear exactly once, found ${occurrences}`);
  }
  assert.equal(Object.keys(OPERATION_SCOPE_CONTRACT.operations).length, FIELD_OPERATION_INVENTORY.length);
});

// Requirement 9: FIELD and VM operation/effect/scope inventories agree. Both
// runtimes load the identical contract file; this asserts the FIELD inventory
// against it, and binds the exact file bytes into the field contract.
test("scope: FIELD inventory, effects and scopes agree with the shared contract", () => {
  assert.deepEqual(Object.keys(OPERATION_EFFECTS).sort(), [...FIELD_OPERATION_INVENTORY].sort());
  assert.deepEqual(Object.keys(OPERATION_SCOPES).sort(), [...FIELD_OPERATION_INVENTORY].sort());
  assert.equal(VM_EFFECTS, OPERATION_EFFECTS, "VM_EFFECTS must be the contract effects object itself");
  for (const operation of FIELD_OPERATION_INVENTORY) {
    assert.equal(typeof OPERATION_EFFECTS[operation], "string");
    assert.ok(OPERATION_EFFECTS[operation].length > 0);
    assert.ok(["NEW_LOCUS", "ACTIVE_LOCUS", "BODY_LOCAL", "DORMANT_BODY", "TARGET_SCOPED"].includes(scopeOf(operation)), operation);
  }
  const raw = readFileSync(OPERATION_SCOPE_CONTRACT_PATH);
  const digest = createHash("sha256").update(raw).digest("hex");
  const fieldContract = JSON.parse(readFileSync(new URL("../contract/PRESENCE_FIELD_CONTRACT_001.json", import.meta.url), "utf8"));
  assert.equal(
    fieldContract.operation_scope_contract.file_sha256,
    digest,
    "field contract must bind the exact operation-scope contract bytes"
  );
});

// Requirement 10: unknown operations fail closed.
test("scope: unknown operations fail closed", () => {
  const { state } = stateWithActiveLocus();
  assert.throws(() => scopeOf("NOT_A_DECLARED_OPERATION"), /failing closed/);
  assert.throws(() => validateDraftLocus("NOT_A_DECLARED_OPERATION", "", state, {}), /failing closed/);
  assert.equal(OPERATION_SCOPES.NOT_A_DECLARED_OPERATION, undefined);
});

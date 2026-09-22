import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { sign, createHash } from "node:crypto";
import { MediumEngine } from "../src/engine.js";
import { loadProfile, validateProfile } from "../src/profile.js";
import { signingBytes, transitionID } from "../src/signature.js";
import { SUPPORTED_VM_OPERATIONS, VM_OPERATIONS, validateOperationContract } from "../src/operations.js";
import { VM_EFFECTS, VM_FORM_SHA256, VM_ID, VM_OPERATION_CONTRACT_SHA256, VM_RPCCHAINVM_PROTOCOL, VM_TRANSITION_SCHEMA, VM_VERSION } from "../src/constants.js";
import { keyMaterial, presentState, statusFor } from "./helpers.js";

const contractURL = new URL("../../vm/protocol/operations.json", import.meta.url);
const contract = JSON.parse(await readFile(contractURL, "utf8"));
const referenceProfile = await loadProfile(new URL("../profiles/LOCALITY_FIELD_001.json", import.meta.url));
const scopeProfile = validateProfile({
  ...referenceProfile,
  capabilities_by_phase: Object.fromEntries(Object.entries(referenceProfile.capabilities_by_phase).map(
    ([phase, capabilities]) => [phase, [...new Set([...capabilities, ...SUPPORTED_VM_OPERATIONS])]]
  ))
});

function fixture({ active = true, profile = scopeProfile } = {}) {
  const keys = keyMaterial();
  const state = presentState(keys.actorId, keys.publicKeyHex);
  const locus = state.loci[state.active_locus_id];
  locus.presentations = {};
  locus.entries = {};
  state.body = { posture: "DORMANT_P0", presence_count: 0, reason: "NO_CURRENT_PARTICIPANT_BINDING" };
  if (!active) state.active_locus_id = "";
  const calls = { draft: 0, submit: 0 };
  const client = {
    state: async () => structuredClone(state),
    status: async () => statusFor(state),
    draft: async (request) => {
      calls.draft += 1;
      return { unsigned: {
        schema: VM_TRANSITION_SCHEMA, operation: request.operation, revision: state.revision + 1,
        previous_state_commitment: state.state_commitment, actor_id: request.actor_id,
        actor_public_key: request.actor_public_key, nonce: state.next_nonces[request.actor_id],
        locus_id: request.locus_id, observed_at: request.observed_at,
        effect: VM_EFFECTS[request.operation], payload: request.payload
      } };
    },
    submit: async (transition) => {
      calls.submit += 1;
      return { transition_id: transitionID(transition), status: "PENDING_CONSENSUS" };
    }
  };
  const config = {
    expected_locality_id: state.host_locality_id, expected_vm_id: VM_ID,
    expected_vm_version: VM_VERSION, expected_rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
    actor_id: keys.actorId, actor_public_key: keys.publicKeyHex
  };
  const engine = new MediumEngine({ config, profile, client, journal: { append: async () => {} } });
  return { engine, state, keys, calls, client };
}

function request(operation, locus_id, payload = {}) {
  return { operation, locus_id, observed_at: 1001, payload };
}

async function refusedBeforeDraft(f, input, message) {
  const before = f.calls.draft;
  const pending = f.engine.pendingDrafts.size;
  await assert.rejects(() => f.engine.prepareDraft(input), message);
  assert.equal(f.calls.draft, before, "invalid scope must not reach VM drafting or external signing");
  assert.equal(f.engine.pendingDrafts.size, pending);
  assert.equal(f.calls.submit, 0);
}

test("FIELD inventory equals shared contract and actual VM supported operation inventory", async () => {
  const protocol = await readFile(new URL("../../vm/protocol.go", import.meta.url), "utf8");
  const transitions = await readFile(new URL("../../vm/transition.go", import.meta.url), "utf8");
  const declarations = new Map([...transitions.matchAll(/\b(op\w+)\s*=\s*"([A-Z][A-Z0-9_]+)"/g)].map((m) => [m[1], m[2]]));
  const inventory = protocol.match(/var supportedOperations = \[\]string\{([\s\S]*?)\n\}/);
  assert.ok(inventory, "VM supported-operation inventory must remain mechanically readable");
  const vmOperations = [...inventory[1].matchAll(/\bop[A-Z]\w*/g)].map(([symbol]) => {
    assert.ok(declarations.has(symbol), `unresolved VM operation ${symbol}`);
    return declarations.get(symbol);
  });
  const expected = [...SUPPORTED_VM_OPERATIONS].sort();
  assert.deepEqual(vmOperations.sort(), expected);
  assert.deepEqual(Object.keys(VM_OPERATIONS).sort(), expected);
  assert.deepEqual(Object.keys(VM_EFFECTS).sort(), expected);
  assert.equal(createHash("sha256").update(await readFile(contractURL)).digest("hex"), VM_OPERATION_CONTRACT_SHA256);
  for (const row of contract.operations) {
    assert.deepEqual(VM_OPERATIONS[row.operation], row);
    assert.equal(VM_EFFECTS[row.operation], row.effect);
  }
  const form = await readFile(new URL("../../vm/profile/PRESENCE_AVALANCHE_FORM_001.json", import.meta.url));
  assert.equal(createHash("sha256").update(form).digest("hex"), VM_FORM_SHA256);
});

test("shared operation inventory fails closed for every omitted, duplicate or unclassified operation", () => {
  for (const row of contract.operations) {
    assert.throws(() => validateOperationContract({
      ...contract, operations: contract.operations.filter((entry) => entry.operation !== row.operation)
    }), /missing supported/);
    assert.throws(() => validateOperationContract({
      ...contract, operations: [...contract.operations, row]
    }), /duplicate/);
    for (const scope of [undefined, "", "UNKNOWN_SCOPE"]) {
      assert.throws(() => validateOperationContract({
        ...contract, operations: contract.operations.map((entry) => entry === row ? { ...entry, scope } : entry)
      }), /unclassified/);
    }
    for (const effect of [undefined, "", "invalid effect"]) {
      assert.throws(() => validateOperationContract({
        ...contract, operations: contract.operations.map((entry) => entry === row ? { ...entry, effect } : entry)
      }), /missing VM effect/);
    }
  }
  assert.throws(() => validateOperationContract({
    ...contract, operations: [...contract.operations, { operation: "UNKNOWN", effect: "UNKNOWN", scope: "BODY_LOCAL" }]
  }), /unknown VM operation/);
  assert.throws(() => validateOperationContract({ ...contract, version: "2.0.0" }), /unsupported/);
  assert.throws(() => validateOperationContract({ ...contract, operations: null }), /unsupported/);
  assert.throws(() => validateOperationContract({ ...contract, operations: [null] }), /invalid/);
});

const activeOperations = [
  "PRESENT_FORM", "GATE_DISPOSITION", "ENTER", "CHECKPOINT_DEPARTURE", "REENTER",
  "OBSERVE_CROSSING", "MATTER_DISPOSITION", "RECORD_EMERGENCE", "EXIT", "CLOSE",
  "RECLAIM_ADMISSION_OFFER"
];

for (const operation of activeOperations) {
  test(`${operation}: ACTIVE_LOCUS requires exact nonempty accepted active locus before draft`, async () => {
    assert.equal(VM_OPERATIONS[operation].scope, "ACTIVE_LOCUS");
    const f = fixture();
    const active = f.state.active_locus_id;
    assert.equal((await f.engine.prepareDraft(request(operation, active))).unsigned.locus_id, active);
    for (const wrong of ["", "DIFFERENT-LOCUS", null]) {
      await refusedBeforeDraft(f, request(operation, wrong), /ACTIVE_LOCUS|must be a string/);
    }
    const absent = fixture({ active: false });
    for (const wrong of ["", active]) {
      await refusedBeforeDraft(absent, request(operation, wrong), /ACTIVE_LOCUS/);
    }
  });
}

for (const operation of [
  "INCORPORATE_ADDRESSED_RESIDUE", "DECLARE_CAPACITY", "PULSE",
  "REGISTER_CONTINUITY_AUTHORITY", "PROPOSE_SUCCESSOR", "ATTEST_SUCCESSOR"
]) {
  test(`${operation}: BODY_LOCAL scope uses empty locus with or without active locus`, async () => {
    assert.equal(VM_OPERATIONS[operation].scope, "BODY_LOCAL");
    for (const active of [true, false]) {
      const f = fixture({ active });
      assert.equal((await f.engine.prepareDraft(request(operation, ""))).unsigned.locus_id, "");
      for (const wrong of ["LOCUS-001", "DIFFERENT-LOCUS", null]) {
        await refusedBeforeDraft(f, request(operation, wrong), /BODY_LOCAL|must be a string/);
      }
    }
  });
}

for (const operation of ["EXHAUST_FORMATION_AUTHORITY", "ACTIVATE_SUCCESSOR"]) {
  test(`${operation}: DORMANT_BODY requires empty locus, no active locus and exact dormancy`, async () => {
    assert.equal(VM_OPERATIONS[operation].scope, "DORMANT_BODY");
    const f = fixture({ active: false });
    assert.equal((await f.engine.prepareDraft(request(operation, ""))).unsigned.locus_id, "");
    await refusedBeforeDraft(f, request(operation, "LOCUS-001"), /DORMANT_BODY/);
    const active = fixture();
    await refusedBeforeDraft(active, request(operation, ""), /DORMANT_BODY/);
    for (const posture of ["DORMANT", "OPEN_PRESENCE", "HOLD_CAPACITY_DEFICIT"]) {
      f.state.body.posture = posture;
      await refusedBeforeDraft(f, request(operation, ""), /DORMANT_BODY/);
    }
    f.state.body.posture = "DORMANT_P0";
    f.state.body.presence_count = 1;
    await refusedBeforeDraft(f, request(operation, ""), /DORMANT_BODY/);
  });
}

test("BOUND: NEW_LOCUS binds unused proposed locus and matching payload, not accepted active locus", async () => {
  assert.equal(VM_OPERATIONS.BOUND.scope, "NEW_LOCUS");
  const f = fixture({ active: false });
  const proposed = request("BOUND", "LOCUS-NEW", { locus_id: "LOCUS-NEW" });
  assert.equal((await f.engine.prepareDraft(proposed)).unsigned.locus_id, "LOCUS-NEW");
  await refusedBeforeDraft(f, request("BOUND", "", { locus_id: "" }), /NEW_LOCUS/);
  await refusedBeforeDraft(f, request("BOUND", "LOCUS-001", { locus_id: "LOCUS-001" }), /NEW_LOCUS/);
  await refusedBeforeDraft(f, request("BOUND", "LOCUS-NEW", { locus_id: "OTHER" }), /NEW_LOCUS/);
  const active = fixture();
  await refusedBeforeDraft(active, request("BOUND", "LOCUS-001", { locus_id: "LOCUS-001" }), /NEW_LOCUS/);
  await refusedBeforeDraft(active, proposed, /NEW_LOCUS/);
});

test("CORRECT: TARGET_SCOPED binds accepted event locus, including body-local and closed-locus targets", async () => {
  assert.equal(VM_OPERATIONS.CORRECT.scope, "TARGET_SCOPED");
  for (const active of [true, false]) {
    const f = fixture({ active });
    for (const targetLocus of ["LOCUS-CLOSED", "LOCUS-001", ""]) {
      const target = "f".repeat(64);
      f.state.events[target] = { transition_id: target, locus_id: targetLocus };
      const payload = { target_transition_id: target };
      assert.equal((await f.engine.prepareDraft(request("CORRECT", targetLocus, payload))).unsigned.locus_id, targetLocus);
      await refusedBeforeDraft(f, request("CORRECT", targetLocus === "LOCUS-001" ? "" : "LOCUS-001", payload), /TARGET_SCOPED/);
    }
    await refusedBeforeDraft(f, request("CORRECT", "", { target_transition_id: "0".repeat(64) }), /TARGET_SCOPED/);
    await refusedBeforeDraft(f, request("CORRECT", ""), /TARGET_SCOPED/);
    await refusedBeforeDraft(f, request("CORRECT", "", { target_transition_id: "toString" }), /TARGET_SCOPED/);
  }
});

test("unknown operations are refused before requesting a draft", async () => {
  const f = fixture();
  await refusedBeforeDraft(f, request("UNKNOWN", ""), /denied/);
  await refusedBeforeDraft(f, request("READ_PUBLIC_STATE", ""), /unknown VM operation/);
});

test("all accepted-state-dependent scopes are checked again before signed submission", async () => {
  const cases = [
    { operation: "BOUND", active: false, locus: "NEW", payload: { locus_id: "NEW" }, mutate: (s) => { s.active_locus_id = "LOCUS-001"; }, error: /NEW_LOCUS/ },
    { operation: "OBSERVE_CROSSING", active: true, locus: "LOCUS-001", mutate: (s) => { s.active_locus_id = ""; }, error: /ACTIVE_LOCUS/ },
    { operation: "EXHAUST_FORMATION_AUTHORITY", active: false, locus: "", mutate: (s) => { s.body.posture = "HOLD_CAPACITY_DEFICIT"; }, error: /DORMANT_BODY/ },
    { operation: "CORRECT", active: true, locus: "OLD", payload: { target_transition_id: "f".repeat(64) }, mutate: (s) => { s.events["f".repeat(64)].locus_id = "OTHER"; }, error: /TARGET_SCOPED/ }
  ];
  for (const entry of cases) {
    const f = fixture({ active: entry.active });
    f.state.events["f".repeat(64)] = { locus_id: "OLD" };
    const draft = await f.engine.prepareDraft(request(entry.operation, entry.locus, entry.payload));
    const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
    // A dishonest endpoint can change represented scope while retaining its claimed commitment.
    entry.mutate(f.state);
    await assert.rejects(() => f.engine.submitTransition({ unsigned: draft.unsigned, signature }), entry.error);
    assert.equal(f.calls.submit, 0);
  }
});

test("real locality profile supports BOUND then empty-locus DECLARE_CAPACITY while authority is unpresented", async () => {
  const f = fixture({ active: false, profile: referenceProfile });
  f.state.authorities[f.keys.actorId] = { actor_id: f.keys.actorId, public_key: f.keys.publicKeyHex, status: "ACTIVE" };
  const bound = await f.engine.prepareDraft(request("BOUND", "NEW", { locus_id: "NEW" }));
  assert.equal(bound.unsigned.operation, "BOUND");
  f.state.loci.NEW = { ...f.state.loci["LOCUS-001"], locus_id: "NEW" };
  f.state.active_locus_id = "NEW";
  f.state.revision += 1;
  f.state.state_commitment = "b".repeat(64);
  assert.equal((await f.engine.observe()).observation.phase, "UNPRESENTED");
  const draft = await f.engine.prepareDraft(request("DECLARE_CAPACITY", "", { declared_actual_units: 8 }));
  const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
  assert.equal((await f.engine.submitTransition({ unsigned: draft.unsigned, signature })).status, "PENDING_CONSENSUS");
  assert.equal(f.calls.submit, 1);
  await refusedBeforeDraft(fixture({ profile: referenceProfile }), request("DECLARE_CAPACITY", "LOCUS-001"), /BODY_LOCAL/);
});

test("scope support does not grant unrelated authority through participant profiles", async () => {
  for (const kind of ["AI", "HUMAN", "DEVICE"]) {
    const profile = await loadProfile(new URL(`../profiles/${kind}_FIELD_001.json`, import.meta.url));
    const f = fixture({ profile });
    for (const operation of ["DECLARE_CAPACITY", "PULSE", "REGISTER_CONTINUITY_AUTHORITY", "PROPOSE_SUCCESSOR", "ATTEST_SUCCESSOR", "ACTIVATE_SUCCESSOR"]) {
      await refusedBeforeDraft(f, request(operation, ""), /denied/);
    }
  }
});

function useHostActor(f) {
  const host = f.state.host_locality_id;
  f.engine.config.actor_id = host;
  f.state.actor_keys[host] = f.keys.publicKeyHex;
  f.state.next_nonces[host] = f.state.next_nonces[f.keys.actorId];
  delete f.state.actor_keys[f.keys.actorId];
  delete f.state.next_nonces[f.keys.actorId];
  return host;
}

test("exact accepted genesis host identity can sign BOUND and live unpresented capacity drafts", async () => {
  const f = fixture({ active: false, profile: referenceProfile });
  const host = useHostActor(f);
  for (const input of [
    request("BOUND", "NEW", { locus_id: "NEW" }),
    request("DECLARE_CAPACITY", "", { declared_actual_units: 8 })
  ]) {
    const draft = await f.engine.prepareDraft(input);
    assert.equal(draft.unsigned.actor_id, host);
    const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
    assert.equal((await f.engine.submitTransition({ unsigned: draft.unsigned, signature })).status, "PENDING_CONSENSUS");
    f.state.loci.NEW = { ...f.state.loci["LOCUS-001"], locus_id: "NEW" };
    f.state.active_locus_id = "NEW";
    f.state.revision += 1;
    f.state.state_commitment = "b".repeat(64);
  }
  assert.equal(f.calls.submit, 2);
});

test("a host alias requires exact accepted key binding even for first presentation", async () => {
  for (const binding of [undefined, "0".repeat(64)]) {
    const f = fixture({ profile: referenceProfile });
    const host = useHostActor(f);
    if (binding === undefined) delete f.state.actor_keys[host];
    else f.state.actor_keys[host] = binding;
    await refusedBeforeDraft(f, request("PRESENT_FORM", f.state.active_locus_id), /EXACT_ACCEPTED_HOST|KEY_BINDING_MISMATCH/);
  }
  const other = fixture({ profile: referenceProfile });
  useHostActor(other);
  other.engine.config.actor_id = "OTHER-ACCEPTED-ALIAS";
  other.state.actor_keys[other.engine.config.actor_id] = other.keys.publicKeyHex;
  await refusedBeforeDraft(other, request("PRESENT_FORM", other.state.active_locus_id), /EXACT_ACCEPTED_HOST/);
});

test("host binding is rechecked against accepted state before signed submission", async () => {
  const f = fixture({ profile: referenceProfile });
  const host = useHostActor(f);
  const draft = await f.engine.prepareDraft(request("DECLARE_CAPACITY", "", { declared_actual_units: 8 }));
  const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
  delete f.state.actor_keys[host];
  await assert.rejects(() => f.engine.submitTransition({ unsigned: draft.unsigned, signature }), /EXACT_ACCEPTED_HOST|UNKNOWN_ACTOR/);
  assert.equal(f.calls.submit, 0);
});

test("missing or stale embedded VM contract hash prevents drafting at the same VM version", async () => {
  for (const hash of [undefined, "", "0".repeat(64)]) {
    const f = fixture();
    f.client.status = async () => statusFor(f.state, { operation_contract_sha256: hash });
    await refusedBeforeDraft(f, request("OBSERVE_CROSSING", f.state.active_locus_id), /operation contract hash mismatch/);
  }
});

test("VM operation contract hash is rechecked before signed submission", async () => {
  const f = fixture();
  const draft = await f.engine.prepareDraft(request("OBSERVE_CROSSING", f.state.active_locus_id));
  const signature = sign(null, signingBytes(draft.unsigned), f.keys.privateKey).toString("hex");
  f.client.status = async () => statusFor(f.state, { operation_contract_sha256: "0".repeat(64) });
  await assert.rejects(() => f.engine.submitTransition({ unsigned: draft.unsigned, signature }), /operation contract hash mismatch/);
  assert.equal(f.calls.submit, 0);
});

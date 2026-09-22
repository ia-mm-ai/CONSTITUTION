import { generateKeyPairSync } from "node:crypto";
import {
  REQUIRED_EFFECT_CEILING,
  VM_FORM_ID,
  VM_FORM_SHA256,
  VM_REQUIRED_MEDIUM_CAPABILITIES,
  VM_STATE_SCHEMA,
  VM_VERSION,
  VM_RPCCHAINVM_PROTOCOL
} from "../src/constants.js";
import { actorIdFromPublicKey } from "../src/config.js";

export const DIGEST_A = "a".repeat(64);
export const DIGEST_B = "b".repeat(64);

export function keyMaterial() {
  const pair = generateKeyPairSync("ed25519");
  const der = pair.publicKey.export({ format: "der", type: "spki" });
  const publicKeyHex = der.subarray(-32).toString("hex");
  return { ...pair, publicKeyHex, actorId: actorIdFromPublicKey(publicKeyHex) };
}

export function presentState(actorId, publicKeyHex, overrides = {}) {
  const presentationId = "c".repeat(64);
  const entryId = "d".repeat(64);
  const locusId = "LOCUS-001";
  const state = {
    schema: VM_STATE_SCHEMA,
    revision: 7,
    state_commitment: DIGEST_A,
    last_transition_id: "e".repeat(64),
    last_observed_at: 1000,
    host_locality_id: "PRESENCE-LOCALITY-001",
    profile_form_id: VM_FORM_ID,
    profile_form_sha256: VM_FORM_SHA256,
    active_locus_id: locusId,
    body: { posture: "OPEN_PRESENCE", presence_count: 1, reason: "ONE_PRESENT" },
    capacity: {},
    authorities: {},
    formation_authority_id: "PRESENCE-LOCALITY-001",
    pulses: {},
    last_pulse_id: "",
    successor_proposals: {},
    successor_boundary: null,
    actor_keys: { [actorId]: publicKeyHex },
    next_nonces: { [actorId]: 3 },
    presentation_history: {},
    entry_history: {},
    departure_checkpoints: {},
    residue_history: {},
    loci: {},
    events: {},
    corrections: {},
    incorporated_residues: {}
  };
  const presentation = {
    presentation_id: presentationId,
    participant_id: actorId,
    locality_reference: "participant://fixture",
    source_reference: "fixture",
    source_sha256: "1".repeat(64),
    nucleus_version: "1",
    nucleus_sha256: "2".repeat(64),
    form_id: VM_FORM_ID,
    form_sha256: VM_FORM_SHA256,
    state_commitment: "3".repeat(64),
    medium_capabilities: [...VM_REQUIRED_MEDIUM_CAPABILITIES],
    effect_ceiling: [...REQUIRED_EFFECT_CEILING]
  };
  const entry = {
    participant_id: actorId,
    presentation_id: presentationId,
    entry_transition_id: entryId,
    ingress_mode: "FIRST_ENTRY",
    prior_entry_transition_id: "",
    prior_departure_checkpoint_id: "",
    prior_residue_id: "",
    work_units: 1,
    resolution_units: 1,
    departure_checkpoint_id: "",
    status: "PRESENT",
    end_transition_id: "",
    end_reason_sha256: ""
  };
  state.presentation_history[presentationId] = presentation;
  state.entry_history[entryId] = entry;
  state.loci[locusId] = {
    locus_id: locusId,
    phase: "OPEN",
    presentations: { [actorId]: presentation },
    gates: {},
    entries: { [actorId]: entry },
    matter: {},
    emergence: {},
    residues: {}
  };
  return Object.assign(state, overrides);
}

export function statusFor(state, overrides = {}) {
  return {
    healthy: true,
    vm_version: VM_VERSION,
    rpcchainvm_protocol: VM_RPCCHAINVM_PROTOCOL,
    avalanchego_target: "v1.15.0",
    locality_id: state.host_locality_id,
    height: 22,
    revision: state.revision,
    state_commitment: state.state_commitment,
    active_locus_id: state.active_locus_id,
    body: state.body,
    capacity: { available_units: 8, deficit_units: 0 },
    pending_transitions: 0,
    successor_boundary: state.successor_boundary,
    ...overrides
  };
}

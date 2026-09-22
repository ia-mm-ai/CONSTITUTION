import {
  MEDIUM_OBSERVATION_SCHEMA,
  VM_FORM_ID,
  VM_FORM_SHA256,
  VM_STATE_SCHEMA
} from "./constants.js";
import { assertPlainObject, isDigest } from "./canonical.js";

function requireString(value, label) {
  if (typeof value !== "string") throw new Error(`${label} must be a string`);
}

function requireSafeUint(value, label) {
  if (!Number.isSafeInteger(value) || value < 0) throw new Error(`${label} is invalid or exceeds the exact integer boundary`);
}

export function validateVMState(state) {
  assertPlainObject(state, "VM state");
  if (state.schema !== VM_STATE_SCHEMA) throw new Error(`unsupported VM state schema ${state.schema}`);
  requireSafeUint(state.revision, "VM revision");
  if (!isDigest(state.state_commitment)) throw new Error("VM state commitment is invalid");
  requireString(state.host_locality_id, "host_locality_id");
  requireString(state.active_locus_id, "active_locus_id");
  if (state.profile_form_id !== VM_FORM_ID || state.profile_form_sha256 !== VM_FORM_SHA256) {
    throw new Error("VM form identity does not match PRESENCE_AVALANCHE_VM_001 v1.0.0");
  }
  assertPlainObject(state.body, "body");
  requireString(state.body.posture, "body.posture");
  requireSafeUint(state.body.presence_count, "body.presence_count");
  assertPlainObject(state.capacity, "capacity");
  for (const field of ["structural_ceiling_units", "declared_actual_units", "correction_egress_reserve_units", "last_declaration_observed_at"]) {
    if (Object.hasOwn(state.capacity, field)) requireSafeUint(state.capacity[field], `capacity.${field}`);
  }
  assertPlainObject(state.presentation_history, "presentation_history");
  assertPlainObject(state.entry_history, "entry_history");
  assertPlainObject(state.departure_checkpoints, "departure_checkpoints");
  assertPlainObject(state.residue_history, "residue_history");
  assertPlainObject(state.loci, "loci");
  assertPlainObject(state.actor_keys, "actor_keys");
  assertPlainObject(state.next_nonces, "next_nonces");
  for (const [actorId, nonce] of Object.entries(state.next_nonces)) {
    requireSafeUint(nonce, `next_nonces.${actorId}`);
  }
  return state;
}

function findSealedCheckpoint(state, actorId) {
  const matches = Object.values(state.departure_checkpoints).filter((checkpoint) =>
    checkpoint.participant_id === actorId && checkpoint.status === "SEALED" && !checkpoint.consumed_by_entry_transition_id
  );
  if (matches.length > 1) throw new Error("actor has multiple unconsumed sealed checkpoints; refusing ambiguous re-entry");
  return matches[0] ?? null;
}

function contextForNoLocus(state, actorId) {
  const checkpoint = findSealedCheckpoint(state, actorId);
  return {
    phase: state.body.posture === "SUCCESSION_COMMITTED" ? "SUCCESSION_COMMITTED" : "NO_ACTIVE_LOCUS",
    presentation: null,
    gate: null,
    entry: null,
    reentry_checkpoint: checkpoint,
    active_locus: null
  };
}

export function deriveActorContext(state, actorId, observedAt = Math.floor(Date.now() / 1000)) {
  validateVMState(state);
  requireString(actorId, "actor_id");
  if (state.body.posture === "SUCCESSION_COMMITTED") return contextForNoLocus(state, actorId);
  if (!state.active_locus_id) return contextForNoLocus(state, actorId);

  const locus = state.loci[state.active_locus_id];
  if (!locus || locus.phase !== "OPEN") throw new Error("active_locus_id does not name an open locus");
  const presentation = locus.presentations?.[actorId] ?? null;
  const gate = locus.gates?.[actorId] ?? null;
  const entry = locus.entries?.[actorId] ?? null;
  if (gate) {
    requireString(gate.disposition, "gate.disposition");
    requireString(gate.offer_status, "gate.offer_status");
    requireSafeUint(gate.offer_expires_at, "gate.offer_expires_at");
  }
  if (entry) requireString(entry.status, "entry.status");
  const checkpoint = entry?.departure_checkpoint_id
    ? state.departure_checkpoints[entry.departure_checkpoint_id] ?? null
    : findSealedCheckpoint(state, actorId);

  let phase = "UNPRESENTED";
  if (presentation) phase = "PRESENTED";
  if (gate?.disposition === "ADMIT" && gate.offer_status === "OPEN" && observedAt < gate.offer_expires_at) {
    phase = findSealedCheckpoint(state, actorId) ? "REENTRY_AVAILABLE" : "ADMISSION_OFFERED";
  }
  if (entry?.status === "PRESENT") {
    phase = checkpoint?.status === "OPEN" ? "DEPARTURE_CHECKPOINTED" : "PRESENT";
  } else if (entry && entry.status !== "PRESENT") {
    phase = "ENDED";
  }
  if (state.body.posture === "HOLD_CAPACITY_DEFICIT") phase = "HOLD_CAPACITY_DEFICIT";

  return {
    phase,
    presentation,
    gate,
    entry,
    reentry_checkpoint: checkpoint,
    active_locus: locus
  };
}

export function makeObservation(status, state, actorId, profile, observedAt = Math.floor(Date.now() / 1000)) {
  const actor = deriveActorContext(state, actorId, observedAt);
  return {
    schema: MEDIUM_OBSERVATION_SCHEMA,
    observed_at: observedAt,
    accepted_revision: state.revision,
    accepted_state_commitment: state.state_commitment,
    host_locality_id: state.host_locality_id,
    active_locus_id: state.active_locus_id,
    body_posture: state.body.posture,
    presence_count: state.body.presence_count,
    actor_id: actorId,
    embodiment_profile_id: profile.profile_id,
    embodiment_kind: profile.embodiment_kind,
    phase: actor.phase,
    presentation_id: actor.presentation?.presentation_id ?? "",
    entry_transition_id: actor.entry?.entry_transition_id ?? "",
    departure_checkpoint_id: actor.entry?.departure_checkpoint_id ?? actor.reentry_checkpoint?.checkpoint_id ?? "",
    vm_height: status.height,
    effect: "OBSERVES_ACCEPTED_LOCALITY_STATE_ONLY",
    non_effect: "DOES_NOT_ESTABLISH_EXTERNAL_BODY_OBEDIENCE"
  };
}

import {
  ATTRIBUTION_CEILING,
  CORRECTION_OPERATIONS,
  DEFAULT_DISPOSITION,
  DISPOSITIONS,
  END_OPERATIONS,
  EVIDENCE_CEILING,
  FIELD_DISPOSITION_SCHEMA,
  NON_EFFECTS,
  REQUIRED_EFFECT_CEILING,
  VM_EFFECTS,
  VM_ID,
  VM_RECEIPT_SCHEMA,
  VM_REQUIRED_ADAPTER_CAPABILITIES,
  VM_RPCCHAINVM_PROTOCOL,
  VM_VERSION
} from "./constants.js";
import { canonicalize, isDigest, requireDigest, sha256Hex } from "./canonical.js";
import { actorIdFromPublicKey } from "./config.js";
import { capacityView, makeObservation, validateVMState } from "./state.js";
import { adapterAllowsOperation } from "./adapter.js";
import { buildPassage, passageCommitment } from "./continuity.js";
import {
  draftCommitment,
  normalizedTransition,
  signingBytes,
  transitionID,
  validateUnsigned,
  verifyTransitionSignature
} from "./signature.js";

const SAFE_ID = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/;

function assertStatus(status, state, config) {
  if (!status || typeof status !== "object" || Array.isArray(status)) throw new Error("VM status is invalid");
  if (status.healthy !== true) throw new Error("VM reports unhealthy");
  if (status.vm_version !== VM_VERSION || config.expected_vm_version !== VM_VERSION) throw new Error("VM version mismatch");
  if (status.rpcchainvm_protocol !== VM_RPCCHAINVM_PROTOCOL || config.expected_rpcchainvm_protocol !== VM_RPCCHAINVM_PROTOCOL) {
    throw new Error("RPCChainVM protocol mismatch");
  }
  if (config.expected_vm_id !== VM_ID) throw new Error("configured VM ID mismatch");
  if (status.locality_id !== config.expected_locality_id || state.host_locality_id !== config.expected_locality_id) {
    throw new Error("host locality identity mismatch");
  }
  if (!Number.isSafeInteger(status.height) || status.height < 0) throw new Error("VM height is invalid");
  if (status.revision !== state.revision || status.state_commitment !== state.state_commitment) {
    throw new Error("incoherent VM status/state snapshot; retry from a fresh accepted state");
  }
  if (status.active_locus_id !== state.active_locus_id) throw new Error("status/state active locus mismatch");
  if (status.body?.posture !== state.body.posture || status.body?.presence_count !== state.body.presence_count) {
    throw new Error("status/state body mismatch");
  }
  if (Boolean(status.successor_boundary) !== Boolean(state.successor_boundary)) {
    throw new Error("status/state successor boundary mismatch");
  }
  if (!Number.isSafeInteger(status.pending_transitions) || status.pending_transitions < 0) {
    throw new Error("VM pending transition count is invalid");
  }
}

function decisionEffect(capability) {
  return VM_EFFECTS[capability] ?? "ALLOWS_ONE_STATE_BOUND_LOCAL_CAPABILITY_INVOCATION_ONLY";
}

// dispositionFor maps a state-bound decision to exactly one of the six bounded
// dispositions. REFUSE is the fail-closed default; HOLD is reserved for the
// capacity-deficit posture; END covers the egress operations; CORRECT covers a
// correction; everything else allowed is SUPPORT. RELEASE is only reached by an
// explicit release of a locally held reservation.
export function dispositionFor(capability, allowed, phase) {
  let disposition;
  if (!allowed) {
    disposition = phase === "HOLD_CAPACITY_DEFICIT" ? "HOLD" : "REFUSE";
  } else if (CORRECTION_OPERATIONS.includes(capability)) {
    disposition = "CORRECT";
  } else if (END_OPERATIONS.includes(capability)) {
    disposition = "END";
  } else {
    disposition = "SUPPORT";
  }
  if (!DISPOSITIONS.includes(disposition)) return DEFAULT_DISPOSITION;
  return disposition;
}

function publicDecision(decision) {
  return {
    schema: decision.schema,
    decision_id: decision.decision_id,
    decided_at: decision.decided_at,
    allowed: decision.allowed,
    disposition: decision.disposition,
    reason: decision.reason,
    capability: decision.capability,
    phase: decision.phase,
    accepted_revision: decision.accepted_revision,
    accepted_state_commitment: decision.accepted_state_commitment,
    actor_id: decision.actor_id,
    active_locus_id: decision.active_locus_id,
    attribution_ceiling: ATTRIBUTION_CEILING,
    evidence_ceiling: EVIDENCE_CEILING,
    effect: decision.effect,
    non_effects: decision.non_effects
  };
}

export class FieldEngine {
  constructor({ config, adapter, client, journal }) {
    this.config = config;
    this.adapter = adapter;
    this.client = client;
    this.journal = journal;
    this.pendingDrafts = new Map();
    this.holds = new Map();
    // Highest accepted revision this process has seen. A read at a lower
    // revision is a rollback and is refused.
    this.lastSeenRevision = -1;
  }

  async observe() {
    const [status, state] = await Promise.all([this.client.status(), this.client.state()]);
    validateVMState(state);
    assertStatus(status, state, this.config);
    if (this.lastSeenRevision >= 0 && state.revision < this.lastSeenRevision) {
      throw new Error(`STALE_ACCEPTED_REVISION_REFUSED: saw ${state.revision} after ${this.lastSeenRevision}`);
    }
    this.lastSeenRevision = Math.max(this.lastSeenRevision, state.revision);
    const observation = makeObservation(status, state, this.config.actor_id, this.adapter);
    return { status, state, observation };
  }

  async decide(capability, { expectedStateCommitment } = {}) {
    if (typeof capability !== "string" || !/^[A-Z][A-Z0-9_]{1,95}$/.test(capability)) {
      throw new Error("capability is invalid");
    }
    if (expectedStateCommitment !== undefined && !isDigest(expectedStateCommitment)) {
      throw new Error("expected state commitment must be lowercase SHA-256 hex");
    }
    const { status, state, observation } = await this.observe();
    const phaseCapabilities = this.adapter.capabilities_by_phase[observation.phase] ?? [];
    let allowed = phaseCapabilities.includes(capability);
    let reason = allowed ? "ADAPTER_AND_ACCEPTED_PHASE_ALLOW" : "ADAPTER_OR_ACCEPTED_PHASE_DENY";

    if (expectedStateCommitment && expectedStateCommitment !== observation.accepted_state_commitment) {
      allowed = false;
      reason = "EXPECTED_STATE_COMMITMENT_IS_STALE";
    }
    if (Object.hasOwn(VM_EFFECTS, capability) && !adapterAllowsOperation(this.adapter, capability)) {
      allowed = false;
      reason = "OPERATION_OUTSIDE_ADAPTER_DECLARED_CAPABILITIES";
    }
    const boundKey = state.actor_keys?.[this.config.actor_id];
    if (boundKey && boundKey !== this.config.actor_public_key) {
      allowed = false;
      reason = "ACCEPTED_ACTOR_KEY_BINDING_MISMATCH";
    }
    if (!boundKey && Object.hasOwn(VM_EFFECTS, capability) && capability !== "PRESENT_FORM") {
      allowed = false;
      reason = "UNKNOWN_ACTOR_MUST_PRESENT_FORM_FIRST";
    }
    if (observation.phase === "SUCCESSION_COMMITTED" && capability !== "READ_PUBLIC_STATE" && capability !== "VERIFY_LOCAL_EVIDENCE") {
      allowed = false;
      reason = "PREDECESSOR_FROZEN_AT_SUCCESSION_BOUNDARY";
    }
    if (status.pending_transitions !== 0 && capability !== "READ_PUBLIC_STATE" && capability !== "VERIFY_LOCAL_EVIDENCE") {
      allowed = false;
      reason = "PENDING_TRANSITION_REQUIRES_FRESH_ACCEPTED_STATE";
    }
    if (Object.hasOwn(VM_EFFECTS, capability) && observation.presentation_id && capability !== "PRESENT_FORM") {
      // medium_capabilities/effect_ceiling are PRESENCE_AVALANCHE_VM_001 wire
      // fields on the accepted presentation; the FIELD only reads them.
      const declared = observation.phase === "ENDED"
        ? state.presentation_history?.[observation.presentation_id]?.medium_capabilities
        : state.loci?.[state.active_locus_id]?.presentations?.[this.config.actor_id]?.medium_capabilities;
      if (Array.isArray(declared) && !declared.includes(capability)) {
        allowed = false;
        reason = "PRESENTED_ADAPTER_DID_NOT_DECLARE_CAPABILITY";
      }
      if (Array.isArray(declared) && VM_REQUIRED_ADAPTER_CAPABILITIES.some((required) => !declared.includes(required))) {
        allowed = false;
        reason = "PRESENTED_ADAPTER_DECLARATION_IS_INCOMPLETE";
      }
      const presentation = observation.phase === "ENDED"
        ? state.presentation_history?.[observation.presentation_id]
        : state.loci?.[state.active_locus_id]?.presentations?.[this.config.actor_id];
      if (Array.isArray(presentation?.effect_ceiling) && REQUIRED_EFFECT_CEILING.some((required) => !presentation.effect_ceiling.includes(required))) {
        allowed = false;
        reason = "PRESENTED_EFFECT_CEILING_IS_INCOMPLETE";
      }
    }

    const disposition = dispositionFor(capability, allowed, observation.phase);
    const decidedAt = Math.floor(Date.now() / 1000);
    const basis = {
      schema: FIELD_DISPOSITION_SCHEMA,
      decided_at: decidedAt,
      allowed,
      disposition,
      reason,
      capability,
      phase: observation.phase,
      accepted_revision: observation.accepted_revision,
      accepted_state_commitment: observation.accepted_state_commitment,
      actor_id: this.config.actor_id,
      active_locus_id: observation.active_locus_id ?? "",
      effect: decisionEffect(capability),
      non_effects: NON_EFFECTS
    };
    const decision = { ...basis, decision_id: sha256Hex(basis) };
    await this.journal.append("CAPABILITY_DECISION", {
      decision_id: decision.decision_id,
      allowed,
      disposition,
      reason,
      capability,
      phase: observation.phase,
      accepted_revision: observation.accepted_revision,
      accepted_state_commitment: observation.accepted_state_commitment
    });
    return { ...decision, observation, state };
  }

  async assertAuthorized(capability, options) {
    const decision = await this.decide(capability, options);
    if (!decision.allowed) {
      const error = new Error(`capability ${capability} denied (${decision.disposition}): ${decision.reason}`);
      error.code = "CAPABILITY_DENIED";
      error.disposition = decision.disposition;
      error.decision = publicDecision(decision);
      throw error;
    }
    return decision;
  }

  async guard(capability, handler, options) {
    if (typeof handler !== "function") throw new Error("guard handler must be a function");
    const decision = await this.assertAuthorized(capability, options);
    return handler(publicDecision(decision));
  }

  dispositionRecord(disposition, fields) {
    if (!DISPOSITIONS.includes(disposition)) throw new Error(`unknown disposition ${disposition}`);
    return {
      schema: FIELD_DISPOSITION_SCHEMA,
      disposition,
      decided_at: Math.floor(Date.now() / 1000),
      actor_id: this.config.actor_id,
      adapter_id: this.adapter.adapter_id,
      controlled_surface: this.adapter.controlled_surface,
      attribution_ceiling: ATTRIBUTION_CEILING,
      evidence_ceiling: EVIDENCE_CEILING,
      ...fields
    };
  }

  // reserve places a bounded FIELD-local hold. It touches no VM state and claims
  // no acceptance; it only prevents the FIELD from over-committing situated
  // capacity before a decision is reached.
  async reserve(reservationId, units, { reason_sha256 } = {}) {
    if (typeof reservationId !== "string" || !SAFE_ID.test(reservationId)) throw new Error("reservation_id is invalid");
    if (!Number.isSafeInteger(units) || units <= 0) throw new Error("reserved units must be a positive integer");
    if (reason_sha256 !== undefined) requireDigest(reason_sha256, "reason_sha256");
    if (this.holds.has(reservationId)) throw new Error("reservation is already held");
    this.holds.set(reservationId, { units, reason_sha256: reason_sha256 ?? null, held_at: Math.floor(Date.now() / 1000) });
    await this.journal.append("RESERVATION_HELD", { reservation_id: reservationId, units, ...(reason_sha256 ? { reason_sha256 } : {}) });
    return this.dispositionRecord("HOLD", {
      reservation_id: reservationId,
      units,
      effect: "LOCAL_RESERVATION_HELD_NO_VM_SUBMISSION",
      non_effect: "DOES_NOT_RESERVE_OR_ACCEPT_ANYTHING_IN_VM_STATE"
    });
  }

  // release frees a held reservation. It never fabricates acceptance: releasing
  // a hold means only that the FIELD is no longer holding it locally.
  async release(reservationId, { reason_sha256 } = {}) {
    if (typeof reservationId !== "string" || !SAFE_ID.test(reservationId)) throw new Error("reservation_id is invalid");
    if (reason_sha256 !== undefined) requireDigest(reason_sha256, "reason_sha256");
    const held = this.holds.get(reservationId);
    if (!held) throw new Error("no such held reservation to release");
    this.holds.delete(reservationId);
    await this.journal.append("RESERVATION_RELEASED", { reservation_id: reservationId, units: held.units, ...(reason_sha256 ? { reason_sha256 } : {}) });
    return this.dispositionRecord("RELEASE", {
      reservation_id: reservationId,
      units: held.units,
      effect: "LOCAL_RESERVATION_RELEASED_NO_ACCEPTANCE_FABRICATED",
      non_effect: "DOES_NOT_CLAIM_ANY_VM_ACCEPTANCE"
    });
  }

  heldReservations() {
    return [...this.holds.keys()].sort();
  }

  // recordCorrection appends a correction and retains the predecessor. The
  // journal is append-only, so the corrected record remains readable forever.
  async recordCorrection({ target_ref, replacement_commitment, reason_sha256 }) {
    if (typeof target_ref !== "string" || !SAFE_ID.test(target_ref)) {
      requireDigest(target_ref, "target_ref");
    }
    requireDigest(replacement_commitment, "replacement_commitment");
    requireDigest(reason_sha256, "reason_sha256");
    const event = await this.journal.append("CORRECTION_APPENDED", {
      target_ref,
      replacement_commitment,
      reason_sha256
    });
    return this.dispositionRecord("CORRECT", {
      target_ref,
      replacement_commitment,
      reason_sha256,
      correction_sequence: event.sequence,
      effect: "APPENDS_LOCAL_CORRECTION_RETAINING_PREDECESSOR_ONLY",
      non_effect: "DOES_NOT_OVERWRITE_OR_DELETE_THE_CORRECTED_RECORD"
    });
  }

  // recordDeparture writes a local departure checkpoint. It explicitly does not
  // claim continued presence: a departure is an ending, not a presence.
  async recordDeparture({ checkpoint_id, from_state_commitment, departure_state_commitment }) {
    requireDigest(from_state_commitment, "from_state_commitment");
    requireDigest(departure_state_commitment, "departure_state_commitment");
    if (typeof checkpoint_id !== "string" || !SAFE_ID.test(checkpoint_id)) throw new Error("checkpoint_id is invalid");
    await this.journal.append("DEPARTURE_CHECKPOINTED_LOCAL", {
      checkpoint_id,
      from_state_commitment,
      departure_state_commitment
    });
    return this.dispositionRecord("END", {
      checkpoint_id,
      from_state_commitment,
      departure_state_commitment,
      claims_continued_presence: false,
      effect: "RECORDS_LOCAL_DEPARTURE_CHECKPOINT_ONLY",
      non_effect: "DOES_NOT_CLAIM_CONTINUED_PRESENCE_OR_LIVENESS"
    });
  }

  // recordReentry carries state succession forward. Re-entry is succession, not
  // restoration: the successor commitment is derived from the predecessor, and
  // the record asserts restoration=false so no reader mistakes it for a rewind.
  async recordReentry({ participant_id, from_state_commitment, delta_digests }) {
    const built = buildPassage(participant_id, from_state_commitment, delta_digests);
    const commitment = passageCommitment(participant_id, from_state_commitment, built.passage, built.result_state_commitment);
    await this.journal.append("REENTRY_SUCCESSION_LOCAL", {
      participant_id,
      from_state_commitment,
      result_state_commitment: built.result_state_commitment,
      passage_commitment: commitment
    });
    return this.dispositionRecord("SUPPORT", {
      participant_id,
      from_state_commitment,
      result_state_commitment: built.result_state_commitment,
      passage_commitment: commitment,
      succession: true,
      restoration: false,
      effect: "CARRIES_STATE_SUCCESSION_FORWARD_ONLY",
      non_effect: "DOES_NOT_RESTORE_OR_REWIND_PRIOR_PRESENCE"
    });
  }

  async prepareDraft(request) {
    if (!request || typeof request !== "object" || Array.isArray(request)) throw new Error("draft request must be an object");
    const fields = Object.keys(request).sort();
    const expectedFields = ["locus_id", "observed_at", "operation", "payload"].sort();
    if (fields.length !== expectedFields.length || fields.some((field, index) => field !== expectedFields[index])) {
      throw new Error("draft request fields do not match the field contract");
    }
    const decision = await this.assertAuthorized(request.operation);
    if (request.locus_id !== (decision.observation.active_locus_id ?? "") && request.operation !== "BOUND") {
      throw new Error("draft locus_id is not the accepted active locus");
    }
    if (!Number.isSafeInteger(request.observed_at) || request.observed_at < 0) throw new Error("observed_at is invalid");
    if (!request.payload || typeof request.payload !== "object" || Array.isArray(request.payload)) throw new Error("payload must be an object");
    const vmRequest = {
      operation: request.operation,
      actor_id: this.config.actor_id,
      actor_public_key: this.config.actor_public_key,
      locus_id: request.locus_id,
      observed_at: request.observed_at,
      payload: request.payload
    };
    const response = await this.client.draft(vmRequest);
    const unsigned = validateUnsigned(response.unsigned);
    if (
      unsigned.operation !== request.operation ||
      unsigned.actor_id !== this.config.actor_id ||
      unsigned.actor_public_key !== this.config.actor_public_key ||
      unsigned.locus_id !== request.locus_id ||
      unsigned.observed_at !== request.observed_at ||
      unsigned.effect !== VM_EFFECTS[request.operation] ||
      unsigned.revision !== decision.accepted_revision + 1 ||
      unsigned.previous_state_commitment !== decision.accepted_state_commitment ||
      canonicalize(unsigned.payload) !== canonicalize(request.payload)
    ) {
      throw new Error("VM draft does not match the authorized state-bound request");
    }
    const commitment = draftCommitment(unsigned);
    this.pendingDrafts.set(commitment, {
      unsigned,
      decision_id: decision.decision_id,
      prepared_at: Math.floor(Date.now() / 1000)
    });
    await this.journal.append("DRAFT_PREPARED", {
      draft_commitment: commitment,
      decision_id: decision.decision_id,
      operation: unsigned.operation,
      revision: unsigned.revision,
      previous_state_commitment: unsigned.previous_state_commitment,
      payload_sha256: sha256Hex(unsigned.payload),
      signing_bytes_sha256: sha256Hex(signingBytes(unsigned))
    });
    return {
      unsigned,
      draft_commitment: commitment,
      signing_bytes_sha256: sha256Hex(signingBytes(unsigned)),
      effect: "DRAFT_ONLY_NO_STATE_CHANGE",
      next_step: "sign the exact unsigned object with the configured external protected Ed25519 signer"
    };
  }

  async submitTransition(transition) {
    verifyTransitionSignature(transition);
    if (actorIdFromPublicKey(transition.unsigned.actor_public_key) !== transition.unsigned.actor_id) {
      throw new Error("transition actor ID does not derive from public key");
    }
    if (transition.unsigned.actor_id !== this.config.actor_id || transition.unsigned.actor_public_key !== this.config.actor_public_key) {
      throw new Error("transition signer is not the configured field actor");
    }
    const commitment = draftCommitment(transition.unsigned);
    const prepared = this.pendingDrafts.get(commitment);
    if (!prepared || canonicalize(prepared.unsigned) !== canonicalize(transition.unsigned)) {
      throw new Error("transition has no live state-bound draft in this field process");
    }
    await this.assertAuthorized(transition.unsigned.operation, {
      expectedStateCommitment: transition.unsigned.previous_state_commitment
    });
    const normalized = normalizedTransition(prepared.unsigned, transition.signature);
    const expectedID = transitionID(normalized);
    // Submit the exact canonical bytes whose SHA-256 is the transition id.
    const response = await this.client.submit(normalized);
    if (response.transition_id !== expectedID || response.status !== "PENDING_CONSENSUS") {
      throw new Error("VM submission response does not match the signed transition");
    }
    this.pendingDrafts.delete(commitment);
    await this.journal.append("TRANSITION_SUBMITTED", {
      transition_id: expectedID,
      draft_commitment: commitment,
      operation: normalized.unsigned.operation,
      revision: normalized.unsigned.revision,
      previous_state_commitment: normalized.unsigned.previous_state_commitment,
      effect: "NO_ACCEPTANCE_UNTIL_BLOCK_ACCEPTED"
    });
    return response;
  }

  async waitReceipt(id) {
    if (!isDigest(id)) throw new Error("transition ID must be lowercase SHA-256 hex");
    const deadline = Date.now() + this.config.receipt_timeout_ms;
    let receipt = null;
    while (Date.now() < deadline) {
      receipt = await this.client.receipt(id);
      if (receipt) break;
      await new Promise((resolve) => setTimeout(resolve, 300));
    }
    if (!receipt) throw new Error("receipt timeout: transition remains unaccepted or unavailable");
    if (
      receipt.schema !== VM_RECEIPT_SCHEMA || receipt.transition_id !== id ||
      receipt.actor_id !== this.config.actor_id || !isDigest(receipt.state_commitment) ||
      !Number.isSafeInteger(receipt.revision) || receipt.revision < 1 ||
      receipt.effect !== VM_EFFECTS[receipt.operation]
    ) {
      throw new Error("accepted receipt violates the frozen VM contract");
    }
    const { observation } = await this.observe();
    if (observation.accepted_revision < receipt.revision) throw new Error("receipt revision is ahead of accepted state");
    if (observation.accepted_revision === receipt.revision && observation.accepted_state_commitment !== receipt.state_commitment) {
      throw new Error("receipt state commitment differs from accepted state");
    }
    await this.journal.append("RECEIPT_OBSERVED", {
      transition_id: id,
      operation: receipt.operation,
      accepted_revision: receipt.revision,
      accepted_state_commitment: receipt.state_commitment,
      effect: receipt.effect,
      posture: "HISTORICAL_EVIDENCE_NOT_CURRENT_PRESENCE"
    });
    return receipt;
  }
}

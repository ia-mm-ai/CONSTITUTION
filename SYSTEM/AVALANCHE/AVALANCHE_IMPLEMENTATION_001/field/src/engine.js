import {
  MEDIUM_DECISION_SCHEMA,
  NON_EFFECTS,
  REQUIRED_EFFECT_CEILING,
  VM_EFFECTS,
  VM_ID,
  VM_RECEIPT_SCHEMA,
  VM_REQUIRED_MEDIUM_CAPABILITIES,
  VM_RPCCHAINVM_PROTOCOL,
  VM_VERSION
} from "./constants.js";
import { canonicalize, isDigest, sha256Hex } from "./canonical.js";
import { actorIdFromPublicKey } from "./config.js";
import { makeObservation, validateVMState } from "./state.js";
import { profileAllowsOperation } from "./profile.js";
import {
  draftCommitment,
  normalizedTransition,
  signingBytes,
  transitionID,
  validateUnsigned,
  verifyTransitionSignature
} from "./signature.js";

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

function publicDecision(decision) {
  return {
    schema: decision.schema,
    decision_id: decision.decision_id,
    decided_at: decision.decided_at,
    allowed: decision.allowed,
    reason: decision.reason,
    capability: decision.capability,
    phase: decision.phase,
    accepted_revision: decision.accepted_revision,
    accepted_state_commitment: decision.accepted_state_commitment,
    actor_id: decision.actor_id,
    active_locus_id: decision.active_locus_id,
    effect: decision.effect,
    non_effects: decision.non_effects
  };
}

export class MediumEngine {
  constructor({ config, profile, client, journal }) {
    this.config = config;
    this.profile = profile;
    this.client = client;
    this.journal = journal;
    this.pendingDrafts = new Map();
  }

  async observe() {
    const [status, state] = await Promise.all([this.client.status(), this.client.state()]);
    validateVMState(state);
    assertStatus(status, state, this.config);
    const observation = makeObservation(status, state, this.config.actor_id, this.profile);
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
    const phaseCapabilities = this.profile.capabilities_by_phase[observation.phase] ?? [];
    let allowed = phaseCapabilities.includes(capability);
    let reason = allowed ? "PROFILE_AND_ACCEPTED_PHASE_ALLOW" : "PROFILE_OR_ACCEPTED_PHASE_DENY";

    if (expectedStateCommitment && expectedStateCommitment !== observation.accepted_state_commitment) {
      allowed = false;
      reason = "EXPECTED_STATE_COMMITMENT_IS_STALE";
    }
    if (Object.hasOwn(VM_EFFECTS, capability) && !profileAllowsOperation(this.profile, capability)) {
      allowed = false;
      reason = "OPERATION_OUTSIDE_PROFILE_ALLOWLIST";
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
      const declared = observation.phase === "ENDED"
        ? state.presentation_history?.[observation.presentation_id]?.medium_capabilities
        : state.loci?.[state.active_locus_id]?.presentations?.[this.config.actor_id]?.medium_capabilities;
      if (Array.isArray(declared) && !declared.includes(capability)) {
        allowed = false;
        reason = "PRESENTED_MEDIUM_DID_NOT_DECLARE_CAPABILITY";
      }
      if (Array.isArray(declared) && VM_REQUIRED_MEDIUM_CAPABILITIES.some((required) => !declared.includes(required))) {
        allowed = false;
        reason = "PRESENTED_MEDIUM_DECLARATION_IS_INCOMPLETE";
      }
      const presentation = observation.phase === "ENDED"
        ? state.presentation_history?.[observation.presentation_id]
        : state.loci?.[state.active_locus_id]?.presentations?.[this.config.actor_id];
      if (Array.isArray(presentation?.effect_ceiling) && REQUIRED_EFFECT_CEILING.some((required) => !presentation.effect_ceiling.includes(required))) {
        allowed = false;
        reason = "PRESENTED_EFFECT_CEILING_IS_INCOMPLETE";
      }
    }

    const decidedAt = Math.floor(Date.now() / 1000);
    const basis = {
      schema: MEDIUM_DECISION_SCHEMA,
      decided_at: decidedAt,
      allowed,
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
      const error = new Error(`capability ${capability} denied: ${decision.reason}`);
      error.code = "CAPABILITY_DENIED";
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

  async prepareDraft(request) {
    if (!request || typeof request !== "object" || Array.isArray(request)) throw new Error("draft request must be an object");
    const fields = Object.keys(request).sort();
    const expectedFields = ["locus_id", "observed_at", "operation", "payload"].sort();
    if (fields.length !== expectedFields.length || fields.some((field, index) => field !== expectedFields[index])) {
      throw new Error("draft request fields do not match the medium contract");
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
      throw new Error("transition signer is not the configured medium actor");
    }
    const commitment = draftCommitment(transition.unsigned);
    const prepared = this.pendingDrafts.get(commitment);
    if (!prepared || canonicalize(prepared.unsigned) !== canonicalize(transition.unsigned)) {
      throw new Error("transition has no live state-bound draft in this medium process");
    }
    await this.assertAuthorized(transition.unsigned.operation, {
      expectedStateCommitment: transition.unsigned.previous_state_commitment
    });
    const normalized = normalizedTransition(prepared.unsigned, transition.signature);
    const expectedID = transitionID(normalized);
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
      effect: receipt.effect
    });
    return receipt;
  }
}

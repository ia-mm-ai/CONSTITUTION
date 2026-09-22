import {
  ATTRIBUTION_CEILING,
  EVIDENCE_CEILING,
  FIELD_ADAPTER_SCHEMA,
  NON_EFFECTS,
  PHASES,
  REQUIRED_EFFECT_CEILING,
  VM_EFFECTS
} from "./constants.js";
import { assertPlainObject, deepFreeze } from "./canonical.js";
import { readFile } from "node:fs/promises";

const SAFE_TOKEN = /^[A-Z][A-Z0-9_]{1,95}$/;

function exactKeys(value, allowed, label) {
  for (const key of Object.keys(value)) {
    if (!allowed.has(key)) throw new Error(`${label} contains unknown field ${key}`);
  }
}

function uniqueTokens(value, label) {
  if (!Array.isArray(value)) throw new Error(`${label} must be an array`);
  const seen = new Set();
  for (const item of value) {
    if (typeof item !== "string" || !SAFE_TOKEN.test(item)) throw new Error(`${label} contains an invalid token`);
    if (seen.has(item)) throw new Error(`${label} contains duplicate ${item}`);
    seen.add(item);
  }
  return [...seen].sort();
}

function boundedStatements(value, label, minItems) {
  if (!Array.isArray(value) || value.length < minItems) throw new Error(`${label} must list at least ${minItems} statement(s)`);
  const seen = new Set();
  for (const item of value) {
    if (typeof item !== "string" || item.length < 8 || item.length > 512) throw new Error(`${label} statements must be 8..512 characters`);
    if (seen.has(item)) throw new Error(`${label} contains a duplicate statement`);
    seen.add(item);
  }
  return [...value];
}

// A FIELD adapter declares a bounded surface under operator control. It never
// declares a kind of being. There is no participant-type ontology: an adapter
// classifies "the HTTP request/response surface of one operator endpoint", not
// "a human" or "an AI". Everywhere the predecessor recorded an embodiment kind,
// the FIELD records adapter_id + controlled_surface + a SURFACE_ONLY attribution.
export function validateAdapter(input) {
  assertPlainObject(input, "adapter");
  exactKeys(input, new Set([
    "schema", "adapter_id", "version", "controlled_surface", "scope_statement",
    "declared_capabilities", "attribution_ceiling", "evidence_ceiling",
    "signer_boundary", "private_state_posture", "capabilities_by_phase",
    "refusals", "non_claims", "effect_ceiling", "non_effects"
  ]), "adapter");

  if (input.schema !== FIELD_ADAPTER_SCHEMA) throw new Error(`adapter.schema must be ${FIELD_ADAPTER_SCHEMA}`);
  if (typeof input.adapter_id !== "string" || !SAFE_TOKEN.test(input.adapter_id)) throw new Error("adapter_id is invalid");
  if (input.version !== "1.0.0") throw new Error("adapter.version must be 1.0.0");
  if (typeof input.controlled_surface !== "string" || input.controlled_surface.length < 16 || input.controlled_surface.length > 512) {
    throw new Error("controlled_surface must be a bounded statement of exactly what surface is controlled");
  }
  if (typeof input.scope_statement !== "string" || input.scope_statement.length < 16 || input.scope_statement.length > 512) {
    throw new Error("scope_statement is required");
  }
  if (input.attribution_ceiling !== ATTRIBUTION_CEILING) throw new Error(`attribution_ceiling must be ${ATTRIBUTION_CEILING}`);
  if (input.evidence_ceiling !== EVIDENCE_CEILING) throw new Error(`evidence_ceiling must be ${EVIDENCE_CEILING}`);
  if (input.signer_boundary !== "EXTERNAL_PROTECTED_SIGNER") throw new Error("adapter must use EXTERNAL_PROTECTED_SIGNER");
  if (input.private_state_posture !== "PARTICIPANT_CUSTODY_COMMITMENTS_ONLY") {
    throw new Error("adapter must keep private state in participant custody");
  }
  if (Object.hasOwn(input, "embodiment_kind") || Object.hasOwn(input, "embodiment_profile_id")) {
    throw new Error("adapters must not declare a participant type or embodiment kind");
  }

  const declared = uniqueTokens(input.declared_capabilities, "declared_capabilities");
  for (const capability of declared) {
    if (!Object.hasOwn(VM_EFFECTS, capability)) throw new Error(`declared_capabilities names unknown VM operation ${capability}`);
  }

  assertPlainObject(input.capabilities_by_phase, "capabilities_by_phase");
  exactKeys(input.capabilities_by_phase, new Set(PHASES), "capabilities_by_phase");
  const capabilitiesByPhase = {};
  for (const phase of PHASES) {
    if (!Object.hasOwn(input.capabilities_by_phase, phase)) throw new Error(`capabilities_by_phase is missing ${phase}`);
    const tokens = uniqueTokens(input.capabilities_by_phase[phase], `capabilities_by_phase.${phase}`);
    for (const token of tokens) {
      // A phase may only gate a VM operation the adapter has explicitly
      // declared. Phase gating can narrow the declared surface, never widen it.
      if (Object.hasOwn(VM_EFFECTS, token) && !declared.includes(token)) {
        throw new Error(`capabilities_by_phase.${phase} grants undeclared VM operation ${token}`);
      }
    }
    capabilitiesByPhase[phase] = tokens;
  }

  const refusals = boundedStatements(input.refusals, "refusals", 1);
  const nonClaims = boundedStatements(input.non_claims, "non_claims", 1);

  const ceilings = uniqueTokens(input.effect_ceiling, "effect_ceiling");
  for (const required of REQUIRED_EFFECT_CEILING) {
    if (!ceilings.includes(required)) throw new Error(`effect_ceiling is missing ${required}`);
  }
  const nonEffects = uniqueTokens(input.non_effects, "non_effects");
  for (const required of NON_EFFECTS) {
    if (!nonEffects.includes(required)) throw new Error(`non_effects is missing ${required}`);
  }

  return deepFreeze({
    ...input,
    declared_capabilities: declared,
    capabilities_by_phase: capabilitiesByPhase,
    refusals,
    non_claims: nonClaims,
    effect_ceiling: ceilings,
    non_effects: nonEffects
  });
}

export function adapterAllowsOperation(adapter, operation) {
  return adapter.declared_capabilities.includes(operation);
}

export async function loadAdapter(path) {
  return validateAdapter(JSON.parse(await readFile(path, "utf8")));
}

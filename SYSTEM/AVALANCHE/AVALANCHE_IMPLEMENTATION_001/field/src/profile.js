import {
  MEDIUM_PROFILE_SCHEMA,
  NON_EFFECTS,
  PHASES,
  REQUIRED_EFFECT_CEILING,
  VM_EFFECTS
} from "./constants.js";
import { assertPlainObject, deepFreeze } from "./canonical.js";
import { readFile } from "node:fs/promises";

const KINDS = new Set(["HUMAN", "AI", "DEVICE", "LOCALITY"]);
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
    if (typeof item !== "string" || !SAFE_TOKEN.test(item)) throw new Error(`${label} contains an invalid capability token`);
    if (seen.has(item)) throw new Error(`${label} contains duplicate ${item}`);
    seen.add(item);
  }
  return [...seen].sort();
}

export function validateProfile(input) {
  assertPlainObject(input, "profile");
  exactKeys(input, new Set([
    "schema", "profile_id", "version", "embodiment_kind", "description",
    "controlled_surface", "signer_boundary", "private_state_posture",
    "operation_allowlist", "capabilities_by_phase", "effect_ceiling", "non_effects"
  ]), "profile");
  if (input.schema !== MEDIUM_PROFILE_SCHEMA) throw new Error(`profile.schema must be ${MEDIUM_PROFILE_SCHEMA}`);
  if (typeof input.profile_id !== "string" || !SAFE_TOKEN.test(input.profile_id)) throw new Error("profile_id is invalid");
  if (input.version !== "1.0.0") throw new Error("profile.version must be 1.0.0");
  if (!KINDS.has(input.embodiment_kind)) throw new Error("unsupported embodiment_kind");
  if (typeof input.description !== "string" || input.description.length < 20) throw new Error("profile.description is too short");
  if (typeof input.controlled_surface !== "string" || input.controlled_surface.length < 8) throw new Error("controlled_surface is required");
  if (input.signer_boundary !== "EXTERNAL_PROTECTED_SIGNER") throw new Error("profile must use EXTERNAL_PROTECTED_SIGNER");
  if (input.private_state_posture !== "PARTICIPANT_CUSTODY_COMMITMENTS_ONLY") {
    throw new Error("profile must keep private state in participant custody");
  }

  const operations = uniqueTokens(input.operation_allowlist, "operation_allowlist");
  for (const operation of operations) {
    if (!Object.hasOwn(VM_EFFECTS, operation)) throw new Error(`unknown VM operation ${operation}`);
  }

  assertPlainObject(input.capabilities_by_phase, "capabilities_by_phase");
  exactKeys(input.capabilities_by_phase, new Set(PHASES), "capabilities_by_phase");
  const capabilitiesByPhase = {};
  for (const phase of PHASES) {
    if (!Object.hasOwn(input.capabilities_by_phase, phase)) throw new Error(`capabilities_by_phase is missing ${phase}`);
    capabilitiesByPhase[phase] = uniqueTokens(input.capabilities_by_phase[phase], `capabilities_by_phase.${phase}`);
  }

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
    operation_allowlist: operations,
    capabilities_by_phase: capabilitiesByPhase,
    effect_ceiling: ceilings,
    non_effects: nonEffects
  });
}

export function profileAllowsOperation(profile, operation) {
  return profile.operation_allowlist.includes(operation);
}

export async function loadProfile(path) {
  return validateProfile(JSON.parse(await readFile(path, "utf8")));
}

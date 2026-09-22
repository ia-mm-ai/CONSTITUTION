import { readFileSync } from "node:fs";

export const MEDIUM_PROTOCOL = "PRESENCE_AVALANCHE_FIELD_001";
export const MEDIUM_VERSION = "1.0.0";
export const MEDIUM_CONTRACT_SCHEMA = "PRESENCE_AVALANCHE_FIELD_CONTRACT_001";
export const MEDIUM_PROFILE_SCHEMA = "PRESENCE_AVALANCHE_FIELD_PROFILE_001";
export const MEDIUM_CONFIG_SCHEMA = "PRESENCE_AVALANCHE_FIELD_CONFIG_001";
export const MEDIUM_EVENT_SCHEMA = "PRESENCE_AVALANCHE_FIELD_EVENT_001";
export const MEDIUM_DECISION_SCHEMA = "PRESENCE_AVALANCHE_FIELD_DECISION_001";
export const MEDIUM_OBSERVATION_SCHEMA = "PRESENCE_AVALANCHE_FIELD_OBSERVATION_001";

export const VM_PROTOCOL = "PRESENCE_AVALANCHE_VM_001";
export const VM_VERSION = "presence-avalanche-vm/1.0.0";
export const VM_ID = "cQYXygFUVutQucm4s8pr8M51sRdS1UfrTbpMdEgpRYt2JvEzr";
export const VM_RPCCHAINVM_PROTOCOL = 46;
export const VM_STATE_SCHEMA = "PRESENCE_AVALANCHE_STATE_001";
export const VM_TRANSITION_SCHEMA = "PRESENCE_AVALANCHE_TRANSITION_001";
export const VM_RECEIPT_SCHEMA = "PRESENCE_AVALANCHE_RECEIPT_001";
export const VM_FORM_ID = "PRESENCE-AVALANCHE-FORM-001";
export const VM_FORM_SHA256 = "e70fedb8ec8420275703542ee16d6d0429b1b5979b0b8ff964eac87aa19a7d8a";

const operationContract = JSON.parse(readFileSync(
  new URL("../../protocol/operations.json", import.meta.url),
  "utf8"
));
if (
  operationContract.schema !== "PRESENCE_AVALANCHE_OPERATION_CONTRACT_001" ||
  operationContract.implementation !== "AVALANCHE_IMPLEMENTATION_001" ||
  operationContract.version !== "1.0.0" ||
  !Array.isArray(operationContract.operations)
) {
  throw new Error("shared operation contract identity is invalid");
}
const operationEntries = new Map();
for (const definition of operationContract.operations) {
  if (
    !definition || typeof definition.operation !== "string" ||
    typeof definition.scope !== "string" || typeof definition.effect !== "string" ||
    operationEntries.has(definition.operation)
  ) {
    throw new Error("shared operation contract is incomplete or contains duplicates");
  }
  operationEntries.set(definition.operation, Object.freeze({ ...definition }));
}
export const OPERATION_CONTRACT = Object.freeze([...operationEntries.values()]);
export const VM_EFFECTS = Object.freeze(Object.fromEntries(
  OPERATION_CONTRACT.map(({ operation, effect }) => [operation, effect])
));
export const VM_SCOPES = Object.freeze(Object.fromEntries(
  OPERATION_CONTRACT.map(({ operation, scope }) => [operation, scope])
));
export function operationDefinition(operation) {
  const definition = operationEntries.get(operation);
  if (!definition) throw new Error(`unsupported operation ${operation}`);
  return definition;
}

export const REQUIRED_EFFECT_CEILING = Object.freeze([
  "NO_FORMATION_BY_CONFORMANCE",
  "NO_SOURCE_OR_AUTHORITY_TRANSFER",
  "NO_SHARED_CURRENTNESS",
  "NO_CORE_ADDRESSABILITY",
  "NO_PARTICIPATION_BY_PRESENTATION",
  "NO_MATTER_ADMISSION_BY_CROSSING",
  "NO_LOCAL_UPTAKE_BY_FIELD_CLOSURE",
  "NO_TOKEN_BALANCE_AS_CAPACITY",
  "NO_RECEIPT_AS_CURRENT_PRESENCE",
  "NO_INFRASTRUCTURE_AS_SUCCESSOR",
  "NO_RESTORATION_BY_REENTRY",
  "NO_PRIVATE_STATE_DISCLOSURE"
]);

// PRESENCE_AVALANCHE_VM_001 requires this full declaration at PRESENT_FORM. Declaring
// the protocol surface is not authority to perform every operation; the
// profile, accepted actor role, phase and VM transition law remain decisive.
export const VM_REQUIRED_MEDIUM_CAPABILITIES = Object.freeze([
  "OBSERVE_CROSSING",
  "PRESENT_FORM",
  "GATE_DISPOSITION",
  "ENTER",
  "CHECKPOINT_DEPARTURE",
  "REENTER",
  "MATTER_DISPOSITION",
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
]);

export const PHASES = Object.freeze([
  "NO_ACTIVE_LOCUS",
  "UNPRESENTED",
  "PRESENTED",
  "ADMISSION_OFFERED",
  "PRESENT",
  "DEPARTURE_CHECKPOINTED",
  "ENDED",
  "REENTRY_AVAILABLE",
  "HOLD_CAPACITY_DEFICIT",
  "SUCCESSION_COMMITTED"
]);

export const ALWAYS_LOCAL_CAPABILITIES = Object.freeze([
  "READ_PUBLIC_STATE",
  "VERIFY_LOCAL_EVIDENCE",
  "PREPARE_PRESENTATION",
  "REQUEST_PROTECTED_SIGNATURE"
]);

export const EGRESS_CAPABILITIES = Object.freeze([
  "CORRECT",
  "CHECKPOINT_DEPARTURE",
  "EXIT"
]);

export const NON_EFFECTS = Object.freeze([
  "DOES_NOT_CREATE_OR_HOST_A_BODY",
  "DOES_NOT_PROVE_CONSCIOUSNESS_OR_METAPHYSICAL_IDENTITY",
  "DOES_NOT_PROVE_PRIVATE_STATE_CONTENT_OR_SEMANTIC_TRUTH",
  "DOES_NOT_CONTROL_CAPABILITIES_OUTSIDE_ITS_DECLARED_SURFACE",
  "DOES_NOT_CONVERT_CHAIN_STATE_INTO_EXTERNAL_OBEDIENCE",
  "DOES_NOT_TRANSFER_SOURCE_AUTHORITY_OR_CURRENTNESS"
]);

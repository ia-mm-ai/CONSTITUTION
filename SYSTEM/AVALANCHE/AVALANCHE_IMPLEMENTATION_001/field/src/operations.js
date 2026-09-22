import { readFileSync } from "node:fs";
import { createHash } from "node:crypto";

export const SUPPORTED_VM_OPERATIONS = Object.freeze([
  "BOUND", "PRESENT_FORM", "GATE_DISPOSITION", "ENTER", "CHECKPOINT_DEPARTURE",
  "REENTER", "OBSERVE_CROSSING", "MATTER_DISPOSITION", "RECORD_EMERGENCE",
  "CORRECT", "EXIT", "CLOSE", "INCORPORATE_ADDRESSED_RESIDUE", "DECLARE_CAPACITY",
  "PULSE", "RECLAIM_ADMISSION_OFFER", "REGISTER_CONTINUITY_AUTHORITY",
  "EXHAUST_FORMATION_AUTHORITY", "PROPOSE_SUCCESSOR", "ATTEST_SUCCESSOR",
  "ACTIVATE_SUCCESSOR"
]);

const SCOPES = new Set(["NEW_LOCUS", "ACTIVE_LOCUS", "BODY_LOCAL", "DORMANT_BODY", "TARGET_SCOPED"]);
const TOKEN = /^[A-Z][A-Z0-9_]+$/;

export function validateOperationContract(contract) {
  if (
    !contract || typeof contract !== "object" || Array.isArray(contract) ||
    Object.keys(contract).sort().join(",") !== "implementation,operations,protocol,schema,version" ||
    contract.schema !== "PRESENCE_AVALANCHE_OPERATION_CONTRACT_001" ||
    contract.implementation !== "AVALANCHE_IMPLEMENTATION_001" ||
    contract.protocol !== "PRESENCE_AVALANCHE_VM_001" || contract.version !== "1.0.0" ||
    !Array.isArray(contract.operations)
  ) throw new Error("unsupported shared VM operation contract");
  const operations = Object.create(null);
  for (const row of contract.operations) {
    if (!row || typeof row !== "object" || Array.isArray(row)) throw new Error("invalid VM operation classification");
    if (Object.keys(row).sort().join(",") !== "effect,operation,scope") throw new Error("invalid VM operation classification fields");
    if (!SUPPORTED_VM_OPERATIONS.includes(row.operation)) throw new Error(`unknown VM operation ${row.operation}`);
    if (Object.hasOwn(operations, row.operation)) throw new Error(`duplicate VM operation ${row.operation}`);
    if (!SCOPES.has(row.scope)) throw new Error(`unclassified VM operation ${row.operation}`);
    if (typeof row.effect !== "string" || !TOKEN.test(row.effect)) throw new Error(`missing VM effect for ${row.operation}`);
    operations[row.operation] = Object.freeze({ ...row });
  }
  for (const operation of SUPPORTED_VM_OPERATIONS) {
    if (!Object.hasOwn(operations, operation)) throw new Error(`missing supported VM operation ${operation}`);
  }
  return Object.freeze(operations);
}

const contractBytes = readFileSync(new URL("../../vm/protocol/operations.json", import.meta.url));
export const VM_OPERATION_CONTRACT_SHA256 = createHash("sha256").update(contractBytes).digest("hex");
export const VM_OPERATIONS = validateOperationContract(JSON.parse(contractBytes.toString("utf8")));

const rootContractBytes = readFileSync(new URL("../../protocol/operations.json", import.meta.url));
export const ROOT_OPERATION_CONTRACT_SHA256 = createHash("sha256").update(rootContractBytes).digest("hex");
const rootOperations = validateOperationContract({
  protocol: "PRESENCE_AVALANCHE_VM_001",
  ...JSON.parse(rootContractBytes.toString("utf8"))
});
for (const operation of SUPPORTED_VM_OPERATIONS) {
  if (rootOperations[operation].scope !== VM_OPERATIONS[operation].scope ||
      rootOperations[operation].effect !== VM_OPERATIONS[operation].effect) {
    throw new Error(`VM operation contracts disagree for ${operation}`);
  }
}

export function assertOperationScope(request, state) {
  const classification = VM_OPERATIONS[request.operation];
  if (!classification) throw new Error(`unknown VM operation ${request.operation}`);
  if (typeof request.locus_id !== "string") throw new Error("operation locus_id must be a string");
  const active = state.active_locus_id;
  if (typeof active !== "string") throw new Error("accepted active_locus_id must be a string");
  switch (classification.scope) {
    case "NEW_LOCUS":
      if (active || !request.locus_id || Object.hasOwn(state.loci, request.locus_id)) {
        throw new Error("NEW_LOCUS requires an unused proposed locus and no accepted active locus");
      }
      if (request.payload?.locus_id !== request.locus_id) {
        throw new Error("NEW_LOCUS requires matching proposed payload and transition locus IDs");
      }
      break;
    case "ACTIVE_LOCUS":
      if (!active || request.locus_id !== active) {
        throw new Error("ACTIVE_LOCUS requires the exact nonempty accepted active locus");
      }
      break;
    case "BODY_LOCAL":
      if (request.locus_id !== "") throw new Error("BODY_LOCAL requires an empty locus_id");
      break;
    case "DORMANT_BODY":
      if (request.locus_id !== "" || active || state.body.posture !== "DORMANT_P0" || state.body.presence_count !== 0) {
        throw new Error("DORMANT_BODY requires an empty locus_id, no active locus and exact VM dormancy");
      }
      break;
    case "TARGET_SCOPED": {
      const targetId = request.payload?.target_transition_id;
      const target = typeof targetId === "string" && Object.hasOwn(state.events, targetId) ? state.events[targetId] : null;
      if (!target || typeof target.locus_id !== "string" || request.locus_id !== target.locus_id) {
        throw new Error("TARGET_SCOPED requires the accepted target transition event locus");
      }
      break;
    }
    default:
      throw new Error(`unclassified VM operation ${request.operation}`);
  }
  return classification;
}

// The operation-scope contract is the single normative source for the scope
// classification and declared effect of every supported operation. This module
// reads the exact JSON file that the VM embeds at build time
// (contract/OPERATION_SCOPE_CONTRACT_001.json), so the FIELD and the VM cannot
// silently maintain divergent scope tables. Unknown operations and operations
// missing a scope classification fail closed.
import { readFileSync } from "node:fs";

export const OPERATION_SCOPE_CONTRACT_PATH = new URL(
  "../../contract/OPERATION_SCOPE_CONTRACT_001.json",
  import.meta.url
);

const contract = JSON.parse(readFileSync(OPERATION_SCOPE_CONTRACT_PATH, "utf8"));

if (contract.schema !== "PRESENCE_OPERATION_SCOPE_CONTRACT_001") {
  throw new Error("operation-scope contract schema is not PRESENCE_OPERATION_SCOPE_CONTRACT_001");
}
if (contract.unknown_operation_rule !== "FAIL_CLOSED" || contract.missing_scope_rule !== "FAIL_CLOSED") {
  throw new Error("operation-scope contract must declare FAIL_CLOSED rules");
}
if (!contract.operations || typeof contract.operations !== "object" || Object.keys(contract.operations).length === 0) {
  throw new Error("operation-scope contract declares no operations");
}

export const SCOPE_NEW_LOCUS = "NEW_LOCUS";
export const SCOPE_ACTIVE_LOCUS = "ACTIVE_LOCUS";
export const SCOPE_BODY_LOCAL = "BODY_LOCAL";
export const SCOPE_DORMANT_BODY = "DORMANT_BODY";
export const SCOPE_TARGET_SCOPED = "TARGET_SCOPED";

const KNOWN_SCOPES = new Set([
  SCOPE_NEW_LOCUS,
  SCOPE_ACTIVE_LOCUS,
  SCOPE_BODY_LOCAL,
  SCOPE_DORMANT_BODY,
  SCOPE_TARGET_SCOPED
]);

for (const [operation, entry] of Object.entries(contract.operations)) {
  if (!Object.hasOwn(contract.scope_classes ?? {}, entry.scope)) {
    throw new Error(`operation ${operation} declares undefined scope class ${entry.scope}`);
  }
  if (!KNOWN_SCOPES.has(entry.scope)) {
    throw new Error(`operation ${operation} declares unsupported scope class ${entry.scope}`);
  }
  if (typeof entry.effect !== "string" || !entry.effect) {
    throw new Error(`operation ${operation} declares no effect`);
  }
}

export const OPERATION_SCOPE_CONTRACT = Object.freeze(contract);

export const OPERATION_SCOPES = Object.freeze(
  Object.fromEntries(Object.entries(contract.operations).map(([operation, entry]) => [operation, entry.scope]))
);

export const OPERATION_EFFECTS = Object.freeze(
  Object.fromEntries(Object.entries(contract.operations).map(([operation, entry]) => [operation, entry.effect]))
);

export function scopeOf(operation) {
  const scope = OPERATION_SCOPES[operation];
  if (!scope) {
    throw new Error(`operation ${operation} has no scope classification in the operation-scope contract; failing closed`);
  }
  return scope;
}

// validateDraftLocus binds a draft request's locus_id to the shared contract
// before anything is signed or submitted.
export function validateDraftLocus(operation, locusId, state, payload) {
  if (typeof locusId !== "string") throw new Error("draft locus_id must be a string");
  const active = state.active_locus_id ?? "";
  const scope = scopeOf(operation);
  switch (scope) {
    case SCOPE_NEW_LOCUS:
      if (!locusId) throw new Error(`operation ${operation} must name the proposed new locus`);
      if (active && locusId === active) {
        throw new Error(`operation ${operation} must not pretend the proposed locus is the already active locus`);
      }
      if (state.loci && Object.hasOwn(state.loci, locusId)) {
        throw new Error(`operation ${operation} must not reuse an existing locus ID`);
      }
      break;
    case SCOPE_ACTIVE_LOCUS:
      if (!active) throw new Error(`operation ${operation} requires an active locus`);
      if (locusId !== active) throw new Error(`operation ${operation} must name the exact accepted active locus`);
      break;
    case SCOPE_BODY_LOCAL:
      if (locusId !== "") throw new Error(`operation ${operation} is body-local and requires an empty locus_id`);
      break;
    case SCOPE_DORMANT_BODY:
      if (locusId !== "") throw new Error(`operation ${operation} is dormant-body scoped and requires an empty locus_id`);
      if (active) throw new Error(`operation ${operation} requires no active locus`);
      break;
    case SCOPE_TARGET_SCOPED: {
      const targetId = payload?.target_transition_id;
      if (typeof targetId !== "string" || !targetId) {
        throw new Error(`operation ${operation} requires payload.target_transition_id`);
      }
      const target = state.events?.[targetId];
      if (!target) {
        throw new Error(`operation ${operation} target transition is not in accepted state; failing closed`);
      }
      if (locusId !== (target.locus_id ?? "")) {
        throw new Error(`operation ${operation} must bind the exact scope of its target transition`);
      }
      break;
    }
    default:
      throw new Error(`operation ${operation} has unsupported scope class ${scope}; failing closed`);
  }
}

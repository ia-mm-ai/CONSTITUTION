import { operationDefinition } from "./constants.js";

export function assertOperationScope({ operation, locusId, payload, state }) {
  const definition = operationDefinition(operation);
  switch (definition.scope) {
    case "NEW_LOCUS":
      if (!locusId || payload?.locus_id !== locusId) {
        throw new Error("NEW_LOCUS operation must name its proposed payload locus");
      }
      break;
    case "ACTIVE_LOCUS":
      if (!state.active_locus_id || locusId !== state.active_locus_id) {
        throw new Error(`ACTIVE_LOCUS operation must name accepted locus ${state.active_locus_id}`);
      }
      break;
    case "BODY_LOCAL":
      if (locusId !== "") throw new Error("BODY_LOCAL operation requires an empty transition locus_id");
      break;
    case "DORMANT_BODY":
      if (locusId !== "" || state.active_locus_id !== "" || state.body?.posture !== "DORMANT_P0") {
        throw new Error("DORMANT_BODY operation requires no active locus and exact DORMANT_P0 posture");
      }
      break;
    case "TARGET_SCOPED": {
      const target = state.events?.[payload?.target_transition_id];
      if (!target) throw new Error("TARGET_SCOPED operation requires an existing target transition");
      if (locusId !== target.locus_id) {
        throw new Error(`TARGET_SCOPED operation must name target locus ${target.locus_id}`);
      }
      break;
    }
    default:
      throw new Error(`operation ${operation} has unknown scope ${definition.scope}`);
  }
  return definition;
}

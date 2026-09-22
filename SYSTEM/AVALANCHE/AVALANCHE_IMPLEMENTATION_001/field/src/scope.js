import { assertOperationScope as assertSharedOperationScope } from "./operations.js";

export function assertOperationScope({ operation, locusId, payload, state }) {
  return assertSharedOperationScope({ operation, locus_id: locusId, payload }, state);
}

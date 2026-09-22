import { CapabilityBroker } from "../src/broker.js";
import { createRuntime } from "../src/runtime.js";

const runtime = await createRuntime(process.argv[2]);
const broker = new CapabilityBroker(runtime.engine);

// The real tool must exist behind this handler. Calling the tool directly would
// bypass the medium and void its enforcement claim.
broker.register("AI_BOUNDED_TOOL_BROKER", async (request, decision) => ({
  accepted_state_commitment: decision.accepted_state_commitment,
  result: await yourBoundedTool(request)
}));

async function yourBoundedTool(_request) {
  throw new Error("replace with one inspectable, bounded tool implementation");
}

export { broker };

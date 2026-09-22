// CapabilityBroker binds a named local capability to a handler that may only run
// when the FIELD engine reaches a supporting disposition for it. The broker
// itself grants nothing; it defers entirely to the engine's fail-closed guard.
export class CapabilityBroker {
  constructor(engine) {
    this.engine = engine;
    this.handlers = new Map();
  }

  register(capability, handler) {
    if (this.handlers.has(capability)) throw new Error(`capability ${capability} already has a handler`);
    if (typeof handler !== "function") throw new Error("capability handler must be a function");
    this.handlers.set(capability, handler);
    return this;
  }

  async invoke(capability, input, { expectedStateCommitment } = {}) {
    const handler = this.handlers.get(capability);
    if (!handler) throw new Error(`no controlled handler is registered for ${capability}`);
    return this.engine.guard(
      capability,
      (decision) => handler(input, decision),
      { expectedStateCommitment }
    );
  }
}

// Non-collapse law. Each rule names one distinct thing the FIELD must never
// silently treat as another. The predecessor carried the first block; this
// implementation keeps it and adds the five explicit collapses named in the
// PRESENCE_AVALANCHE_FIELD_001 requirement.
//
// A collapse is a category error: reading a token balance as if it were
// capacity, a wallet as if it were a participant, a validator as if it were an
// authority, schema conformance as if it were formation, or API access as if it
// were presence. Every one is refused in code, not merely documented.

export const NON_COLLAPSE_RULES = Object.freeze({
  TOKEN_BALANCE_IS_NOT_CAPACITY:
    "A token or balance is an economic quantity. Capacity is the signed structural/situated account in accepted VM state. Never derive one from the other.",
  WALLET_KEY_IS_NOT_PARTICIPANT:
    "A wallet or key is signing material. A participant is an accepted actor binding in VM state on a controlled surface. Holding a key is not being a participant.",
  VALIDATOR_IS_NOT_AUTHORITY:
    "A consensus validator secures block production. Authority is an accepted, capability-scoped actor in VM state. Running a validator grants no host authority.",
  SCHEMA_CONFORMANCE_IS_NOT_FORMATION:
    "Passing a schema check proves representation shape only. Formation is a substantive act reserved to the human core and the VM. Conformance never forms anything.",
  API_ACCESS_IS_NOT_PRESENCE:
    "Reaching the FIELD or VM HTTP surface is connectivity. Presence is a current accepted entry binding in VM state. Access is not admission and not presence."
});

// The FIELD-facing effect ceilings that correspond to the five collapses. These
// are asserted on emitted observations so no downstream reader can quietly
// upgrade a surface signal into a substantive claim.
export const NON_COLLAPSE_CEILING = Object.freeze([
  "NO_TOKEN_BALANCE_AS_CAPACITY",
  "NO_WALLET_KEY_AS_PARTICIPANT",
  "NO_VALIDATOR_AS_AUTHORITY",
  "NO_SCHEMA_CONFORMANCE_AS_FORMATION",
  "NO_API_ACCESS_AS_PRESENCE"
]);

const COLLAPSE_KINDS = new Set(Object.keys(NON_COLLAPSE_RULES));

export class CollapseError extends Error {
  constructor(kind, message) {
    super(message);
    this.name = "CollapseError";
    this.code = "NON_COLLAPSE_VIOLATION";
    this.kind = kind;
  }
}

// refuseCollapse is called at every boundary where a caller might try to treat
// one category as another. It always throws; there is no permissive path.
export function refuseCollapse(kind) {
  const rule = NON_COLLAPSE_RULES[kind];
  if (!rule) throw new Error(`unknown non-collapse kind ${kind}`);
  throw new CollapseError(kind, `refused category collapse: ${rule}`);
}

// capacityFromAcceptedState derives capacity ONLY from the signed capacity
// account in accepted VM state. Any attempt to pass token/balance material is a
// collapse and is refused before a number is produced.
export function capacityFromAcceptedState(capacityAccount) {
  if (!capacityAccount || typeof capacityAccount !== "object" || Array.isArray(capacityAccount)) {
    throw new Error("capacity account must be an object drawn from accepted VM state");
  }
  for (const forbidden of ["token", "tokens", "balance", "wallet_balance", "supply"]) {
    if (Object.hasOwn(capacityAccount, forbidden)) refuseCollapse("TOKEN_BALANCE_IS_NOT_CAPACITY");
  }
  return {
    structural_ceiling_units: capacityAccount.structural_ceiling_units,
    available_units: capacityAccount.available_units
  };
}

// participantBindingFromState refuses to treat mere key possession as
// participation. A binding exists only when accepted state records the key for
// the actor.
export function participantBindingFromState(state, actorId) {
  const bound = state?.actor_keys?.[actorId];
  return bound ? { actor_id: actorId, accepted_key_binding: bound } : null;
}

export function isCollapseKind(kind) {
  return COLLAPSE_KINDS.has(kind);
}

package main

// The operation-scope contract is the single normative source for the scope
// classification and declared effect of every supported operation. The exact
// JSON file embedded here is the same file the FIELD runtime loads at import
// time (contract/OPERATION_SCOPE_CONTRACT_001.json), so the VM and FIELD
// cannot silently maintain divergent scope tables. Unknown operations and
// operations missing a scope classification fail closed.

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed contract/OPERATION_SCOPE_CONTRACT_001.json
var operationScopeContractBytes []byte

const operationScopeContractSchema = "PRESENCE_OPERATION_SCOPE_CONTRACT_001"

const (
	scopeNewLocus     = "NEW_LOCUS"
	scopeActiveLocus  = "ACTIVE_LOCUS"
	scopeBodyLocal    = "BODY_LOCAL"
	scopeDormantBody  = "DORMANT_BODY"
	scopeTargetScoped = "TARGET_SCOPED"
)

type contractOperation struct {
	Scope  string `json:"scope"`
	Effect string `json:"effect"`
}

type operationScopeContract struct {
	Schema               string                       `json:"schema"`
	Version              string                       `json:"version"`
	Implementation       string                       `json:"implementation"`
	VMProtocol           string                       `json:"vm_protocol"`
	Authority            string                       `json:"authority"`
	UnknownOperationRule string                       `json:"unknown_operation_rule"`
	MissingScopeRule     string                       `json:"missing_scope_rule"`
	ScopeClasses         map[string]json.RawMessage   `json:"scope_classes"`
	Operations           map[string]contractOperation `json:"operations"`
}

var (
	scopeContract    = mustLoadOperationScopeContract()
	operationEffects = contractEffects(scopeContract)
	operationScopes  = contractScopes(scopeContract)
)

func mustLoadOperationScopeContract() *operationScopeContract {
	contract := new(operationScopeContract)
	if err := decodeStrict(operationScopeContractBytes, contract); err != nil {
		panic(fmt.Errorf("operation-scope contract is invalid: %w", err))
	}
	if contract.Schema != operationScopeContractSchema {
		panic(fmt.Errorf("operation-scope contract schema %q is not %s", contract.Schema, operationScopeContractSchema))
	}
	if contract.UnknownOperationRule != "FAIL_CLOSED" || contract.MissingScopeRule != "FAIL_CLOSED" {
		panic(fmt.Errorf("operation-scope contract must declare FAIL_CLOSED rules"))
	}
	if len(contract.Operations) == 0 {
		panic(fmt.Errorf("operation-scope contract declares no operations"))
	}
	for operation, entry := range contract.Operations {
		if _, exists := contract.ScopeClasses[entry.Scope]; !exists {
			panic(fmt.Errorf("operation %s declares undefined scope class %q", operation, entry.Scope))
		}
		switch entry.Scope {
		case scopeNewLocus, scopeActiveLocus, scopeBodyLocal, scopeDormantBody, scopeTargetScoped:
		default:
			panic(fmt.Errorf("operation %s declares unsupported scope class %q", operation, entry.Scope))
		}
		if entry.Effect == "" {
			panic(fmt.Errorf("operation %s declares no effect", operation))
		}
	}
	return contract
}

func contractEffects(contract *operationScopeContract) map[string]string {
	effects := make(map[string]string, len(contract.Operations))
	for operation, entry := range contract.Operations {
		effects[operation] = entry.Effect
	}
	return effects
}

func contractScopes(contract *operationScopeContract) map[string]string {
	scopes := make(map[string]string, len(contract.Operations))
	for operation, entry := range contract.Operations {
		scopes[operation] = entry.Scope
	}
	return scopes
}

// operationScope fails closed: an operation without an explicit scope
// classification in the shared contract is not executable.
func operationScope(operation string) (string, error) {
	scope, exists := operationScopes[operation]
	if !exists {
		return "", fmt.Errorf("operation %q has no scope classification in the operation-scope contract; failing closed", operation)
	}
	return scope, nil
}

// validateScopeShape checks the state-independent locus shape of a transition
// against the shared contract. It runs during syntax validation, before any
// signature or state work.
func validateScopeShape(operation, locusID string) error {
	scope, err := operationScope(operation)
	if err != nil {
		return err
	}
	switch scope {
	case scopeNewLocus, scopeActiveLocus:
		if locusID == "" {
			return fmt.Errorf("operation %s has %s scope and requires a nonempty transition locus_id", operation, scope)
		}
	case scopeBodyLocal, scopeDormantBody:
		if locusID != "" {
			return fmt.Errorf("operation %s has %s scope and requires an empty transition locus_id", operation, scope)
		}
	case scopeTargetScoped:
		// The exact locus is bound to the target transition's recorded scope
		// against accepted state; both empty and nonempty shapes are lawful.
	default:
		return fmt.Errorf("operation %s has unsupported scope class %q; failing closed", operation, scope)
	}
	return nil
}

// enforceOperationScope applies the state-bound scope rules of the shared
// contract before any operation handler runs.
func (s *RuntimeState) enforceOperationScope(u *UnsignedTransition) error {
	scope, err := operationScope(u.Operation)
	if err != nil {
		return err
	}
	switch scope {
	case scopeNewLocus:
		if u.LocusID == "" {
			return fmt.Errorf("operation %s must name the proposed new locus", u.Operation)
		}
		if s.ActiveLocusID != "" && u.LocusID == s.ActiveLocusID {
			return fmt.Errorf("operation %s must not pretend the proposed locus is the already active locus", u.Operation)
		}
	case scopeActiveLocus:
		if s.ActiveLocusID == "" {
			return fmt.Errorf("operation %s requires an active locus", u.Operation)
		}
		if u.LocusID != s.ActiveLocusID {
			return fmt.Errorf("operation %s must name the exact accepted active locus", u.Operation)
		}
	case scopeBodyLocal:
		if u.LocusID != "" {
			return fmt.Errorf("operation %s is body-local and requires an empty transition locus_id", u.Operation)
		}
	case scopeDormantBody:
		if u.LocusID != "" {
			return fmt.Errorf("operation %s is dormant-body scoped and requires an empty transition locus_id", u.Operation)
		}
		if s.ActiveLocusID != "" {
			return fmt.Errorf("operation %s requires no active locus", u.Operation)
		}
	case scopeTargetScoped:
		// The operation handler binds the exact recorded scope of its target.
	default:
		return fmt.Errorf("operation %s has unsupported scope class %q; failing closed", u.Operation, scope)
	}
	return nil
}

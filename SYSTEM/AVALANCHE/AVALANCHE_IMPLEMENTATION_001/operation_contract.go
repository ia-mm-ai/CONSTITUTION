package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	scopeNewLocus    = "NEW_LOCUS"
	scopeActiveLocus = "ACTIVE_LOCUS"
	scopeBodyLocal   = "BODY_LOCAL"
	scopeDormantBody = "DORMANT_BODY"
	scopeTarget      = "TARGET_SCOPED"
)

type operationDefinition struct {
	Operation string `json:"operation"`
	Scope     string `json:"scope"`
	Effect    string `json:"effect"`
}

type operationContractDocument struct {
	Schema         string                `json:"schema"`
	Implementation string                `json:"implementation"`
	Version        string                `json:"version"`
	Operations     []operationDefinition `json:"operations"`
}

//go:embed protocol/operations.json
var operationContractBytes []byte

var (
	operationDefinitions map[string]operationDefinition
	operationEffects     map[string]string
)

func init() {
	var document operationContractDocument
	if err := decodeStrict(operationContractBytes, &document); err != nil {
		panic(fmt.Errorf("decode embedded operation contract: %w", err))
	}
	if document.Schema != "PRESENCE_AVALANCHE_OPERATION_CONTRACT_001" ||
		document.Implementation != "AVALANCHE_IMPLEMENTATION_001" ||
		document.Version != "1.0.0" {
		panic("embedded operation contract has an unsupported identity")
	}

	knownOperations := map[string]struct{}{
		opBound: {}, opPresentForm: {}, opGateDisposition: {}, opEnter: {},
		opCheckpointDeparture: {}, opReenter: {}, opObserveCrossing: {},
		opMatterDisposition: {}, opRecordEmergence: {}, opCorrect: {},
		opExit: {}, opClose: {}, opIncorporateResidue: {}, opDeclareCapacity: {},
		opPulse: {}, opReclaimOffer: {}, opRegisterAuthority: {},
		opExhaustFormation: {}, opProposeSuccessor: {}, opAttestSuccessor: {},
		opActivateSuccessor: {},
	}
	validScopes := map[string]struct{}{
		scopeNewLocus: {}, scopeActiveLocus: {}, scopeBodyLocal: {},
		scopeDormantBody: {}, scopeTarget: {},
	}
	operationDefinitions = make(map[string]operationDefinition, len(document.Operations))
	operationEffects = make(map[string]string, len(document.Operations))
	for _, definition := range document.Operations {
		if _, known := knownOperations[definition.Operation]; !known {
			panic(fmt.Sprintf("operation contract declares unsupported operation %q", definition.Operation))
		}
		if _, duplicate := operationDefinitions[definition.Operation]; duplicate {
			panic(fmt.Sprintf("operation contract repeats operation %q", definition.Operation))
		}
		if _, valid := validScopes[definition.Scope]; !valid || definition.Effect == "" {
			panic(fmt.Sprintf("operation contract has invalid definition for %q", definition.Operation))
		}
		operationDefinitions[definition.Operation] = definition
		operationEffects[definition.Operation] = definition.Effect
	}
	for operation := range knownOperations {
		if _, declared := operationDefinitions[operation]; !declared {
			panic(fmt.Sprintf("operation %q has no scope contract", operation))
		}
	}
}

func validateOperationScope(state *RuntimeState, operation, locusID string, payload json.RawMessage) error {
	definition, exists := operationDefinitions[operation]
	if !exists {
		return fmt.Errorf("unsupported operation %q", operation)
	}
	switch definition.Scope {
	case scopeNewLocus:
		if locusID == "" {
			return errors.New("NEW_LOCUS operation requires its proposed locus_id")
		}
		var bound BoundPayload
		if err := decodeStrict(payload, &bound); err != nil {
			return fmt.Errorf("decode NEW_LOCUS payload: %w", err)
		}
		if bound.LocusID != locusID {
			return errors.New("NEW_LOCUS transition must name its payload locus")
		}
	case scopeActiveLocus:
		if state.ActiveLocusID == "" {
			return errors.New("ACTIVE_LOCUS operation requires an accepted active locus")
		}
		if locusID != state.ActiveLocusID {
			return fmt.Errorf("ACTIVE_LOCUS operation must name accepted locus %q", state.ActiveLocusID)
		}
	case scopeBodyLocal:
		if locusID != "" {
			return errors.New("BODY_LOCAL operation requires an empty transition locus_id")
		}
	case scopeDormantBody:
		if locusID != "" {
			return errors.New("DORMANT_BODY operation requires an empty transition locus_id")
		}
		if state.ActiveLocusID != "" || state.derivedBody().Posture != "DORMANT_P0" {
			return errors.New("DORMANT_BODY operation requires no active locus and exact DORMANT_P0 posture")
		}
	case scopeTarget:
		var correction CorrectPayload
		if err := decodeStrict(payload, &correction); err != nil {
			return fmt.Errorf("decode TARGET_SCOPED payload: %w", err)
		}
		target, exists := state.Events[correction.TargetTransitionID]
		if !exists {
			return errors.New("TARGET_SCOPED operation requires an existing target transition")
		}
		if locusID != target.LocusID {
			return fmt.Errorf("TARGET_SCOPED operation must name target locus %q", target.LocusID)
		}
	default:
		return fmt.Errorf("operation %q has unknown scope %q", operation, definition.Scope)
	}
	return nil
}

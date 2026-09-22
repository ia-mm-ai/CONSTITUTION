package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	scopeNewLocus    = "NEW_LOCUS"
	scopeActiveLocus = "ACTIVE_LOCUS"
	scopeBodyLocal   = "BODY_LOCAL"
	scopeDormantBody = "DORMANT_BODY"
	scopeTarget      = "TARGET_SCOPED"
)

//go:embed protocol/operations.json
var protocolContractJSON []byte

var protocolContractSHA256 = func() string {
	digest := sha256.Sum256(protocolContractJSON)
	return hex.EncodeToString(digest[:])
}()

type operationContract struct {
	Operation string `json:"operation"`
	Effect    string `json:"effect"`
	Scope     string `json:"scope"`
}

type scopeClassContract struct {
	TransitionLocusRule string `json:"transition_locus_rule"`
	ActiveLocusRule     string `json:"active_locus_rule"`
	Description         string `json:"description"`
}

type protocolContract struct {
	Schema               string                        `json:"schema"`
	Implementation       string                        `json:"implementation"`
	Protocol             string                        `json:"protocol"`
	Version              string                        `json:"version"`
	UnknownOperationRule string                        `json:"unknown_operation_rule"`
	MissingScopeRule     string                        `json:"missing_scope_rule"`
	ScopeClasses         map[string]scopeClassContract `json:"scope_classes"`
	Operations           []operationContract           `json:"operations"`
}

var supportedOperations = []string{
	opBound, opPresentForm, opGateDisposition, opEnter, opCheckpointDeparture,
	opReenter, opObserveCrossing, opMatterDisposition, opRecordEmergence, opCorrect,
	opExit, opClose, opIncorporateResidue, opDeclareCapacity, opPulse, opReclaimOffer,
	opRegisterAuthority, opExhaustFormation, opProposeSuccessor, opAttestSuccessor,
	opActivateSuccessor,
}

var operationContracts = mustLoadOperationContracts()

var operationEffects = func() map[string]string {
	effects := make(map[string]string, len(operationContracts))
	for operation, contract := range operationContracts {
		effects[operation] = contract.Effect
	}
	return effects
}()

func loadOperationContracts(data []byte) (map[string]operationContract, error) {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return nil, err
	}
	var document protocolContract
	if err := decodeStrict(data, &document); err != nil {
		return nil, err
	}
	if document.Schema != "PRESENCE_AVALANCHE_OPERATION_CONTRACT_001" ||
		document.Implementation != implementationID || document.Protocol != vmIDDomain ||
		document.Version != implementationVersion {
		return nil, errors.New("operation contract identity mismatch")
	}
	if document.UnknownOperationRule != "FAIL_CLOSED" || document.MissingScopeRule != "FAIL_CLOSED" {
		return nil, errors.New("operation contract must declare FAIL_CLOSED unknown-operation and missing-scope rules")
	}
	if len(document.ScopeClasses) != 5 {
		return nil, errors.New("operation contract scope classes are not exhaustive")
	}
	for _, scope := range []string{scopeNewLocus, scopeActiveLocus, scopeBodyLocal, scopeDormantBody, scopeTarget} {
		class, exists := document.ScopeClasses[scope]
		if !exists || class.TransitionLocusRule == "" || class.ActiveLocusRule == "" || class.Description == "" {
			return nil, fmt.Errorf("operation contract scope class %q is missing or incomplete", scope)
		}
	}
	contracts := make(map[string]operationContract, len(document.Operations))
	for _, operation := range document.Operations {
		if !contains(supportedOperations, operation.Operation) || operation.Effect == "" {
			return nil, fmt.Errorf("unknown or unclassified operation %q", operation.Operation)
		}
		if _, duplicate := contracts[operation.Operation]; duplicate {
			return nil, fmt.Errorf("duplicate operation %q", operation.Operation)
		}
		switch operation.Scope {
		case scopeNewLocus, scopeActiveLocus, scopeBodyLocal, scopeDormantBody, scopeTarget:
		default:
			return nil, fmt.Errorf("unknown scope for operation %q", operation.Operation)
		}
		contracts[operation.Operation] = operation
	}
	if len(contracts) != len(supportedOperations) {
		return nil, errors.New("operation contract is not exhaustive")
	}
	return contracts, nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var value func() error
	value = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token {
		case json.Delim('{'):
			seen := make(map[string]bool)
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok || seen[name] {
					return fmt.Errorf("duplicate or invalid JSON key %v", key)
				}
				seen[name] = true
				if err := value(); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		case json.Delim('['):
			for decoder.More() {
				if err := value(); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		default:
			return nil
		}
	}
	if err := value(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("unexpected trailing JSON data")
	}
	return nil
}

func mustLoadOperationContracts() map[string]operationContract {
	contracts, err := loadOperationContracts(protocolContractJSON)
	if err != nil {
		panic(err)
	}
	return contracts
}

func validateScopeSyntax(u *UnsignedTransition) error {
	contract, exists := operationContracts[u.Operation]
	if !exists {
		return fmt.Errorf("unsupported operation %q", u.Operation)
	}
	if u.Effect != contract.Effect {
		return fmt.Errorf("operation %s requires explicit effect %s", u.Operation, contract.Effect)
	}
	switch contract.Scope {
	case scopeNewLocus, scopeActiveLocus:
		return requireSafeID("locus_id", u.LocusID)
	case scopeBodyLocal, scopeDormantBody:
		if u.LocusID != "" {
			return fmt.Errorf("%s requires an empty locus_id", contract.Scope)
		}
	case scopeTarget:
		if u.LocusID != "" {
			return requireSafeID("locus_id", u.LocusID)
		}
	default:
		return fmt.Errorf("unclassified operation %q", u.Operation)
	}
	return nil
}

func (s *RuntimeState) validateOperationScope(u *UnsignedTransition) error {
	if err := validateScopeSyntax(u); err != nil {
		return err
	}
	switch operationContracts[u.Operation].Scope {
	case scopeNewLocus:
		if s.ActiveLocusID != "" {
			return errors.New("a locus is already active")
		}
		if _, used := s.Loci[u.LocusID]; used {
			return errors.New("locus ID has already been used")
		}
		payload, err := decodePayload[BoundPayload](u.Payload)
		if err != nil {
			return err
		}
		if payload.LocusID != u.LocusID {
			return errors.New("BOUND payload and transition locus IDs differ")
		}
	case scopeActiveLocus:
		_, err := s.activeLocus(u.LocusID)
		return err
	case scopeBodyLocal:
		return nil
	case scopeDormantBody:
		if s.ActiveLocusID != "" || s.derivedBody().Posture != "DORMANT_P0" {
			return errors.New("operation requires a dormant body checkpoint with no open locus")
		}
	case scopeTarget:
		payload, err := decodePayload[CorrectPayload](u.Payload)
		if err != nil {
			return err
		}
		target, exists := s.Events[payload.TargetTransitionID]
		if !exists {
			return errors.New("correction target does not exist")
		}
		if err := validateScopeSyntax(&UnsignedTransition{
			Operation: target.Operation, Effect: target.Effect, LocusID: target.LocusID,
		}); err != nil {
			return fmt.Errorf("correction target is not a classified protocol event: %w", err)
		}
		if target.LocusID != u.LocusID {
			return errors.New("CORRECT locus_id must equal the exact target event locus_id")
		}
	default:
		return fmt.Errorf("unclassified operation %q", u.Operation)
	}
	return nil
}

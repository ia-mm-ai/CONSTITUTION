package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ava-labs/avalanchego/database/memdb"
)

func TestOperationContractIsExhaustiveAndEmbeddedVerbatim(t *testing.T) {
	raw, err := os.ReadFile("protocol/operations.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, protocolContractJSON) {
		t.Fatal("runtime contract differs from the shared normative JSON")
	}
	want := make(map[string]bool)
	declarations, err := parser.ParseFile(token.NewFileSet(), "transition.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(declarations, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, name := range spec.Names {
			if strings.HasPrefix(name.Name, "op") {
				literal, ok := spec.Values[i].(*ast.BasicLit)
				if !ok {
					t.Fatalf("operation %s is not an explicit string", name.Name)
				}
				operation, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				want[operation] = true
			}
		}
		return true
	})
	got := make(map[string]bool)
	for name, entry := range operationContracts {
		got[name] = true
		if operationEffects[name] != entry.Effect {
			t.Fatalf("effect drift for %s", name)
		}
	}
	if !reflect.DeepEqual(got, want) || len(want) != 21 {
		t.Fatalf("operation inventory mismatch: contract=%v declarations=%v", got, want)
	}
	for filename, function := range map[string]string{
		"operations.go": "applyOperation",
		"transition.go": "validateAndCanonicalizePayload",
	} {
		file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		handled := make(map[string]bool)
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Name.Name != function {
				continue
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				clause, ok := node.(*ast.CaseClause)
				if ok {
					for _, expression := range clause.List {
						if identifier, ok := expression.(*ast.Ident); ok {
							for _, declaration := range declarations.Decls {
								group, ok := declaration.(*ast.GenDecl)
								if !ok || group.Tok != token.CONST {
									continue
								}
								for _, spec := range group.Specs {
									value := spec.(*ast.ValueSpec)
									if value.Names[0].Name == identifier.Name {
										operation, _ := strconv.Unquote(value.Values[0].(*ast.BasicLit).Value)
										handled[operation] = true
									}
								}
							}
						}
					}
				}
				return true
			})
		}
		if !reflect.DeepEqual(got, handled) {
			t.Fatalf("%s dispatch and contract differ: %v", function, handled)
		}
	}
}

func TestOperationContractRejectsIncompleteOrAmbiguousInventory(t *testing.T) {
	duplicateKey := bytes.Replace(protocolContractJSON, []byte(`"scope": "NEW_LOCUS"`), []byte(`"scope": "BODY_LOCAL", "scope": "NEW_LOCUS"`), 1)
	if _, err := loadOperationContracts(duplicateKey); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	mutations := map[string]func(*protocolContract){
		"missing": func(c *protocolContract) { c.Operations = c.Operations[1:] },
		"duplicate": func(c *protocolContract) {
			c.Operations = append(c.Operations, c.Operations[0])
		},
		"unknown":      func(c *protocolContract) { c.Operations[0].Operation = "UNCLASSIFIED" },
		"empty effect": func(c *protocolContract) { c.Operations[0].Effect = "" },
		"empty scope":  func(c *protocolContract) { c.Operations[0].Scope = "" },
		"unknown scope": func(c *protocolContract) {
			c.Operations[0].Scope = "DEFAULT"
		},
		"wrong identity": func(c *protocolContract) { c.Protocol = "OTHER" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var document protocolContract
			if err := json.Unmarshal(protocolContractJSON, &document); err != nil {
				t.Fatal(err)
			}
			mutate(&document)
			raw, _ := json.Marshal(document)
			if _, err := loadOperationContracts(raw); err == nil {
				t.Fatal("invalid contract accepted")
			}
		})
	}
}

func TestTargetScopedDraftsResolveHistoricalAndBodyTargets(t *testing.T) {
	vm, _, genesis, _ := initializeTestVM(t, memdb.New())
	state := vm.genesisBlock.postState
	state.ActiveLocusID = "CURRENT"
	state.Loci["CURRENT"] = &LocusState{LocusID: "CURRENT", Phase: "OPEN", CapacityCeilingUnits: 100}
	for _, locus := range []string{"HISTORICAL", ""} {
		id := digestText("event " + locus)
		operation := opBound
		if locus == "" {
			operation = opDeclareCapacity
		}
		state.Events[id] = EventRecord{
			TransitionID: id, ActorID: genesis.Locality.ID, LocusID: locus,
			Operation: operation, Effect: operationEffects[operation],
		}
		raw, _ := json.Marshal(CorrectPayload{
			TargetTransitionID: id, ReplacementCommitment: digestText("replacement"), ReasonSHA256: digestText("reason"),
		})
		request := &TransitionDraftRequest{
			Operation: opCorrect, ActorID: genesis.Locality.ID, ActorPublicKey: genesis.Locality.Authority.PublicKey,
			LocusID: locus, ObservedAt: 1, Payload: raw,
		}
		if _, err := vm.draftTransition(request); err != nil {
			t.Fatalf("exact target scope rejected for %q: %v", locus, err)
		}
		for _, wrong := range []string{"CURRENT", "FOREIGN"} {
			request.LocusID = wrong
			if _, err := vm.draftTransition(request); err == nil {
				t.Fatalf("draft accepted %q instead of target scope %q", wrong, locus)
			}
		}
	}
}

func TestAllOperationScopesFailClosed(t *testing.T) {
	_, genesis, _ := testGenesis(t)
	dormant, err := initialRuntimeState(genesis)
	if err != nil {
		t.Fatal(err)
	}
	active, _ := dormant.clone()
	active.ActiveLocusID = "ACTIVE"
	active.Loci["ACTIVE"] = &LocusState{
		LocusID: "ACTIVE", Phase: "OPEN", CapacityCeilingUnits: 100,
	}
	targetID := digestText("target")
	active.Events[targetID] = EventRecord{
		LocusID: "HISTORICAL", Operation: opBound, Effect: operationEffects[opBound],
	}
	groups := map[string][]string{
		scopeNewLocus: {opBound},
		scopeActiveLocus: {
			opPresentForm, opGateDisposition, opEnter, opCheckpointDeparture, opReenter,
			opObserveCrossing, opMatterDisposition, opRecordEmergence, opExit, opClose, opReclaimOffer,
		},
		scopeBodyLocal: {
			opIncorporateResidue, opDeclareCapacity, opPulse, opRegisterAuthority,
			opProposeSuccessor, opAttestSuccessor,
		},
		scopeDormantBody: {opExhaustFormation, opActivateSuccessor},
		scopeTarget:      {opCorrect},
	}
	checked := 0
	for scope, operations := range groups {
		for _, operation := range operations {
			checked++
			t.Run(operation, func(t *testing.T) {
				contract := operationContracts[operation]
				if contract.Scope != scope {
					t.Fatalf("wrong normative scope: %s", contract.Scope)
				}
				u := &UnsignedTransition{Operation: operation, Effect: contract.Effect}
				validState := active
				switch scope {
				case scopeNewLocus:
					validState = dormant
					u.LocusID = "FRESH"
					u.Payload, _ = json.Marshal(BoundPayload{LocusID: "FRESH"})
				case scopeActiveLocus:
					u.LocusID = "ACTIVE"
				case scopeDormantBody:
					validState = dormant
				case scopeTarget:
					u.LocusID = "HISTORICAL"
					u.Payload, _ = json.Marshal(CorrectPayload{TargetTransitionID: targetID})
				}
				if err := validState.validateOperationScope(u); err != nil {
					t.Fatalf("valid scope rejected: %v", err)
				}
				invalid := *u
				switch scope {
				case scopeNewLocus, scopeActiveLocus:
					invalid.LocusID = ""
				case scopeTarget:
					invalid.LocusID = "ACTIVE"
				default:
					invalid.LocusID = "ACTIVE"
				}
				if err := validState.validateOperationScope(&invalid); err == nil {
					t.Fatal("invalid scope accepted")
				}
				if err := validState.applyOperation(genesis, &Transition{Unsigned: invalid}, "invalid"); err == nil {
					t.Fatal("application bypassed scope guard")
				}
				if scope == scopeActiveLocus {
					if err := dormant.validateOperationScope(u); err == nil {
						t.Fatal("active operation accepted without an active locus")
					}
					invalid.LocusID = "FOREIGN"
					if err := active.validateOperationScope(&invalid); err == nil {
						t.Fatal("foreign locus accepted")
					}
				}
				if scope == scopeDormantBody || scope == scopeNewLocus {
					if err := active.validateOperationScope(u); err == nil {
						t.Fatal("operation accepted with open locus")
					}
				}
				if scope == scopeBodyLocal {
					if err := dormant.validateOperationScope(u); err != nil {
						t.Fatalf("body-local scope incorrectly requires an open locus: %v", err)
					}
				}
			})
		}
	}
	if checked != len(operationContracts) {
		t.Fatal("scope test inventory is incomplete")
	}
	if err := dormant.validateOperationScope(&UnsignedTransition{Operation: "UNKNOWN"}); err == nil {
		t.Fatal("unknown operation received a default scope")
	}
}

func TestDormantScopeRequiresExactDerivedPosture(t *testing.T) {
	_, genesis, _ := testGenesis(t)
	for _, operation := range []string{opExhaustFormation, opActivateSuccessor} {
		for _, posture := range []string{"presence", "deficit", "succession"} {
			t.Run(operation+"/"+posture, func(t *testing.T) {
				state, _ := initialRuntimeState(genesis)
				switch posture {
				case "presence":
					state.Loci["HISTORY"] = &LocusState{
						Entries: map[string]EntryState{"actor": {Status: "PRESENT"}},
					}
				case "deficit":
					state.Capacity.DeclaredActualUnits = 0
				case "succession":
					state.SuccessorBoundary = &SuccessorBoundary{}
				}
				u := &UnsignedTransition{Operation: operation, Effect: operationEffects[operation]}
				if err := state.validateOperationScope(u); err == nil {
					t.Fatalf("derived %s posture was mistaken for dormancy", state.derivedBody().Posture)
				}
			})
		}
	}
}

func TestDraftAndConsensusScopesWithActiveLocus(t *testing.T) {
	vm, _, genesis, keys := initializeTestVM(t, memdb.New())
	draft := func(operation, locus string, payload any) (*UnsignedTransition, error) {
		raw, _ := json.Marshal(payload)
		return vm.draftTransition(&TransitionDraftRequest{
			Operation: operation, ActorID: genesis.Locality.ID, ActorPublicKey: genesis.Locality.Authority.PublicKey,
			LocusID: locus, ObservedAt: 1001, Payload: raw,
		})
	}
	boundPayload := BoundPayload{
		LocusID: "FRESH", PurposeSHA256: digestText("purpose"),
		ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	}
	for _, invalid := range []string{"", "DIFFERENT"} {
		if _, err := draft(opBound, invalid, boundPayload); err == nil {
			t.Fatal("BOUND draft accepted empty or payload-mismatched locus")
		}
	}
	unsigned, err := draft(opBound, "FRESH", boundPayload)
	if err != nil {
		t.Fatalf("BOUND must accept a genuinely new ID: %v", err)
	}
	transition, err := signTransition(*unsigned, keys.host)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := transition.Bytes()
	if _, err := vm.issueTransition(raw); err != nil {
		t.Fatal(err)
	}
	// Drafts must see pending state as well as accepted state.
	capacity := DeclareCapacityPayload{
		ActualUnits: 90, ResourceCommitmentSHA256: digestText("resource"), BasisSHA256: digestText("basis"),
	}
	if _, err := draft(opDeclareCapacity, "", capacity); err != nil {
		t.Fatalf("empty body-local draft rejected with pending active locus: %v", err)
	}
	if _, err := draft(opDeclareCapacity, "FRESH", capacity); err == nil {
		t.Fatal("nonempty body-local draft accepted")
	}
	block, err := vm.BuildBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := block.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := block.Accept(context.Background()); err != nil {
		t.Fatal(err)
	}
	state := acceptedState(t, vm)
	capacityTransition := makeTransition(t, state, genesis.Locality.ID, keys.host, opDeclareCapacity, "", capacity)
	next, _, err := applyToClone(state, genesis, capacityTransition)
	if err != nil || next.Capacity.DeclaredActualUnits != 90 {
		t.Fatalf("body-local capacity not accepted in active locus: %v", err)
	}
	capacityTransition.Unsigned.LocusID = "FRESH"
	if _, _, err := applyToClone(state, genesis, capacityTransition); err == nil {
		t.Fatal("application accepted nonempty body-local scope")
	}
	if err := capacityTransition.ValidateSyntax(); err == nil {
		t.Fatal("transition validation accepted nonempty body-local scope")
	}
	for name, payload := range map[string]any{
		opBound:            boundPayload,
		opExhaustFormation: ExhaustFormationAuthorityPayload{BasisSHA256: digestText("exhaust")},
		opActivateSuccessor: ActivateSuccessorPayload{
			ProposalID: "CANDIDATE", ProposalCommitment: digestText("proposal"),
		},
	} {
		locus := ""
		if name == opBound {
			locus = "FRESH"
		}
		if _, err := draft(name, locus, payload); err == nil {
			t.Fatalf("%s draft accepted with active locus", name)
		}
	}
	closePayload := ClosePayload{ClosureBasisSHA256: digestText("close")}
	if _, err := draft(opClose, "FOREIGN", closePayload); err == nil {
		t.Fatal("draft accepted foreign active-locus ID")
	}
	wrongClose := makeTransition(t, state, genesis.Locality.ID, keys.host, opClose, "FOREIGN", closePayload)
	invalidBlock, err := newWireBlock(vm, &WireBlock{
		Schema: blockSchema, ParentID: block.ID().String(), Height: block.Height() + 1,
		Timestamp: wrongClose.Unsigned.ObservedAt, Transitions: []*Transition{wrongClose},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := invalidBlock.Verify(context.Background()); err == nil {
		t.Fatal("consensus verification accepted foreign active-locus ID")
	}
}

func TestCorrectionsFollowExactHistoricalTargetScope(t *testing.T) {
	_, genesis, keys := testGenesis(t)
	state, _ := initialRuntimeState(genesis)
	capacity := makeTransition(t, state, genesis.Locality.ID, keys.host, opDeclareCapacity, "", DeclareCapacityPayload{
		ActualUnits: 100, ResourceCommitmentSHA256: digestText("resource"), BasisSHA256: digestText("basis"),
	})
	state, _, bodyID := mustApply(t, state, genesis, capacity)
	bound := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "HISTORICAL", BoundPayload{
		LocusID: "HISTORICAL", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	state, _, oldID := mustApply(t, state, genesis, bound)
	close := makeTransition(t, state, genesis.Locality.ID, keys.host, opClose, "HISTORICAL", ClosePayload{
		ClosureBasisSHA256: digestText("close"),
	})
	state, _, _ = mustApply(t, state, genesis, close)
	reused := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "HISTORICAL", BoundPayload{
		LocusID: "HISTORICAL", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	if _, _, err := applyToClone(state, genesis, reused); err == nil {
		t.Fatal("BOUND reused a closed locus ID")
	}
	newBound := makeTransition(t, state, genesis.Locality.ID, keys.host, opBound, "CURRENT", BoundPayload{
		LocusID: "CURRENT", PurposeSHA256: digestText("new purpose"), ClosureConditionSHA256: digestText("new closure"), CapacityCeilingUnits: 100,
	})
	state, _, _ = mustApply(t, state, genesis, newBound)
	for _, target := range []struct{ id, locus string }{{oldID, "HISTORICAL"}, {bodyID, ""}} {
		payload := CorrectPayload{
			TargetTransitionID: target.id, ReplacementCommitment: digestText("replacement"), ReasonSHA256: digestText("correction"),
		}
		correct := makeTransition(t, state, genesis.Locality.ID, keys.host, opCorrect, target.locus, payload)
		next, _, err := applyToClone(state, genesis, correct)
		if err != nil {
			t.Fatalf("exact historical target rejected: %v", err)
		}
		if next.ActiveLocusID != "CURRENT" || next.Loci["HISTORICAL"].Phase != "CLOSED" {
			t.Fatal("correction reopened or changed a locus")
		}
		for _, wrong := range []string{"CURRENT", "FOREIGN"} {
			correct.Unsigned.LocusID = wrong
			if _, _, err := applyToClone(state, genesis, correct); err == nil {
				t.Fatal("correction accepted active/foreign locus instead of target locus")
			}
		}
	}
	missing := makeTransition(t, state, genesis.Locality.ID, keys.host, opCorrect, "", CorrectPayload{
		TargetTransitionID: digestText("missing"), ReplacementCommitment: digestText("replacement"), ReasonSHA256: digestText("reason"),
	})
	if _, _, err := applyToClone(state, genesis, missing); err == nil {
		t.Fatal("unknown correction target accepted")
	}
}

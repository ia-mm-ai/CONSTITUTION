package main

import (
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/database/memdb"
)

func TestCorrectionsRejectUnclassifiedTargetEvents(t *testing.T) {
	targetID := digestText("target")
	payload, _ := json.Marshal(CorrectPayload{
		TargetTransitionID: targetID, ReplacementCommitment: digestText("replacement"), ReasonSHA256: digestText("reason"),
	})
	for name, target := range map[string]EventRecord{
		"unknown operation": {Operation: "UNKNOWN", Effect: operationEffects[opDeclareCapacity]},
		"missing operation": {Effect: operationEffects[opDeclareCapacity]},
		"missing effect":    {Operation: opDeclareCapacity},
		"wrong effect":      {Operation: opDeclareCapacity, Effect: operationEffects[opBound]},
		"body with locus": {
			Operation: opDeclareCapacity, Effect: operationEffects[opDeclareCapacity], LocusID: "FOREIGN",
		},
		"active without locus": {Operation: opClose, Effect: operationEffects[opClose]},
	} {
		t.Run(name, func(t *testing.T) {
			vm, _, genesis, _ := initializeTestVM(t, memdb.New())
			state := vm.genesisBlock.postState
			target.TransitionID = targetID
			target.ActorID = genesis.Locality.ID
			state.Events[targetID] = target
			u := UnsignedTransition{
				Operation: opCorrect, Effect: operationEffects[opCorrect], LocusID: target.LocusID, Payload: payload,
			}
			if err := state.validateOperationScope(&u); err == nil {
				t.Fatal("unclassified target accepted by scope guard")
			}
			if err := state.applyOperation(genesis, &Transition{Unsigned: u}, digestText("correction")); err == nil {
				t.Fatal("unclassified target accepted during application")
			}
			if _, err := vm.draftTransition(&TransitionDraftRequest{
				Operation: opCorrect, ActorID: genesis.Locality.ID, ActorPublicKey: genesis.Locality.Authority.PublicKey,
				LocusID: target.LocusID, ObservedAt: 1, Payload: payload,
			}); err == nil {
				t.Fatal("unclassified target accepted by draft construction")
			}
		})
	}
}

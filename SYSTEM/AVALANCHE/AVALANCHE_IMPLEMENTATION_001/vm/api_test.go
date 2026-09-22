package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ava-labs/avalanchego/database/memdb"
)

func TestHTTPDraftSubmitAndReceiptLifecycle(t *testing.T) {
	vm, _, genesis, keys := initializeTestVM(t, memdb.New())
	handler := vm.apiHandler()

	for _, path := range []string{"/status", "/genesis", "/state", "/capacity", "/currentness"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s = %d: %s", path, response.Code, response.Body.String())
		}
		if path == "/status" {
			var status map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &status); err != nil {
				t.Fatal(err)
			}
			if status["protocol_contract_sha256"] != digestText(string(protocolContractJSON)) {
				t.Fatal("status does not expose the exact embedded operation contract digest")
			}
		}
	}
	currentnessResponse := httptest.NewRecorder()
	currentnessPath := "/currentness?observed_at=1&carrier_set_sha256=" + digestText("api carriers")
	handler.ServeHTTP(currentnessResponse, httptest.NewRequest(http.MethodGet, currentnessPath, nil))
	if currentnessResponse.Code != http.StatusOK {
		t.Fatalf("currentness material = %d: %s", currentnessResponse.Code, currentnessResponse.Body.String())
	}
	var currentness struct {
		NextPulse struct {
			Commitment string `json:"currentness_commitment_sha256"`
		} `json:"next_pulse"`
	}
	if err := json.Unmarshal(currentnessResponse.Body.Bytes(), &currentness); err != nil {
		t.Fatal(err)
	}
	if len(currentness.NextPulse.Commitment) != 64 {
		t.Fatal("currentness endpoint omitted the exact next-pulse commitment")
	}

	payload, _ := json.Marshal(BoundPayload{
		LocusID: "LOCUS-API-001", PurposeSHA256: digestText("api purpose"), ClosureConditionSHA256: digestText("api closure"), CapacityCeilingUnits: 100,
	})
	draftRequest, _ := json.Marshal(&TransitionDraftRequest{
		Operation: opBound, ActorID: genesis.Locality.ID, ActorPublicKey: genesis.Locality.Authority.PublicKey,
		LocusID: "LOCUS-API-001", ObservedAt: 1, Payload: payload,
	})
	draftResponse := httptest.NewRecorder()
	handler.ServeHTTP(draftResponse, httptest.NewRequest(http.MethodPost, "/drafts", bytes.NewReader(draftRequest)))
	if draftResponse.Code != http.StatusOK {
		t.Fatalf("draft = %d: %s", draftResponse.Code, draftResponse.Body.String())
	}
	var draft struct {
		Unsigned UnsignedTransition `json:"unsigned"`
	}
	if err := json.Unmarshal(draftResponse.Body.Bytes(), &draft); err != nil {
		t.Fatal(err)
	}
	transition, err := signTransition(draft.Unsigned, keys.host)
	if err != nil {
		t.Fatal(err)
	}
	transitionBytes, _ := transition.Bytes()
	submitResponse := httptest.NewRecorder()
	handler.ServeHTTP(submitResponse, httptest.NewRequest(http.MethodPost, "/transitions", bytes.NewReader(transitionBytes)))
	if submitResponse.Code != http.StatusAccepted {
		t.Fatalf("submit = %d: %s", submitResponse.Code, submitResponse.Body.String())
	}
	var submitted struct {
		TransitionID string `json:"transition_id"`
	}
	if err := json.Unmarshal(submitResponse.Body.Bytes(), &submitted); err != nil {
		t.Fatal(err)
	}

	pendingReceipt := httptest.NewRecorder()
	handler.ServeHTTP(pendingReceipt, httptest.NewRequest(http.MethodGet, "/receipts/"+submitted.TransitionID, nil))
	if pendingReceipt.Code != http.StatusNotFound {
		t.Fatalf("pending receipt = %d", pendingReceipt.Code)
	}

	built, err := vm.BuildBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := built.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := built.Accept(context.Background()); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"/receipts/" + submitted.TransitionID,
		"/transitions/" + submitted.TransitionID,
		"/blocks/" + built.ID().String(),
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s = %d: %s", path, response.Code, response.Body.String())
		}
	}
}

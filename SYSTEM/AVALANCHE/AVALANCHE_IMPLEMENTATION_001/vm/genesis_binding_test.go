package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestCurrentProfileAndGenesisTemplateBinding(t *testing.T) {
	binding, err := os.ReadFile("../bindings/SOURCE_STATE_BINDING_001.json")
	if err != nil {
		t.Fatal(err)
	}
	bindingDigest := sha256.Sum256(binding)
	if hex.EncodeToString(bindingDigest[:]) != constitutionBindingSHA256 || uint64(len(binding)) != constitutionBindingBytes {
		t.Fatal("runtime Source/State binding differs from the compiled exact-byte identity")
	}
	profile, err := os.ReadFile("profile/PRESENCE_AVALANCHE_FORM_001.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(profile)
	if hex.EncodeToString(digest[:]) != formSHA256 {
		t.Fatal("runtime profile digest does not bind the exact shipped bytes")
	}
	raw, err := os.ReadFile("genesis/PRESENCE_AVALANCHE_RUNTIME_GENESIS_TEMPLATE_001.json")
	if err != nil {
		t.Fatal(err)
	}
	var template Genesis
	if err := decodeStrict(raw, &template); err != nil {
		t.Fatal(err)
	}
	if template.Locality.ID != "" || template.Locality.Authority.PublicKey != "" || template.InitialState.Revision != 0 {
		t.Fatal("genesis template falsely claims a formed or historical deployment")
	}
	_, fixture, _ := testGenesis(t)
	template.Domain.GenesisID = fixture.Domain.GenesisID
	template.Domain.Purpose = fixture.Domain.Purpose
	template.Locality = fixture.Locality
	if err := template.Validate(); err != nil {
		t.Fatalf("current genesis template cannot be materialized: %v", err)
	}
	if template.Constitution.PublicOrigin.Commit != "5e09d1fbe3c1008937d95497fcc162a6ebd4190d" {
		t.Fatal("Source origin is not pinned to current main")
	}
	template.Constitution.FormBindingBytes++
	if err := template.Validate(); err == nil {
		t.Fatal("incorrect successor-local binding length accepted")
	}
	template.Constitution.FormBindingBytes--
	template.Constitution.PublicOrigin.Binding = "EXACT_COMMIT_PATH_LENGTH_SHA256"
	if err := template.Validate(); err == nil {
		t.Fatal("ambiguous Source/binding origin description accepted")
	}
	template.Constitution.PublicOrigin.Binding = "SOURCE_AT_STARTING_COMMIT_WITH_SUCCESSOR_LOCAL_STATE_BINDING"
	template.Constitution.MachineSHA256 = digestText("stale predecessor source")
	if err := template.Validate(); err == nil {
		t.Fatal("stale Source digest accepted")
	}
}

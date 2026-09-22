package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthorityGenerationAndGenesisMaterialization(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "authority")
	keyID, privatePath, publicPath, err := generateAuthority(directory)
	if err != nil {
		t.Fatalf("generate authority: %v", err)
	}
	if keyID == "" {
		t.Fatal("empty key ID")
	}
	privateInfo, err := os.Stat(privatePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := privateInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("private authority permissions = %#o", got)
	}
	if _, _, _, err := generateAuthority(directory); err == nil {
		t.Fatal("authority generation overwrote existing custody files")
	}

	_, genesis, _ := testGenesis(t)
	genesis.Domain.GenesisID = ""
	genesis.Domain.Purpose = ""
	genesis.Locality.ID = ""
	genesis.Locality.SourceReference = ""
	genesis.Locality.SourceSHA256 = ""
	genesis.Locality.Authority.KeyID = "MATERIALIZE-WITH-PRESENCE-AVALANCHE-VM"
	genesis.Locality.Authority.PublicKey = ""
	templateBytes, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	templatePath := filepath.Join(t.TempDir(), "template.json")
	if err := os.WriteFile(templatePath, templateBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	descriptor := DeploymentDescriptor{
		Schema:          deploymentDescriptorSchema,
		GenesisID:       "LOCALITY-RUNTIME-TEST-001",
		LocalityID:      "LOCALITY-RUNTIME-TEST",
		Purpose:         "Exercise the LOCALITY executable grammar.",
		SourceReference: "CONTINUITY",
		SourceSHA256:    digestText("test continuity source"),
	}
	descriptorBytes, err := json.MarshalIndent(&descriptor, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	descriptorPath := filepath.Join(t.TempDir(), "deployment.json")
	if err := os.WriteFile(descriptorPath, descriptorBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(t.TempDir(), "genesis.json")
	if err := materializeGenesis(templatePath, descriptorPath, publicPath, outputPath); err != nil {
		t.Fatalf("materialize genesis: %v", err)
	}
	materializedBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	materialized, err := parseGenesis(materializedBytes)
	if err != nil {
		t.Fatalf("materialized genesis is invalid: %v", err)
	}
	if materialized.Locality.Authority.KeyID != keyID {
		t.Fatalf("materialized key ID = %s", materialized.Locality.Authority.KeyID)
	}
	if materialized.Locality.Authority.PublicKey == "" {
		t.Fatal("materialized genesis has no public authority")
	}
	if materialized.Domain.GenesisID != descriptor.GenesisID || materialized.Locality.ID != descriptor.LocalityID {
		t.Fatal("materialized genesis does not bind the deployment descriptor")
	}
	state, err := initialRuntimeState(materialized)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(BoundPayload{
		LocusID: "LOCUS-001", PurposeSHA256: digestText("purpose"), ClosureConditionSHA256: digestText("closure"), CapacityCeilingUnits: 100,
	})
	unsigned := UnsignedTransition{
		Schema: transitionSchema, Operation: opBound, Revision: 1,
		PreviousStateCommitment: state.StateCommitment, ActorID: materialized.Locality.ID,
		ActorPublicKey: materialized.Locality.Authority.PublicKey, Nonce: 0, LocusID: "LOCUS-001",
		ObservedAt: 1, Effect: operationEffects[opBound], Payload: payload,
	}
	unsignedBytes, _ := json.MarshalIndent(&unsigned, "", "  ")
	unsignedPath := filepath.Join(t.TempDir(), "unsigned.json")
	if err := os.WriteFile(unsignedPath, unsignedBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	signedPath := filepath.Join(t.TempDir(), "transition.json")
	if err := signUnsignedFile(unsignedPath, privatePath, signedPath); err != nil {
		t.Fatalf("sign unsigned transition: %v", err)
	}
	signedBytes, err := os.ReadFile(signedPath)
	if err != nil {
		t.Fatal(err)
	}
	transition, _, err := parseTransition(signedBytes)
	if err != nil {
		t.Fatalf("signed transition is invalid: %v", err)
	}
	if _, _, err := applyToClone(state, materialized, transition); err != nil {
		t.Fatalf("signed host transition cannot execute: %v", err)
	}
	if err := signUnsignedFile(unsignedPath, privatePath, signedPath); err == nil {
		t.Fatal("transition signer overwrote existing output")
	}
}

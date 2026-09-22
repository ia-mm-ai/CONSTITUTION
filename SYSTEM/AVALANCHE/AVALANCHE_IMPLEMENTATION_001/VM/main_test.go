package main

import (
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/version"
)

func TestMachineVersionContract(t *testing.T) {
	wantVMID := "25tZjky6SecZA1Gwc6VwD2ouy64xAUdgNLo1C8dkXfbsFNxaTk"
	if got := presenceVMID().String(); got != wantVMID {
		t.Fatalf("unexpected PRESENCE AVALANCHE VM ID: got %s want %s", got, wantVMID)
	}
	if got := presenceVMID().String(); got == "pJHx1NU8ghWsiwg1vrqwaqE5uKh1k4EJRBkpB1tKV1QuQFhMi" {
		t.Fatal("successor VM ID must not reuse the formation VM ID")
	}

	document := machineVersion{
		Name:               "presence-avalanche-vm",
		Version:            vmVersion,
		AvalancheGo:        "v1.15.0",
		AvalancheGoProfile: avalancheGoProfile,
		RPCChainVM:         version.RPCChainVMProtocol,
		VMID:               presenceVMID().String(),
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	decoded := new(machineVersion)
	if err := json.Unmarshal(encoded, decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.RPCChainVM != 46 {
		t.Fatalf("unexpected RPCChainVM protocol: %d", decoded.RPCChainVM)
	}
	if decoded.AvalancheGoProfile != "v1.15.0+LOCALITY_SECURITY_OVERLAY_001" {
		t.Fatalf("unexpected AvalancheGo profile: %s", decoded.AvalancheGoProfile)
	}
}

package main

import (
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/version"
)

func TestMachineVersionContract(t *testing.T) {
	wantVMID := "cQYXygFUVutQucm4s8pr8M51sRdS1UfrTbpMdEgpRYt2JvEzr"
	if got := presenceAvalancheVMID().String(); got != wantVMID {
		t.Fatalf("unexpected PRESENCE Avalanche VM ID: got %s want %s", got, wantVMID)
	}
	if got := presenceAvalancheVMID().String(); got == "25tZjky6SecZA1Gwc6VwD2ouy64xAUdgNLo1C8dkXfbsFNxaTk" {
		t.Fatal("successor VM ID must not reuse the direct predecessor VM ID")
	}

	document := machineVersion{
		Name:               "presence-avalanche-vm",
		Version:            vmVersion,
		AvalancheGo:        "v1.15.0",
		AvalancheGoProfile: avalancheGoProfile,
		RPCChainVM:         version.RPCChainVMProtocol,
		VMID:               presenceAvalancheVMID().String(),
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

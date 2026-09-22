package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/version"
)

func TestMachineVersionContract(t *testing.T) {
	got := localityVMID()
	if got.String() != "cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8" {
		t.Fatalf("unexpected canonical VM ID: %s", got)
	}
	if !bytes.Equal(got[:len(vmIDDomain)], []byte(vmIDDomain)) ||
		!bytes.Equal(got[len(vmIDDomain):], make([]byte, len(got)-len(vmIDDomain))) {
		t.Fatal("VM ID must be the exact protocol name zero-padded to an Avalanche ID")
	}
	if got.String() == directPredecessorVMID || got.String() == predecessorVMID {
		t.Fatal("implementation VM ID must not reuse a historical VM ID")
	}

	document := machineVersion{
		Name:               "presence-avalanche-vm",
		Version:            implementationVersion,
		Implementation:     implementationID,
		Protocol:           vmIDDomain,
		ProtocolContractSHA256: protocolContractSHA256,
		AvalancheGo:        "v1.15.0",
		AvalancheGoProfile: avalancheGoProfile,
		RPCChainVM:         version.RPCChainVMProtocol,
		VMID:               localityVMID().String(),
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
	if decoded.AvalancheGoProfile != "v1.15.0" {
		t.Fatalf("unexpected AvalancheGo profile: %s", decoded.AvalancheGoProfile)
	}
	if decoded.ProtocolContractSHA256 != digestText(string(protocolContractJSON)) {
		t.Fatal("machine version does not expose the exact embedded operation contract digest")
	}
}

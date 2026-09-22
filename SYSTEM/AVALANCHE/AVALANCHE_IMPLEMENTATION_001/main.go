package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/version"
	"github.com/ava-labs/avalanchego/vms/rpcchainvm"
)

const (
	vmVersion          = "presence-avalanche-vm/1.0.0"
	vmIDDomain         = "PRESENCE_AVALANCHE_VM_001"
	avalancheGoProfile = "v1.15.0+PRESENCE_AVALANCHE_SECURITY_OVERLAY_001"
)

type machineVersion struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	AvalancheGo        string `json:"avalanchego"`
	AvalancheGoProfile string `json:"avalanchego_profile"`
	RPCChainVM         uint   `json:"rpcchainvm"`
	VMID               string `json:"vm_id"`
}

func presenceAvalancheVMID() ids.ID {
	return ids.ID(hashing.ComputeHash256Array([]byte(vmIDDomain)))
}

func main() {
	switch {
	case len(os.Args) == 2 && os.Args[1] == "--version":
		fmt.Printf("%s avalanchego-profile=%s rpcchainvm-protocol=%d vm-id=%s\n", vmVersion, avalancheGoProfile, version.RPCChainVMProtocol, presenceAvalancheVMID())
		return
	case len(os.Args) == 2 && os.Args[1] == "--version-json":
		if err := json.NewEncoder(os.Stdout).Encode(machineVersion{
			Name:               "presence-avalanche-vm",
			Version:            vmVersion,
			AvalancheGo:        "v1.15.0",
			AvalancheGoProfile: avalancheGoProfile,
			RPCChainVM:         version.RPCChainVMProtocol,
			VMID:               presenceAvalancheVMID().String(),
		}); err != nil {
			exitWithError("encode machine-readable version", err)
		}
		return
	case len(os.Args) == 2 && os.Args[1] == "--vm-id":
		fmt.Println(presenceAvalancheVMID())
		return
	case len(os.Args) == 3 && os.Args[1] == "--check-genesis":
		genesisBytes, err := os.ReadFile(os.Args[2])
		if err != nil {
			exitWithError("read genesis", err)
		}
		genesis, err := parseGenesis(genesisBytes)
		if err != nil {
			exitWithError("validate genesis", err)
		}
		digest := sha256.Sum256(genesisBytes)
		fmt.Printf("valid PRESENCE Avalanche genesis: id=%s locality=%s sha256=%s\n", genesis.Domain.GenesisID, genesis.Locality.ID, hex.EncodeToString(digest[:]))
		return
	case len(os.Args) == 3 && os.Args[1] == "--check-transition":
		transitionBytes, err := os.ReadFile(os.Args[2])
		if err != nil {
			exitWithError("read transition", err)
		}
		transition, transitionID, err := parseTransition(transitionBytes)
		if err != nil {
			exitWithError("validate transition", err)
		}
		fmt.Printf("valid PRESENCE Avalanche transition: operation=%s revision=%d id=%s\n", transition.Unsigned.Operation, transition.Unsigned.Revision, hex.EncodeToString(transitionID[:]))
		return
	case len(os.Args) == 3 && os.Args[1] == "--generate-authority":
		keyID, privatePath, publicPath, err := generateAuthority(os.Args[2])
		if err != nil {
			exitWithError("generate PRESENCE Avalanche authority", err)
		}
		fmt.Printf("generated PRESENCE Avalanche authority %s\nprivate: %s\npublic: %s\n", keyID, privatePath, publicPath)
		return
	case len(os.Args) == 6 && os.Args[1] == "--materialize-genesis":
		if err := materializeGenesis(os.Args[2], os.Args[3], os.Args[4], os.Args[5]); err != nil {
			exitWithError("materialize genesis", err)
		}
		fmt.Printf("materialized PRESENCE Avalanche genesis: %s\n", os.Args[5])
		return
	case len(os.Args) == 5 && os.Args[1] == "--sign-unsigned":
		if err := signUnsignedFile(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			exitWithError("sign unsigned transition", err)
		}
		fmt.Printf("signed canonical PRESENCE Avalanche transition: %s\n", os.Args[4])
		return
	case len(os.Args) != 1:
		fmt.Fprintln(os.Stderr, "usage: presence-avalanche-vm [--version | --version-json | --vm-id | --check-genesis PATH | --check-transition PATH | --generate-authority DIRECTORY | --materialize-genesis TEMPLATE DEPLOYMENT_DESCRIPTOR PUBLIC_AUTHORITY OUTPUT | --sign-unsigned UNSIGNED PRIVATE_AUTHORITY OUTPUT]")
		os.Exit(2)
	}

	if err := rpcchainvm.Serve(context.Background(), &VM{}); err != nil {
		exitWithError("serve RPCChainVM", err)
	}
}

func exitWithError(action string, err error) {
	fmt.Fprintf(os.Stderr, "PRESENCE Avalanche VM: %s: %v\n", action, err)
	os.Exit(1)
}

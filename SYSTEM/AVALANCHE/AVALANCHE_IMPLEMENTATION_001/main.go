package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/version"
	"github.com/ava-labs/avalanchego/vms/rpcchainvm"
)

const (
	vmVersion          = "presence-avalanche-vm/1.0.0"
	vmIDDomain         = "PRESENCE_AVALANCHE_VM_001"
	avalancheGoProfile = "v1.15.0+PRESENCE_SECURITY_OVERLAY_001"
)

type machineVersion struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	AvalancheGo        string `json:"avalanchego"`
	AvalancheGoProfile string `json:"avalanchego_profile"`
	RPCChainVM         uint   `json:"rpcchainvm"`
	VMID               string `json:"vm_id"`
	VMIDDerivation     string `json:"vm_id_derivation"`
	VMIDInput          string `json:"vm_id_input"`
}

// presenceVMID derives the VM ID through the official pinned AvalancheGo API:
// the exact canonical identity PRESENCE_AVALANCHE_VM_001 is zero-extended to
// the 32-byte ids.ID width and converted with ids.ToID. This is the standard
// AvalancheGo VM-name derivation. No transport-safe substitute name was
// required: ids.ToID accepts the canonical identity, including underscores,
// so the derivation input equals the canonical implementation VM identity.
func presenceVMID() ids.ID {
	padded := make([]byte, ids.IDLen)
	copy(padded, vmIDDomain)
	id, err := ids.ToID(padded)
	if err != nil {
		panic(fmt.Errorf("derive PRESENCE AVALANCHE VM ID: %w", err))
	}
	return id
}

func main() {
	switch {
	case len(os.Args) == 2 && os.Args[1] == "--version":
		fmt.Printf("%s avalanchego-profile=%s rpcchainvm-protocol=%d vm-id=%s\n", vmVersion, avalancheGoProfile, version.RPCChainVMProtocol, presenceVMID())
		return
	case len(os.Args) == 2 && os.Args[1] == "--version-json":
		if err := json.NewEncoder(os.Stdout).Encode(machineVersion{
			Name:               "presence-avalanche-vm",
			Version:            vmVersion,
			AvalancheGo:        "v1.15.0",
			AvalancheGoProfile: avalancheGoProfile,
			RPCChainVM:         version.RPCChainVMProtocol,
			VMID:               presenceVMID().String(),
			VMIDDerivation:     "avalanchego/ids.ToID over the canonical identity zero-extended to 32 bytes",
			VMIDInput:          vmIDDomain,
		}); err != nil {
			exitWithError("encode machine-readable version", err)
		}
		return
	case len(os.Args) == 2 && os.Args[1] == "--vm-id":
		fmt.Println(presenceVMID())
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
		fmt.Printf("valid PRESENCE runtime genesis: id=%s locality=%s sha256=%s\n", genesis.Domain.GenesisID, genesis.Locality.ID, hex.EncodeToString(digest[:]))
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
		fmt.Printf("valid PRESENCE transition: operation=%s revision=%d id=%s\n", transition.Unsigned.Operation, transition.Unsigned.Revision, hex.EncodeToString(transitionID[:]))
		return
	case len(os.Args) == 3 && os.Args[1] == "--generate-authority":
		keyID, privatePath, publicPath, err := generateAuthority(os.Args[2])
		if err != nil {
			exitWithError("generate locality authority", err)
		}
		fmt.Printf("generated locality authority %s\nprivate: %s\npublic: %s\n", keyID, privatePath, publicPath)
		return
	case len(os.Args) == 6 && os.Args[1] == "--materialize-genesis":
		if err := materializeGenesis(os.Args[2], os.Args[3], os.Args[4], os.Args[5]); err != nil {
			exitWithError("materialize genesis", err)
		}
		fmt.Printf("materialized PRESENCE genesis: %s\n", os.Args[5])
		return
	case len(os.Args) == 5 && os.Args[1] == "--sign-unsigned":
		if err := signUnsignedFile(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			exitWithError("sign unsigned transition", err)
		}
		fmt.Printf("signed canonical PRESENCE transition: %s\n", os.Args[4])
		return
	case len(os.Args) == 4 && os.Args[1] == "--serve-rehearsal":
		if err := serveRehearsal(os.Args[2], os.Args[3]); err != nil {
			exitWithError("serve disposable rehearsal VM", err)
		}
		return
	case len(os.Args) != 1:
		fmt.Fprintln(os.Stderr, "usage: presence-avalanche-vm [--version | --version-json | --vm-id | --check-genesis PATH | --check-transition PATH | --generate-authority DIRECTORY | --materialize-genesis TEMPLATE DEPLOYMENT_DESCRIPTOR PUBLIC_AUTHORITY OUTPUT | --sign-unsigned UNSIGNED PRIVATE_AUTHORITY OUTPUT | --serve-rehearsal GENESIS LOOPBACK_ADDRESS]")
		os.Exit(2)
	}

	if err := rpcchainvm.Serve(context.Background(), &VM{}); err != nil {
		exitWithError("serve RPCChainVM", err)
	}
}

func exitWithError(action string, err error) {
	fmt.Fprintf(os.Stderr, "presence-avalanche-vm: %s: %v\n", action, err)
	os.Exit(1)
}

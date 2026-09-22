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
	vmName             = "PRESENCE_AVALANCHE_VM_001"
	avalancheGoVersion = "v1.15.0"
	avalancheGoCommit  = "70bd6d063b7343fd2cd8217200aaf77b57f19f68"
	avalancheGoProfile = "v1.15.0+PRESENCE_AVALANCHE_SECURITY_OVERLAY_001"

	// expectedVMID is the identifier produced by the official AvalancheGo
	// mechanism for this exact VM name. It is asserted at startup so a drifted
	// build can never be installed under this identity.
	expectedVMID = "cNhhBznc1YN29QVJMQbK7sxy6GsimFKN6yvftRjPK7qWEvZw8"
)

type machineVersion struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	Implementation     string `json:"implementation"`
	AvalancheGo        string `json:"avalanchego"`
	AvalancheGoCommit  string `json:"avalanchego_commit"`
	AvalancheGoProfile string `json:"avalanchego_profile"`
	RPCChainVM         uint   `json:"rpcchainvm"`
	VMName             string `json:"vm_name"`
	VMID               string `json:"vm_id"`
	CoreDigest         string `json:"core_digest"`
	CoreCommit         string `json:"core_commit"`
}

// presenceVMID derives the VM identifier exactly as the official Avalanche
// tooling does: the VM name is zero-padded into a 32-byte identifier and
// CB58-encoded. It is not a digest of the name.
func presenceVMID() (ids.ID, error) {
	if len(vmName) > ids.IDLen {
		return ids.Empty, fmt.Errorf("vm name exceeds %d bytes", ids.IDLen)
	}
	padded := make([]byte, ids.IDLen)
	copy(padded, vmName)
	id, err := ids.ToID(padded)
	if err != nil {
		return ids.Empty, err
	}
	if id.String() != expectedVMID {
		return ids.Empty, fmt.Errorf("derived vm id %s differs from the recorded identity %s", id, expectedVMID)
	}
	return id, nil
}

func rpcChainVMProtocol() uint {
	return version.RPCChainVMProtocol
}

func mustPresenceVMID() ids.ID {
	id, err := presenceVMID()
	if err != nil {
		exitWithError("derive vm id", err)
	}
	return id
}

func main() {
	switch {
	case len(os.Args) == 2 && os.Args[1] == "--version":
		fmt.Printf("%s avalanchego-profile=%s rpcchainvm-protocol=%d vm-id=%s\n", vmVersion, avalancheGoProfile, version.RPCChainVMProtocol, mustPresenceVMID())
		return
	case len(os.Args) == 2 && os.Args[1] == "--version-json":
		if err := json.NewEncoder(os.Stdout).Encode(machineVersion{
			Name:               "presence-avalanche-vm",
			Version:            vmVersion,
			Implementation:     implementationID + "/" + implementationLevel,
			AvalancheGo:        avalancheGoVersion,
			AvalancheGoCommit:  avalancheGoCommit,
			AvalancheGoProfile: avalancheGoProfile,
			RPCChainVM:         version.RPCChainVMProtocol,
			VMName:             vmName,
			VMID:               mustPresenceVMID().String(),
			CoreDigest:         coreDigest,
			CoreCommit:         coreCommit,
		}); err != nil {
			exitWithError("encode machine-readable version", err)
		}
		return
	case len(os.Args) == 2 && os.Args[1] == "--vm-id":
		fmt.Println(mustPresenceVMID())
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
		fmt.Printf("valid PRESENCE AVALANCHE runtime genesis: id=%s locality=%s sha256=%s\n", genesis.Domain.GenesisID, genesis.Locality.ID, hex.EncodeToString(digest[:]))
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
		fmt.Printf("valid PRESENCE AVALANCHE transition: operation=%s revision=%d id=%s\n", transition.Unsigned.Operation, transition.Unsigned.Revision, hex.EncodeToString(transitionID[:]))
		return
	case len(os.Args) == 2 && os.Args[1] == "--core-binding":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(currentCore()); err != nil {
			exitWithError("encode core binding", err)
		}
		return
	case len(os.Args) == 3 && os.Args[1] == "--check-core-binding":
		summary, err := checkCoreBindingFile(os.Args[2])
		if err != nil {
			exitWithError("check core binding", err)
		}
		fmt.Println(summary)
		return
	case len(os.Args) == 3 && os.Args[1] == "--generate-test-authority":
		keyID, privatePath, publicPath, err := generateTestAuthority(os.Args[2])
		if err != nil {
			exitWithError("generate test authority", err)
		}
		fmt.Printf("generated TEST-ONLY authority %s\nprivate: %s\npublic: %s\nthis material is for local rehearsal only and is never administrator input or production custody\n", keyID, privatePath, publicPath)
		return
	case len(os.Args) == 5 && os.Args[1] == "--materialize-genesis":
		if err := materializeGenesis(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			exitWithError("materialize genesis", err)
		}
		fmt.Printf("materialized PRESENCE AVALANCHE genesis: %s\n", os.Args[4])
		return
	case len(os.Args) == 5 && os.Args[1] == "--materialize-deployment-plan":
		if err := materializeDeploymentPlan(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			exitWithError("materialize deployment plan", err)
		}
		fmt.Printf("materialized public deployment plan: %s\n", os.Args[4])
		return
	case len(os.Args) == 3 && os.Args[1] == "--check-admin-input":
		raw, err := os.ReadFile(os.Args[2])
		if err != nil {
			exitWithError("read administrator input", err)
		}
		input, err := parseAdminInput(raw)
		if err != nil {
			exitWithError("validate administrator input", err)
		}
		fmt.Printf("valid administrator input: locality=%s genesis=%s network=%s validators=%d\n",
			input.Instantiation.LocalityID, input.Instantiation.GenesisID, input.Instantiation.Network, len(input.Validators))
		return
	case len(os.Args) == 5 && os.Args[1] == "--sign-unsigned":
		if err := signUnsignedFile(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			exitWithError("sign unsigned transition", err)
		}
		fmt.Printf("signed canonical PRESENCE AVALANCHE transition: %s\n", os.Args[4])
		return
	case len(os.Args) != 1:
		fmt.Fprintln(os.Stderr, "usage: presence-avalanche-vm [--version | --version-json | --vm-id | --core-binding | --check-core-binding PATH | --check-genesis PATH | --check-transition PATH | --check-admin-input PATH | --materialize-genesis TEMPLATE ADMIN_INPUT OUTPUT | --materialize-deployment-plan ADMIN_INPUT GENESIS OUTPUT | --generate-test-authority DIRECTORY | --sign-unsigned UNSIGNED PRIVATE_AUTHORITY OUTPUT]")
		os.Exit(2)
	}

	if err := rpcchainvm.Serve(context.Background(), &VM{}); err != nil {
		exitWithError("serve RPCChainVM", err)
	}
}

func exitWithError(action string, err error) {
	fmt.Fprintf(os.Stderr, "presence avalanche vm: %s: %v\n", action, err)
	os.Exit(1)
}

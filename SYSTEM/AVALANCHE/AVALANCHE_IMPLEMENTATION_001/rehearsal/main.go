package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ava-labs/avalanchego/config"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/tests/fixture/tmpnet"
	"github.com/ava-labs/avalanchego/utils/logging"
)

const (
	rehearsalName = "PRESENCE-REHEARSAL-001"
	nodeCount     = 3
	commandWait   = 5 * time.Minute
)

type nodeSummary struct {
	NodeID string `json:"node_id"`
	URI    string `json:"uri"`
}

type networkSummary struct {
	Schema       string        `json:"schema"`
	NetworkDir   string        `json:"network_dir"`
	NetworkID    uint32        `json:"network_id"`
	VMID         string        `json:"vm_id"`
	SubnetID     string        `json:"subnet_id"`
	BlockchainID string        `json:"blockchain_id"`
	Nodes        []nodeSummary `json:"nodes"`
}

func main() {
	if len(os.Args) < 2 {
		fatal(errors.New("usage: locality-rehearsal <start|inspect|restart|bounce-one|stop> [flags]"))
	}

	var err error
	switch os.Args[1] {
	case "start":
		err = start(os.Args[2:])
	case "inspect":
		err = inspect(os.Args[2:])
	case "restart":
		err = restart(os.Args[2:])
	case "bounce-one":
		err = bounceOne(os.Args[2:])
	case "stop":
		err = stop(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fatal(err)
	}
}

func start(args []string) error {
	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	avalancheGoPath := flags.String("avalanchego", "", "absolute path to AvalancheGo v1.15.0")
	pluginDir := flags.String("plugin-dir", "", "directory containing the PRESENCE AVALANCHE VM binary named by VM ID")
	genesisPath := flags.String("genesis", "", "materialized PRESENCE runtime genesis")
	rootDir := flags.String("root", "", "parent directory for disposable rehearsal networks")
	vmIDText := flags.String("vm-id", "", "PRESENCE AVALANCHE VM ID")
	if err := flags.Parse(args); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"avalanchego": *avalancheGoPath,
		"plugin-dir":  *pluginDir,
		"genesis":     *genesisPath,
		"root":        *rootDir,
		"vm-id":       *vmIDText,
	} {
		if value == "" {
			return fmt.Errorf("--%s is required", name)
		}
	}

	avalancheGoBinary, err := checkedFile(*avalancheGoPath)
	if err != nil {
		return fmt.Errorf("avalanchego: %w", err)
	}
	genesisFile, err := checkedFile(*genesisPath)
	if err != nil {
		return fmt.Errorf("genesis: %w", err)
	}
	pluginDirectory, err := filepath.Abs(*pluginDir)
	if err != nil {
		return err
	}
	rootDirectory, err := filepath.Abs(*rootDir)
	if err != nil {
		return err
	}
	vmID, err := ids.FromString(*vmIDText)
	if err != nil {
		return fmt.Errorf("vm-id: %w", err)
	}
	pluginPath := filepath.Join(pluginDirectory, vmID.String())
	if _, err := checkedFile(pluginPath); err != nil {
		return fmt.Errorf("plugin: %w", err)
	}
	genesisBytes, err := os.ReadFile(genesisFile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(rootDirectory, 0o700); err != nil {
		return err
	}

	nodes := tmpnet.NewNodesOrPanic(nodeCount)
	validatorIDs := make([]ids.NodeID, 0, len(nodes))
	for _, node := range nodes {
		validatorIDs = append(validatorIDs, node.NodeID)
	}
	upgradeFlags, err := tmpnet.UpgradeFlags(tmpnet.UpgradeConfig(0))
	if err != nil {
		return err
	}
	defaultFlags := tmpnet.DefaultTmpnetFlags()
	// Keep the disposable network below constrained CI/container hard limits.
	// Production operators should size this independently for their host.
	defaultFlags[config.FdLimitKey] = "8192"
	for key, value := range upgradeFlags {
		defaultFlags[key] = value
	}

	network := &tmpnet.Network{
		Owner:        "locality-rehearsal-003",
		Nodes:        nodes,
		DefaultFlags: defaultFlags,
		DefaultRuntimeConfig: tmpnet.NodeRuntimeConfig{
			Process: &tmpnet.ProcessRuntimeConfig{
				AvalancheGoPath:   avalancheGoBinary,
				PluginDir:         pluginDirectory,
				ReuseDynamicPorts: true,
			},
		},
		Subnets: []*tmpnet.Subnet{
			{
				Name:         rehearsalName,
				ValidatorIDs: validatorIDs,
				Chains: []*tmpnet.Chain{
					{
						VMID:        vmID,
						Genesis:     genesisBytes,
						VersionArgs: []string{"--version-json"},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandWait)
	defer cancel()
	if err := tmpnet.BootstrapNewNetwork(ctx, logging.NoLog{}, network, rootDirectory); err != nil {
		_ = network.Stop(context.Background())
		return fmt.Errorf("bootstrap three-node LOCALITY rehearsal: %w", err)
	}
	return emitSummary(network)
}

func inspect(args []string) error {
	networkDir, err := networkDirFlag("inspect", args)
	if err != nil {
		return err
	}
	network, err := readNetwork(networkDir)
	if err != nil {
		return err
	}
	return emitSummary(network)
}

func restart(args []string) error {
	networkDir, err := networkDirFlag("restart", args)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), commandWait)
	defer cancel()
	if err := tmpnet.RestartNetwork(ctx, logging.NoLog{}, networkDir); err != nil {
		return fmt.Errorf("restart rehearsal network: %w", err)
	}
	network, err := readNetwork(networkDir)
	if err != nil {
		return err
	}
	return emitSummary(network)
}

func bounceOne(args []string) error {
	networkDir, err := networkDirFlag("bounce-one", args)
	if err != nil {
		return err
	}
	network, err := readNetwork(networkDir)
	if err != nil {
		return err
	}
	if len(network.Nodes) != nodeCount {
		return fmt.Errorf("expected %d rehearsal nodes, found %d", nodeCount, len(network.Nodes))
	}
	node := network.Nodes[len(network.Nodes)-1]
	ctx, cancel := context.WithTimeout(context.Background(), commandWait)
	defer cancel()
	if err := node.Stop(ctx); err != nil {
		return fmt.Errorf("stop %s: %w", node.NodeID, err)
	}
	if err := network.StartNode(ctx, node); err != nil {
		return fmt.Errorf("restart %s: %w", node.NodeID, err)
	}
	if err := tmpnet.WaitForHealthyNodes(ctx, logging.NoLog{}, []*tmpnet.Node{node}); err != nil {
		return fmt.Errorf("wait for %s: %w", node.NodeID, err)
	}
	return emitSummary(network)
}

func stop(args []string) error {
	networkDir, err := networkDirFlag("stop", args)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), commandWait)
	defer cancel()
	if err := tmpnet.StopNetwork(ctx, logging.NoLog{}, networkDir); err != nil {
		return fmt.Errorf("stop rehearsal network: %w", err)
	}
	fmt.Printf("{\"schema\":\"PRESENCE_REHEARSAL_STOP_001\",\"network_dir\":%q,\"stopped\":true}\n", networkDir)
	return nil
}

func networkDirFlag(name string, args []string) (string, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	networkDir := flags.String("network-dir", "", "LOCALITY rehearsal network directory")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	if *networkDir == "" {
		return "", errors.New("--network-dir is required")
	}
	return filepath.Abs(*networkDir)
}

func readNetwork(networkDir string) (*tmpnet.Network, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandWait)
	defer cancel()
	network, err := tmpnet.ReadNetwork(ctx, logging.NoLog{}, networkDir)
	if err != nil {
		return nil, fmt.Errorf("read rehearsal network: %w", err)
	}
	return network, nil
}

func emitSummary(network *tmpnet.Network) error {
	if len(network.Subnets) != 1 || len(network.Subnets[0].Chains) != 1 {
		return errors.New("rehearsal network does not contain exactly one LOCALITY chain")
	}
	subnet := network.Subnets[0]
	chain := subnet.Chains[0]
	nodes := make([]nodeSummary, 0, len(network.Nodes))
	for _, node := range network.Nodes {
		nodes = append(nodes, nodeSummary{
			NodeID: node.NodeID.String(),
			URI:    node.GetAccessibleURI(),
		})
	}
	summary := networkSummary{
		Schema:       "PRESENCE_REHEARSAL_NETWORK_001",
		NetworkDir:   network.Dir,
		NetworkID:    network.GetNetworkID(),
		VMID:         chain.VMID.String(),
		SubnetID:     subnet.SubnetID.String(),
		BlockchainID: chain.ChainID.String(),
		Nodes:        nodes,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}

func checkedFile(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", absolute)
	}
	return absolute, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "locality rehearsal:", err)
	os.Exit(1)
}

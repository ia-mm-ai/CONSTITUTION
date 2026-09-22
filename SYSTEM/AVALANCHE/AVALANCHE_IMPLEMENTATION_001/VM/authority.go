package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	publicAuthoritySchema  = "PRESENCE_AVALANCHE_AUTHORITY_PUBLIC_001"
	privateAuthoritySchema = "PRESENCE_AVALANCHE_AUTHORITY_PRIVATE_001"
	testOnlyUsage          = "TEST_ONLY_LOCAL_REHEARSAL_NOT_ADMINISTRATOR_INPUT"
	deploymentPlanSchema   = "PRESENCE_AVALANCHE_DEPLOYMENT_PLAN_001"
)

type PublicAuthorityFile struct {
	Schema    string `json:"schema"`
	Scheme    string `json:"scheme"`
	KeyID     string `json:"key_id"`
	ActorID   string `json:"actor_id"`
	PublicKey string `json:"public_key"`
	Usage     string `json:"usage,omitempty"`
}

type PrivateAuthorityFile struct {
	Schema     string `json:"schema"`
	Scheme     string `json:"scheme"`
	KeyID      string `json:"key_id"`
	ActorID    string `json:"actor_id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Usage      string `json:"usage,omitempty"`
}

// generateTestAuthority produces a disposable ED25519 authority pair for local
// rehearsal only. It is never administrator input, never production custody, and
// the private document it writes is deleted by the qualification run.
func generateTestAuthority(directory string) (string, string, string, error) {
	if directory == "" {
		return "", "", "", errors.New("authority directory is required")
	}
	if err := ensurePrivateDirectory(directory); err != nil {
		return "", "", "", err
	}
	privatePath := filepath.Join(directory, "presence-avalanche-test-authority.private.json")
	publicPath := filepath.Join(directory, "presence-avalanche-test-authority.public.json")
	for _, path := range []string{privatePath, publicPath} {
		if _, err := os.Lstat(path); err == nil {
			return "", "", "", fmt.Errorf("refusing to overwrite %s", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", "", "", err
		}
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", "", err
	}
	publicDigest := sha256.Sum256(publicKey)
	keyID := "presence-test-" + hex.EncodeToString(publicDigest[:12])
	actorID := operationalActorID(publicKey)
	publicDocument := PublicAuthorityFile{
		Schema: publicAuthoritySchema, Scheme: "ED25519", KeyID: keyID, ActorID: actorID,
		PublicKey: hex.EncodeToString(publicKey), Usage: testOnlyUsage,
	}
	privateDocument := PrivateAuthorityFile{
		Schema: privateAuthoritySchema, Scheme: "ED25519", KeyID: keyID, ActorID: actorID,
		PublicKey: hex.EncodeToString(publicKey), PrivateKey: hex.EncodeToString(privateKey),
		Usage: testOnlyUsage,
	}
	privateBytes, err := json.MarshalIndent(&privateDocument, "", "  ")
	if err != nil {
		return "", "", "", err
	}
	publicBytes, err := json.MarshalIndent(&publicDocument, "", "  ")
	if err != nil {
		return "", "", "", err
	}
	if err := writeExclusive(privatePath, append(privateBytes, '\n'), 0o600); err != nil {
		return "", "", "", err
	}
	if err := writeExclusive(publicPath, append(publicBytes, '\n'), 0o644); err != nil {
		_ = os.Remove(privatePath)
		return "", "", "", err
	}
	return keyID, privatePath, publicPath, nil
}

func materializeGenesis(templatePath, adminInputPath, outputPath string) error {
	if err := refuseExistingOutput(outputPath); err != nil {
		return err
	}
	genesis, err := loadGenesisTemplate(templatePath)
	if err != nil {
		return err
	}
	adminBytes, err := os.ReadFile(adminInputPath)
	if err != nil {
		return fmt.Errorf("read administrator input: %w", err)
	}
	input, err := parseAdminInput(adminBytes)
	if err != nil {
		return err
	}
	publicKey, err := decodeLowerHex("authority.formation_public_key", input.Authority.FormationPublicKey, ed25519.PublicKeySize)
	if err != nil {
		return err
	}
	publicDigest := sha256.Sum256(publicKey)
	genesis.Domain.GenesisID = input.Instantiation.GenesisID
	genesis.Domain.Purpose = input.Instantiation.Purpose
	genesis.Locality.ID = input.Instantiation.LocalityID
	genesis.Locality.Authority = GenesisAuthority{
		Scheme:    "ED25519",
		KeyID:     "presence-" + hex.EncodeToString(publicDigest[:12]),
		PublicKey: input.Authority.FormationPublicKey,
		Scope:     "FORMATION_AND_BOOTSTRAP_ONLY",
	}
	genesis.Core = currentCore()
	genesis.Lineage = currentLineage()
	if err := genesis.Validate(); err != nil {
		return fmt.Errorf("materialized genesis is invalid: %w", err)
	}
	outputBytes, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return err
	}
	return writeExclusive(outputPath, append(outputBytes, '\n'), 0o644)
}

// loadGenesisTemplate reads the reusable template and refuses any template that
// already carries instantiation-bound identity, authority or core material.
func loadGenesisTemplate(templatePath string) (*Genesis, error) {
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("read genesis template: %w", err)
	}
	genesis := new(Genesis)
	if err := decodeStrict(templateBytes, genesis); err != nil {
		return nil, fmt.Errorf("decode genesis template: %w", err)
	}
	if genesis.Domain.GenesisID != "" || genesis.Domain.Purpose != "" || genesis.Locality.ID != "" {
		return nil, errors.New("genesis template must leave instantiation-bound identity fields empty")
	}
	if genesis.Locality.Authority.Scheme != "ED25519" ||
		genesis.Locality.Authority.KeyID != "MATERIALIZE-WITH-PRESENCE-AVALANCHE-VM" ||
		genesis.Locality.Authority.PublicKey != "" ||
		genesis.Locality.Authority.Scope != "FORMATION_AND_BOOTSTRAP_ONLY" {
		return nil, errors.New("genesis template contains an unexpected authority placeholder")
	}
	if len(genesis.Core.Files) != 0 || genesis.Core.CoreDigest != "" || genesis.Core.Commit != "" {
		return nil, errors.New("genesis template must not carry core references: they are injected from the core binding")
	}
	if len(genesis.Lineage.Predecessors) != 0 {
		return nil, errors.New("genesis template must not carry lineage predecessors: they are injected from the core binding")
	}
	return genesis, nil
}

// materializeDeploymentPlan emits the public, derived deployment plan that binds
// administrator input to the exact genesis bytes and the compiled core. It
// carries no private material of any kind.
func materializeDeploymentPlan(adminInputPath, genesisPath, outputPath string) error {
	if err := refuseExistingOutput(outputPath); err != nil {
		return err
	}
	adminBytes, err := os.ReadFile(adminInputPath)
	if err != nil {
		return fmt.Errorf("read administrator input: %w", err)
	}
	input, err := parseAdminInput(adminBytes)
	if err != nil {
		return err
	}
	genesisBytes, err := os.ReadFile(genesisPath)
	if err != nil {
		return fmt.Errorf("read genesis: %w", err)
	}
	genesis, err := parseGenesis(genesisBytes)
	if err != nil {
		return fmt.Errorf("validate genesis: %w", err)
	}
	if genesis.Domain.GenesisID != input.Instantiation.GenesisID || genesis.Locality.ID != input.Instantiation.LocalityID {
		return errors.New("genesis identity does not match administrator input")
	}
	if genesis.Locality.Authority.PublicKey != input.Authority.FormationPublicKey {
		return errors.New("genesis formation authority does not match administrator input")
	}
	digest := sha256.Sum256(genesisBytes)
	vmID, err := presenceVMID()
	if err != nil {
		return err
	}
	plan := DeploymentPlan{
		Schema:                deploymentPlanSchema,
		Implementation:        implementationID + "/" + implementationLevel,
		VMName:                vmName,
		VMID:                  vmID.String(),
		VMVersion:             vmVersion,
		PluginFileName:        vmID.String(),
		AvalancheGo:           avalancheGoVersion,
		AvalancheGoCommit:     avalancheGoCommit,
		RPCChainVM:            uint32(rpcChainVMProtocol()),
		Network:               input.Instantiation.Network,
		FinalName:             input.Instantiation.FinalName,
		LocalityID:            input.Instantiation.LocalityID,
		GenesisID:             input.Instantiation.GenesisID,
		GenesisSHA256:         hex.EncodeToString(digest[:]),
		GenesisBytes:          uint64(len(genesisBytes)),
		CoreCommit:            coreCommit,
		CoreDigest:            coreDigest,
		Architecture:          input.Architecture,
		Control:               input.Control,
		Validators:            input.Validators,
		ContinuityAuthorities: uint32(len(input.Authority.ContinuityPublicKeys)),
		ExhaustionPlanned:     input.Authority.ExhaustionPlanned,
		Status:                "READY_FOR_ADMIN_INSTANTIATION",
		Effect: []string{
			"NO_DEPLOYMENT_PERFORMED",
			"NO_NETWORK_WRITE_PERFORMED",
			"NO_FORMATION_CLAIMED",
			"NO_ADOPTION_CLAIMED",
			"NO_PRESENCE_CLAIMED",
		},
	}
	planBytes, err := json.MarshalIndent(&plan, "", "  ")
	if err != nil {
		return err
	}
	if err := refuseForbiddenKeys(planBytes); err != nil {
		return err
	}
	return writeExclusive(outputPath, append(planBytes, '\n'), 0o644)
}

type DeploymentPlan struct {
	Schema                string            `json:"schema"`
	Implementation        string            `json:"implementation"`
	VMName                string            `json:"vm_name"`
	VMID                  string            `json:"vm_id"`
	VMVersion             string            `json:"vm_version"`
	PluginFileName        string            `json:"plugin_file_name"`
	AvalancheGo           string            `json:"avalanchego_version"`
	AvalancheGoCommit     string            `json:"avalanchego_commit"`
	RPCChainVM            uint32            `json:"rpcchainvm_protocol"`
	Network               string            `json:"network"`
	FinalName             string            `json:"final_name"`
	LocalityID            string            `json:"locality_id"`
	GenesisID             string            `json:"genesis_id"`
	GenesisSHA256         string            `json:"genesis_sha256"`
	GenesisBytes          uint64            `json:"genesis_bytes"`
	CoreCommit            string            `json:"core_commit"`
	CoreDigest            string            `json:"core_digest"`
	Architecture          AdminArchitecture `json:"architecture"`
	Control               AdminControl      `json:"control"`
	Validators            []AdminValidator  `json:"validators"`
	ContinuityAuthorities uint32            `json:"declared_continuity_authorities"`
	ExhaustionPlanned     bool              `json:"formation_authority_exhaustion_planned"`
	Status                string            `json:"status"`
	Effect                []string          `json:"effect_ceiling"`
}

// checkCoreBindingFile verifies that an on-disk CORE_BINDING document matches the
// core binding compiled into this binary, byte value for byte value.
func checkCoreBindingFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	document := new(coreBindingDocument)
	if err := decodeStrictLenient(raw, document); err != nil {
		return "", err
	}
	if document.Schema != "PRESENCE_AVALANCHE_CORE_BINDING_001" {
		return "", errors.New("unsupported core binding schema")
	}
	if document.Repository != coreRepository || document.StartingCommit != coreCommit || document.CoreDigest != coreDigest {
		return "", errors.New("core binding repository, commit or digest differs from the compiled reference")
	}
	if document.VMID != expectedVMID || document.VMName != vmName || document.VMVersion != vmVersion {
		return "", errors.New("core binding identity differs from the compiled reference")
	}
	if len(document.Files) != len(coreFiles) {
		return "", errors.New("core binding file count differs from the compiled reference")
	}
	for index, want := range coreFiles {
		if document.Files[index] != want {
			return "", fmt.Errorf("core binding file %d differs from the compiled reference", index)
		}
	}
	if computeCoreDigest(document.Files) != coreDigest {
		return "", errors.New("core digest does not recompute from the bound core files")
	}
	return fmt.Sprintf("core binding matches: commit=%s core_digest=%s files=%d", coreCommit, coreDigest, len(coreFiles)), nil
}

type coreBindingDocument struct {
	Schema         string            `json:"schema"`
	VMName         string            `json:"vm_identity"`
	VMID           string            `json:"vm_id"`
	VMVersion      string            `json:"vm_version"`
	Repository     string            `json:"repository"`
	StartingCommit string            `json:"starting_commit"`
	CoreDigest     string            `json:"core_digest"`
	Files          []CoreFileBinding `json:"files"`
}

func refuseExistingOutput(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("refusing to overwrite existing output: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func ensurePrivateDirectory(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(path, 0o700)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("authority directory %s must not be accessible by group or other users", path)
	}
	return nil
}

func writeExclusive(path string, contents []byte, permissions os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, permissions)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(contents); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

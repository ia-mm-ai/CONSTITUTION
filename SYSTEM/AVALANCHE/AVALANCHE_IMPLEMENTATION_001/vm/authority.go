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
)

type PublicAuthorityFile struct {
	Schema    string `json:"schema"`
	Scheme    string `json:"scheme"`
	KeyID     string `json:"key_id"`
	ActorID   string `json:"actor_id"`
	PublicKey string `json:"public_key"`
}

type PrivateAuthorityFile struct {
	Schema     string `json:"schema"`
	Scheme     string `json:"scheme"`
	KeyID      string `json:"key_id"`
	ActorID    string `json:"actor_id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func generateAuthority(directory string) (string, string, string, error) {
	if directory == "" {
		return "", "", "", errors.New("authority directory is required")
	}
	if err := ensurePrivateDirectory(directory); err != nil {
		return "", "", "", err
	}
	privatePath := filepath.Join(directory, "locality-authority.private.json")
	publicPath := filepath.Join(directory, "locality-authority.public.json")
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
	keyID := "locality-" + hex.EncodeToString(publicDigest[:12])
	actorID := operationalActorID(publicKey)
	publicDocument := PublicAuthorityFile{
		Schema: publicAuthoritySchema, Scheme: "ED25519", KeyID: keyID, ActorID: actorID, PublicKey: hex.EncodeToString(publicKey),
	}
	privateDocument := PrivateAuthorityFile{
		Schema: privateAuthoritySchema, Scheme: "ED25519", KeyID: keyID, ActorID: actorID,
		PublicKey: hex.EncodeToString(publicKey), PrivateKey: hex.EncodeToString(privateKey),
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

func materializeGenesis(templatePath, deploymentDescriptorPath, publicAuthorityPath, outputPath string) error {
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read genesis template: %w", err)
	}
	genesis := new(Genesis)
	if err := decodeStrict(templateBytes, genesis); err != nil {
		return fmt.Errorf("decode genesis template: %w", err)
	}
	if genesis.Domain.GenesisID != "" ||
		genesis.Domain.Purpose != "" ||
		genesis.Locality.ID != "" ||
		genesis.Locality.SourceReference != "" ||
		genesis.Locality.SourceSHA256 != "" {
		return errors.New("genesis template must leave deployment-bound identity and source fields empty")
	}
	if genesis.Locality.Authority.Scheme != "ED25519" ||
		genesis.Locality.Authority.KeyID != "MATERIALIZE-WITH-PRESENCE-AVALANCHE-VM" ||
		genesis.Locality.Authority.PublicKey != "" ||
		genesis.Locality.Authority.Scope != "FORMATION_AND_BOOTSTRAP_ONLY" {
		return errors.New("genesis template contains an unexpected authority placeholder")
	}
	deploymentBytes, err := os.ReadFile(deploymentDescriptorPath)
	if err != nil {
		return fmt.Errorf("read deployment descriptor: %w", err)
	}
	deployment := new(DeploymentDescriptor)
	if err := decodeStrict(deploymentBytes, deployment); err != nil {
		return fmt.Errorf("decode deployment descriptor: %w", err)
	}
	if err := deployment.Validate(); err != nil {
		return fmt.Errorf("validate deployment descriptor: %w", err)
	}
	publicBytes, err := os.ReadFile(publicAuthorityPath)
	if err != nil {
		return fmt.Errorf("read public authority: %w", err)
	}
	publicAuthority := new(PublicAuthorityFile)
	if err := decodeStrict(publicBytes, publicAuthority); err != nil {
		return fmt.Errorf("decode public authority: %w", err)
	}
	if publicAuthority.Schema != publicAuthoritySchema || publicAuthority.Scheme != "ED25519" {
		return errors.New("unsupported public authority document")
	}
	if err := requireSafeID("public authority key_id", publicAuthority.KeyID); err != nil {
		return err
	}
	publicKey, err := decodeLowerHex("public authority public_key", publicAuthority.PublicKey, ed25519.PublicKeySize)
	if err != nil {
		return err
	}
	if publicAuthority.ActorID != operationalActorID(publicKey) {
		return errors.New("public authority actor_id does not match its public key")
	}
	genesis.Domain.GenesisID = deployment.GenesisID
	genesis.Domain.Purpose = deployment.Purpose
	genesis.Locality.ID = deployment.LocalityID
	genesis.Locality.SourceReference = deployment.SourceReference
	genesis.Locality.SourceSHA256 = deployment.SourceSHA256
	genesis.Locality.Authority = GenesisAuthority{
		Scheme: "ED25519", KeyID: publicAuthority.KeyID, PublicKey: publicAuthority.PublicKey,
		Scope: "FORMATION_AND_BOOTSTRAP_ONLY",
	}
	if err := genesis.Validate(); err != nil {
		return fmt.Errorf("materialized genesis is invalid: %w", err)
	}
	outputBytes, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return err
	}
	if err := writeExclusive(outputPath, append(outputBytes, '\n'), 0o644); err != nil {
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

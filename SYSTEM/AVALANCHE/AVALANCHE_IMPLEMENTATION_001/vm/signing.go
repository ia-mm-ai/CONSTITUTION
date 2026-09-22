package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

func signUnsignedFile(unsignedPath, privateAuthorityPath, outputPath string) error {
	unsignedBytes, err := os.ReadFile(unsignedPath)
	if err != nil {
		return fmt.Errorf("read unsigned transition: %w", err)
	}
	unsigned := new(UnsignedTransition)
	if err := decodeStrict(unsignedBytes, unsigned); err != nil {
		return fmt.Errorf("decode unsigned transition: %w", err)
	}
	privateAuthority, privateKey, err := readPrivateAuthority(privateAuthorityPath)
	if err != nil {
		return err
	}
	if unsigned.ActorPublicKey != "" && unsigned.ActorPublicKey != privateAuthority.PublicKey {
		return errors.New("unsigned transition actor_public_key does not match the selected private authority")
	}
	unsigned.ActorPublicKey = privateAuthority.PublicKey
	canonicalPayload, err := validateAndCanonicalizePayload(unsigned.Operation, unsigned.Payload)
	if err != nil {
		return fmt.Errorf("validate transition payload: %w", err)
	}
	unsigned.Payload = canonicalPayload
	transition, err := signTransition(*unsigned, privateKey)
	if err != nil {
		return fmt.Errorf("sign transition: %w", err)
	}
	transitionBytes, err := transition.Bytes()
	if err != nil {
		return err
	}
	if err := writeExclusive(outputPath, transitionBytes, 0o644); err != nil {
		return err
	}
	return nil
}

func readPrivateAuthority(path string) (*PrivateAuthorityFile, ed25519.PrivateKey, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, nil, fmt.Errorf("private authority permissions are too broad: %s", path)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	document := new(PrivateAuthorityFile)
	if err := decodeStrict(contents, document); err != nil {
		return nil, nil, err
	}
	if document.Schema != privateAuthoritySchema || document.Scheme != "ED25519" {
		return nil, nil, errors.New("unsupported private authority document")
	}
	if err := requireSafeID("private authority key_id", document.KeyID); err != nil {
		return nil, nil, err
	}
	publicKey, err := decodeLowerHex("private authority public_key", document.PublicKey, ed25519.PublicKeySize)
	if err != nil {
		return nil, nil, err
	}
	privateKeyBytes, err := decodeLowerHex("private authority private_key", document.PrivateKey, ed25519.PrivateKeySize)
	if err != nil {
		return nil, nil, err
	}
	privateKey := ed25519.PrivateKey(privateKeyBytes)
	derivedPublic := privateKey.Public().(ed25519.PublicKey)
	if !ed25519.PublicKey(publicKey).Equal(derivedPublic) || hex.EncodeToString(derivedPublic) != document.PublicKey {
		return nil, nil, errors.New("private authority public/private key mismatch")
	}
	if document.ActorID != operationalActorID(publicKey) {
		return nil, nil, errors.New("private authority actor_id does not match its public key")
	}
	return document, privateKey, nil
}

type TransitionDraftRequest struct {
	Operation      string          `json:"operation"`
	ActorID        string          `json:"actor_id"`
	ActorPublicKey string          `json:"actor_public_key"`
	LocusID        string          `json:"locus_id"`
	ObservedAt     int64           `json:"observed_at"`
	Payload        json.RawMessage `json:"payload"`
}

func (vm *VM) draftTransition(request *TransitionDraftRequest) (*UnsignedTransition, error) {
	effect, exists := operationEffects[request.Operation]
	if !exists {
		return nil, fmt.Errorf("unsupported operation %q", request.Operation)
	}
	if err := requireSafeID("actor_id", request.ActorID); err != nil {
		return nil, err
	}
	if _, err := decodeLowerHex("actor_public_key", request.ActorPublicKey, ed25519.PublicKeySize); err != nil {
		return nil, err
	}
	if request.LocusID != "" {
		if err := requireSafeID("locus_id", request.LocusID); err != nil {
			return nil, err
		}
	}
	if request.ObservedAt < 0 {
		return nil, errors.New("observed_at cannot be negative")
	}
	if request.ObservedAt > time.Now().UTC().Add(maxClockSkew).Unix() {
		return nil, errors.New("observed_at exceeds the local future-skew ceiling")
	}
	payload, err := validateAndCanonicalizePayload(request.Operation, request.Payload)
	if err != nil {
		return nil, err
	}

	vm.lock.Lock()
	defer vm.lock.Unlock()
	if !vm.initialized {
		return nil, errNotInitialized
	}
	preferred, err := vm.getBlockLocked(vm.preferredID)
	if err != nil || preferred.postState == nil {
		return nil, errInvalidPreference
	}
	preview, err := preferred.postState.clone()
	if err != nil {
		return nil, err
	}
	for _, pendingID := range vm.pendingOrder {
		pending, exists := vm.pending[pendingID]
		if !exists {
			continue
		}
		if _, err := preview.apply(vm.genesis, pending, pendingID); err != nil {
			return nil, fmt.Errorf("pending transition %s became invalid: %w", pendingID, err)
		}
	}
	if knownKey, known := preview.ActorKeys[request.ActorID]; known && knownKey != request.ActorPublicKey {
		return nil, errors.New("actor_public_key differs from the actor's established local binding")
	}
	if _, known := preview.ActorKeys[request.ActorID]; !known && request.Operation != opPresentForm {
		return nil, errors.New("unknown actor must first draft PRESENT_FORM")
	}
	if _, known := preview.ActorKeys[request.ActorID]; !known && request.ActorID != operationalActorIDFromHex(request.ActorPublicKey) {
		return nil, errors.New("new participant actor_id must be derived from actor_public_key")
	}
	unsigned := &UnsignedTransition{
		Schema:                  transitionSchema,
		Operation:               request.Operation,
		Revision:                preview.Revision + 1,
		PreviousStateCommitment: preview.StateCommitment,
		ActorID:                 request.ActorID,
		ActorPublicKey:          request.ActorPublicKey,
		Nonce:                   preview.NextNonces[request.ActorID],
		LocusID:                 request.LocusID,
		ObservedAt:              request.ObservedAt,
		Effect:                  effect,
		Payload:                 payload,
	}
	if err := preview.validateOperationScope(unsigned); err != nil {
		return nil, err
	}
	return unsigned, nil
}

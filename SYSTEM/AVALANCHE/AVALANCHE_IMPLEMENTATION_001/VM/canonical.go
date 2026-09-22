package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"unicode"
	"unicode/utf8"
)

var safeIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

func decodeStrict(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing JSON value")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

// decodeStrictLenient decodes a document that legitimately carries more fields
// than the checked subset, while still refusing trailing content.
func decodeStrictLenient(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing JSON value")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

func decodeCanonical(data []byte, value any) error {
	if err := decodeStrict(data, value); err != nil {
		return err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, canonical) {
		return errors.New("JSON is not in the canonical PRESENCE AVALANCHE encoding")
	}
	return nil
}

func requireSafeID(name, value string) error {
	if !safeIDPattern.MatchString(value) {
		return fmt.Errorf("%s must match %s", name, safeIDPattern)
	}
	return nil
}

func requireBoundedText(name, value string, maximumBytes int) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	if len(value) > maximumBytes {
		return fmt.Errorf("%s exceeds %d UTF-8 bytes", name, maximumBytes)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s is not valid UTF-8", name)
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return fmt.Errorf("%s contains a control character", name)
		}
	}
	return nil
}

func decodeLowerHex(name, value string, byteLength int) ([]byte, error) {
	if len(value) != byteLength*2 {
		return nil, fmt.Errorf("%s must contain exactly %d lowercase hexadecimal characters", name, byteLength*2)
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if hex.EncodeToString(decoded) != value {
		return nil, fmt.Errorf("%s must use lowercase hexadecimal", name)
	}
	return decoded, nil
}

func requireDigest(name, value string) error {
	_, err := decodeLowerHex(name, value, 32)
	return err
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func requireUniqueNonEmpty(name string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		if value == "" {
			return fmt.Errorf("%s[%d] is empty", name, i)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s contains duplicate %q", name, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func requireExactSet(name string, values, expected []string) error {
	if err := requireUniqueNonEmpty(name, values); err != nil {
		return err
	}
	if len(values) != len(expected) {
		return fmt.Errorf("%s must contain exactly %d values", name, len(expected))
	}
	for _, value := range expected {
		if !contains(values, value) {
			return fmt.Errorf("%s is missing %s", name, value)
		}
	}
	return nil
}

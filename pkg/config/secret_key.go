package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gopkg.in/yaml.v3"
)

// secretKey represents a secure byte slice used for encryption or token generation.
// It provides methods to marshal and unmarshal the key into/from YAML.
type secretKey []byte

// MarshalYAML converts the secretKey to a hexadecimal string for YAML serialization.
//
// Returns:
//   - interface{}: A hexadecimal string representation of the secretKey.
//   - error: An error if marshaling fails (unlikely in this implementation).
func (k secretKey) MarshalYAML() (interface{}, error) {
	return hex.EncodeToString(k), nil
}

// UnmarshalYAML converts a hexadecimal string from YAML into a secretKey.
//
// Parameters:
//   - node: A YAML node containing the hexadecimal string.
//
// Returns:
//   - error: An error if the hexadecimal string is invalid or decoding fails.
func (k *secretKey) UnmarshalYAML(node *yaml.Node) error {
	value := node.Value
	ba, err := hex.DecodeString(value)
	if err != nil {
		return err
	}
	*k = ba
	return nil
}

// generateSecureHex generates a cryptographically secure random hexadecimal string of the specified length.
//
// Parameters:
//   - length: The length of the byte array to generate.
//
// Returns:
//   - secretKey: A hexadecimal-encoded secure key.
//   - error: An error if random byte generation fails.
func generateSecureHex(length int) (secretKey, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return []byte(hex.EncodeToString(bytes)), nil
}

package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gopkg.in/yaml.v3"
)

// secretKey is a secure byte slice that is typically used for encryption or token generation.
// It can be serialized into a hexadecimal string for easy storage in YAML files
// and deserialized back into its original form.
type secretKey []byte

// MarshalYAML converts the SecretKey into a hexadecimal string format.
// This allows the key to be stored in YAML files in a human-readable way.
// The function ensures the key is safely serialized for external storage.
func (k secretKey) MarshalYAML() (interface{}, error) {
	return hex.EncodeToString(k), nil
}

// UnmarshalYAML reads a hexadecimal string from a YAML node and decodes it back into a SecretKey.
// This process ensures that the key stored in the YAML file is correctly parsed
// and restored to its original byte slice format.
func (k *secretKey) UnmarshalYAML(node *yaml.Node) error {
	value := node.Value
	decodedBytes, err := hex.DecodeString(value)
	if err != nil {
		return err
	}
	*k = decodedBytes
	return nil
}

// generateSecureHex creates a secure random key and encodes it as a hexadecimal string.
// The function generates a random byte array of the specified length, ensuring
// the key is cryptographically secure and suitable for sensitive operations.
func generateSecureHex(length int) (secretKey, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate random bytes: %w", err)
	}
	return []byte(hex.EncodeToString(bytes)), nil
}

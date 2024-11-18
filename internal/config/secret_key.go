package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gopkg.in/yaml.v3"
)

type secretKey []byte

func (k secretKey) MarshalYAML() (interface{}, error) {
	return hex.EncodeToString(k), nil
}

func (k *secretKey) UnmarshalYAML(node *yaml.Node) error {
	value := node.Value
	ba, err := hex.DecodeString(value)
	if err != nil {
		return err
	}
	*k = ba
	return nil
}

func generateSecureHex(length int) (secretKey, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return []byte(hex.EncodeToString(bytes)), nil
}

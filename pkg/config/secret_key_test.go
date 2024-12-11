package config

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// TestMarshalYAML ensures that secretKey is marshaled correctly.
func TestSecretKeyMarshalYAML(t *testing.T) {
	key := secretKey([]byte{0x12, 0x34, 0x56, 0x78})
	data, err := key.MarshalYAML()
	assert.NoError(t, err, "MarshalYAML should not return an error")

	encoded, ok := data.(string)
	assert.True(t, ok, "Marshaled data should be a string")
	assert.Equal(t, "12345678", encoded, "Hex encoding of the secretKey should be correct")
}

// TestUnmarshalYAML ensures that secretKey is unmarshaled correctly.
func TestSecretKeyUnmarshalYAML(t *testing.T) {
	var key secretKey
	node := yaml.Node{
		Value: "12345678",
	}

	err := key.UnmarshalYAML(&node)
	assert.NoError(t, err, "UnmarshalYAML should not return an error")
	assert.Equal(t, secretKey([]byte{0x12, 0x34, 0x56, 0x78}), key, "Unmarshaled key should match the original bytes")
}

// TestUnmarshalYAMLError ensures that invalid hex strings return an error during unmarshaling.
func TestSecretKeyUnmarshalYAMLError(t *testing.T) {
	var key secretKey
	node := yaml.Node{
		Value: "invalidhex",
	}

	err := key.UnmarshalYAML(&node)
	assert.Error(t, err, "UnmarshalYAML should return an error for invalid hex strings")
}

// TestGenerateSecureHex ensures that generateSecureHex produces a key of the correct length and format.
func TestGenerateSecureHex(t *testing.T) {
	length := 16
	key, err := generateSecureHex(length)

	assert.NoError(t, err, "generateSecureHex should not return an error")
	assert.Len(t, key, length*2, "Generated key should have the correct length in hex (length * 2)")
	_, decodeErr := hex.DecodeString(string(key))
	assert.NoError(t, decodeErr, "Generated key should be a valid hex string")
}

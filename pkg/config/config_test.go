package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefaultConfig ensures that defaultConfig generates a valid Config structure.
func TestDefaultConfig(t *testing.T) {
	cfg, err := defaultConfig()

	assert.NoError(t, err, "defaultConfig should not return an error")
	assert.NotNil(t, cfg, "defaultConfig should return a valid Config")

	// Verify some default values
	assert.Equal(t, 8080, cfg.Server.Port, "Default server port should be 8080")
	assert.True(t, cfg.DevMode.Enabled, "DevMode should be enabled by default")
	assert.NotEmpty(t, cfg.Security.Csrf.Token.Secret, "CSRF Token Secret should be generated")
	assert.NotEmpty(t, cfg.Security.Session.Token.Secret, "Session Token Secret should be generated")
	assert.Equal(t, "/favicon.ico", cfg.Routes.Favicon, "Default favicon route should be set")
}

// TestCreateDefaultConfig ensures that createDefaultConfig writes a valid config file.
func TestCreateDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test_config.yml")

	// Call createDefaultConfig
	err := createDefaultConfig(cfgPath)

	assert.NoError(t, err, "createDefaultConfig should not return an error")
	assert.FileExists(t, cfgPath, "createDefaultConfig should create a config file")

	// Read and verify the file
	data, err := os.ReadFile(cfgPath)
	assert.NoError(t, err, "Reading the created config file should not return an error")
	assert.Contains(t, string(data), "csrf_token", "Config file should contain CSRF token information")
}

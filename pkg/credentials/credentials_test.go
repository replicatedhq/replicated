package credentials

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFilePath_BackwardCompatibility(t *testing.T) {
	// Create a temporary directory to act as home
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	// Create the legacy config file
	legacyDir := filepath.Join(tempHome, ".replicated")
	legacyPath := filepath.Join(legacyDir, "config.yaml")
	err := os.MkdirAll(legacyDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(legacyPath, []byte("{}"), 0600)
	require.NoError(t, err)

	// The configFilePath should return the legacy path when it exists
	path := configFilePath()
	assert.Equal(t, legacyPath, path)
}

func TestConfigFilePath_XDGCompliant(t *testing.T) {
	// Create a temporary directory to act as home
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
		xdg.Reload()
	}()
	os.Setenv("HOME", tempHome)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))

	// Reset xdg package's cached values by re-reading environment
	xdg.Reload()

	// Ensure legacy path does NOT exist
	legacyPath := filepath.Join(tempHome, ".replicated", "config.yaml")

	// The configFilePath should return the XDG-compliant path
	path := configFilePath()
	assert.NotEqual(t, legacyPath, path)
	assert.Contains(t, path, "replicated")
	assert.Contains(t, path, "config.yaml")
	assert.Contains(t, path, xdg.ConfigHome)
}

func TestSetCurrentCredentials_XDGPath(t *testing.T) {
	// Create a temporary directory to act as home
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
		xdg.Reload()
	}()
	os.Setenv("HOME", tempHome)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))

	// Reset xdg package's cached values
	xdg.Reload()

	// Set credentials
	err := SetCurrentCredentials("test-token-12345")
	require.NoError(t, err)

	// Verify the file was created in the XDG-compliant location
	expectedPath := filepath.Join(xdg.ConfigHome, "replicated", "config.yaml")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err, "Config file should be created at XDG path: %s", expectedPath)

	// Verify the legacy path was NOT used
	legacyPath := filepath.Join(tempHome, ".replicated", "config.yaml")
	_, err = os.Stat(legacyPath)
	assert.True(t, os.IsNotExist(err), "Legacy config file should not exist")
}

func TestSetCurrentCredentials_LegacyPath(t *testing.T) {
	// Create a temporary directory to act as home
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	// Create the legacy config file first
	legacyDir := filepath.Join(tempHome, ".replicated")
	legacyPath := filepath.Join(legacyDir, "config.yaml")
	err := os.MkdirAll(legacyDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(legacyPath, []byte(`{"token":"old-token"}`), 0600)
	require.NoError(t, err)

	// Set new credentials
	err = SetCurrentCredentials("new-token-12345")
	require.NoError(t, err)

	// Verify the file was updated at the legacy location
	content, err := os.ReadFile(legacyPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "new-token-12345")
}

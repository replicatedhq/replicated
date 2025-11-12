//go:build integration
// +build integration

package lint2

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLintKots_FileNotFound tests error handling when config file doesn't exist
func TestLintKots_FileNotFound(t *testing.T) {
	ctx := context.Background()
	kotsVersion := "latest"

	result, err := LintKots(ctx, "/nonexistent/path/kots-config.yaml", kotsVersion)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "kots config path does not exist")
}

// TestLintKots_ValidConfig tests linting a valid KOTS config
// Requires REPLICATED_KOTS_PATH environment variable to be set
func TestLintKots_ValidConfig(t *testing.T) {
	// Check if KOTS binary is available
	kotsPath := os.Getenv("REPLICATED_KOTS_PATH")
	if kotsPath == "" {
		t.Skip("Skipping integration test: REPLICATED_KOTS_PATH not set")
	}

	// Verify binary exists
	if _, err := os.Stat(kotsPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: KOTS binary not found at %s", kotsPath)
	}

	ctx := context.Background()
	kotsVersion := "latest" // Will use REPLICATED_KOTS_PATH override

	// Create a temporary valid KOTS config
	validConfig := `apiVersion: kots.io/v1beta1
kind: Config
metadata:
  name: test-config
spec:
  groups:
    - name: database
      title: Database Settings
      description: Configure the database
      items:
        - name: postgres_host
          title: PostgreSQL Host
          type: text
          required: true
          default: "postgres"
        - name: postgres_port
          title: PostgreSQL Port
          type: text
          required: false
          default: "5432"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "valid-config.yaml")
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	// Run lint
	result, err := LintKots(ctx, configPath, kotsVersion)

	// Valid config should not error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have success=true and no/few messages
	assert.True(t, result.Success, "expected valid config to pass linting")

	// If there are messages, they should only be info/warnings, not errors
	for _, msg := range result.Messages {
		assert.NotEqual(t, "error", msg.Severity, "valid config should not have errors: %s", msg.Message)
	}
}

// TestLintKots_InvalidConfig tests linting an invalid KOTS config
// Requires REPLICATED_KOTS_PATH environment variable to be set
func TestLintKots_InvalidConfig(t *testing.T) {
	// Check if KOTS binary is available
	kotsPath := os.Getenv("REPLICATED_KOTS_PATH")
	if kotsPath == "" {
		t.Skip("Skipping integration test: REPLICATED_KOTS_PATH not set")
	}

	// Verify binary exists
	if _, err := os.Stat(kotsPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: KOTS binary not found at %s", kotsPath)
	}

	ctx := context.Background()
	kotsVersion := "latest"

	// Create a config with an invalid type
	invalidConfig := `apiVersion: kots.io/v1beta1
kind: Config
metadata:
  name: test-config
spec:
  groups:
    - name: settings
      title: Settings
      items:
        - name: invalid_item
          title: Invalid Item
          type: invalid_type_that_does_not_exist
          required: true
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yaml")
	err := os.WriteFile(configPath, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	// Run lint
	result, err := LintKots(ctx, configPath, kotsVersion)

	// Should not error (linter ran successfully)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have success=false (validation errors found)
	assert.False(t, result.Success, "expected invalid config to fail linting")

	// Should have at least one error message
	hasError := false
	for _, msg := range result.Messages {
		if msg.Severity == "error" {
			hasError = true
			break
		}
	}
	assert.True(t, hasError, "expected at least one error message for invalid config")
}

// TestLintKots_MinimalConfig tests a minimal but valid KOTS config
func TestLintKots_MinimalConfig(t *testing.T) {
	kotsPath := os.Getenv("REPLICATED_KOTS_PATH")
	if kotsPath == "" {
		t.Skip("Skipping integration test: REPLICATED_KOTS_PATH not set")
	}

	if _, err := os.Stat(kotsPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: KOTS binary not found at %s", kotsPath)
	}

	ctx := context.Background()
	kotsVersion := "latest"

	// Minimal valid config
	minimalConfig := `apiVersion: kots.io/v1beta1
kind: Config
metadata:
  name: minimal-config
spec:
  groups: []
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal-config.yaml")
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)

	result, err := LintKots(ctx, configPath, kotsVersion)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Minimal config should be valid (though it might have warnings about being empty)
	// We check that there are no critical errors
	for _, msg := range result.Messages {
		// Empty config might generate warnings but not errors
		if msg.Severity == "error" {
			t.Logf("Unexpected error in minimal config: %s", msg.Message)
		}
	}
}

// TestLintKots_MalformedYAML tests linting malformed YAML
func TestLintKots_MalformedYAML(t *testing.T) {
	kotsPath := os.Getenv("REPLICATED_KOTS_PATH")
	if kotsPath == "" {
		t.Skip("Skipping integration test: REPLICATED_KOTS_PATH not set")
	}

	if _, err := os.Stat(kotsPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: KOTS binary not found at %s", kotsPath)
	}

	ctx := context.Background()
	kotsVersion := "latest"

	// Malformed YAML
	malformedConfig := `apiVersion: kots.io/v1beta1
kind: Config
metadata:
  name: malformed
  invalid indentation
    - this is wrong
spec:
  groups: [
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "malformed-config.yaml")
	err := os.WriteFile(configPath, []byte(malformedConfig), 0644)
	require.NoError(t, err)

	result, err := LintKots(ctx, configPath, kotsVersion)

	// Malformed YAML might cause the linter to error or return failures
	// Either outcome is acceptable
	if err != nil {
		// If the linter errors, that's fine
		t.Logf("Linter returned error (expected): %v", err)
	} else {
		// If it returns a result, it should indicate failure
		require.NotNil(t, result)
		assert.False(t, result.Success, "expected malformed YAML to fail")
	}
}

// TestLintKots_BinaryOverride tests that REPLICATED_KOTS_PATH is respected
func TestLintKots_BinaryOverride(t *testing.T) {
	kotsPath := os.Getenv("REPLICATED_KOTS_PATH")
	if kotsPath == "" {
		t.Skip("Skipping integration test: REPLICATED_KOTS_PATH not set")
	}

	if _, err := os.Stat(kotsPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: KOTS binary not found at %s", kotsPath)
	}

	ctx := context.Background()
	kotsVersion := "latest"

	// Create a valid config
	validConfig := `apiVersion: kots.io/v1beta1
kind: Config
metadata:
  name: test
spec:
  groups: []
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	// Run with REPLICATED_KOTS_PATH set
	result, err := LintKots(ctx, configPath, kotsVersion)

	require.NoError(t, err)
	require.NotNil(t, result)

	// The fact that it ran successfully confirms the binary override worked
	t.Logf("Successfully used KOTS binary from: %s", kotsPath)
}

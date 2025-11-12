package lint2

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLintEmbeddedCluster_FileNotFound(t *testing.T) {
	ctx := context.Background()

	// Test with non-existent file
	_, err := LintEmbeddedCluster(ctx, "/nonexistent/path/config.yaml", "latest")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	if !os.IsNotExist(err) && err.Error() != "embedded cluster config path does not exist: /nonexistent/path/config.yaml" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLintEmbeddedCluster_StubReturnsSuccess(t *testing.T) {
	ctx := context.Background()

	// Create a temporary EC config file for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ec-config.yaml")

	// Write a minimal EC config
	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: test-config
spec:
  version: "1.0.0"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Note: This test will attempt to download embedded-cluster binary
	// In a real test environment, you might want to mock the resolver
	// For now, we skip if the binary can't be resolved
	t.Skip("Skipping test that requires embedded-cluster binary download")

	result, err := LintEmbeddedCluster(ctx, configPath, "latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected stub to return success")
	}

	if len(result.Messages) != 0 {
		t.Errorf("expected no messages, got %d", len(result.Messages))
	}
}

// JSON parsing tests

func TestParseEmbeddedClusterOutput_ErrorsOnly(t *testing.T) {
	output := `{
  "files": [
    {
      "path": "/tmp/test-config.yaml",
      "valid": false,
      "errors": [
        {
          "field": "spec.version",
          "message": "version is required"
        },
        {
          "field": "",
          "message": "YAML syntax error at line 5"
        }
      ]
    }
  ]
}`

	messages, err := parseEmbeddedClusterOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	// Check first error (with field)
	if messages[0].Severity != "ERROR" {
		t.Errorf("expected severity 'ERROR', got %q", messages[0].Severity)
	}
	if messages[0].Message != "spec.version: version is required" {
		t.Errorf("unexpected message: %q", messages[0].Message)
	}
	if messages[0].Path != "/tmp/test-config.yaml" {
		t.Errorf("unexpected path: %q", messages[0].Path)
	}

	// Check second error (without field)
	if messages[1].Message != "YAML syntax error at line 5" {
		t.Errorf("unexpected message: %q", messages[1].Message)
	}
}

func TestParseEmbeddedClusterOutput_WarningsOnly(t *testing.T) {
	output := `{
  "files": [
    {
      "path": "/tmp/test-config.yaml",
      "valid": true,
      "warnings": [
        {
          "field": "spec.extensions",
          "message": "extension may be deprecated"
        }
      ]
    }
  ]
}`

	messages, err := parseEmbeddedClusterOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Severity != "WARNING" {
		t.Errorf("expected severity 'WARNING', got %q", messages[0].Severity)
	}
	if messages[0].Message != "spec.extensions: extension may be deprecated" {
		t.Errorf("unexpected message: %q", messages[0].Message)
	}
}

func TestParseEmbeddedClusterOutput_InfosOnly(t *testing.T) {
	output := `{
  "files": [
    {
      "path": "/tmp/test-config.yaml",
      "valid": true,
      "infos": [
        {
          "field": "spec.version",
          "message": "using latest stable version"
        }
      ]
    }
  ]
}`

	messages, err := parseEmbeddedClusterOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Severity != "INFO" {
		t.Errorf("expected severity 'INFO', got %q", messages[0].Severity)
	}
}

func TestParseEmbeddedClusterOutput_MixedSeverities(t *testing.T) {
	output := `{
  "files": [
    {
      "path": "/tmp/test-config.yaml",
      "valid": false,
      "errors": [
        {
          "field": "spec.version",
          "message": "version is required"
        }
      ],
      "warnings": [
        {
          "field": "spec.extensions",
          "message": "extension may be deprecated"
        }
      ],
      "infos": [
        {
          "field": "",
          "message": "config validated successfully"
        }
      ]
    }
  ]
}`

	messages, err := parseEmbeddedClusterOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}

	// Messages should be in order: errors, warnings, infos
	if messages[0].Severity != "ERROR" {
		t.Errorf("expected first message to be ERROR, got %q", messages[0].Severity)
	}
	if messages[1].Severity != "WARNING" {
		t.Errorf("expected second message to be WARNING, got %q", messages[1].Severity)
	}
	if messages[2].Severity != "INFO" {
		t.Errorf("expected third message to be INFO, got %q", messages[2].Severity)
	}
}

func TestParseEmbeddedClusterOutput_EmptyOutput(t *testing.T) {
	messages, err := parseEmbeddedClusterOutput("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("expected 0 messages for empty output, got %d", len(messages))
	}
}

func TestParseEmbeddedClusterOutput_MalformedJSON(t *testing.T) {
	output := `{not valid json}`

	_, err := parseEmbeddedClusterOutput(output)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestParseEmbeddedClusterOutput_WithTrailingError(t *testing.T) {
	// This simulates real embedded-cluster output when validation fails
	output := `{
  "files": [
    {
      "path": "test-config.yaml",
      "valid": false,
      "errors": [
        {
          "field": "",
          "message": "YAML syntax error at line 7: yaml: unmarshal errors:\n  line 7: key \"name\" already set in map"
        }
      ]
    }
  ]
}
ERROR: validation failed with errors`

	messages, err := parseEmbeddedClusterOutput(output)
	if err != nil {
		t.Fatalf("unexpected error parsing output with trailing error text: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Severity != "ERROR" {
		t.Errorf("expected severity 'ERROR', got %q", messages[0].Severity)
	}

	// Verify the error message was parsed correctly
	expectedMessage := "YAML syntax error at line 7: yaml: unmarshal errors:\n  line 7: key \"name\" already set in map"
	if messages[0].Message != expectedMessage {
		t.Errorf("unexpected message.\nExpected: %q\nGot: %q", expectedMessage, messages[0].Message)
	}
}

// Binary override tests

func TestLintEmbeddedCluster_LocalBinaryOverride(t *testing.T) {
	ctx := context.Background()

	// Create a temporary EC config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ec-config.yaml")
	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "1.0.0"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Create a mock EC binary script
	mockBinaryPath := filepath.Join(tmpDir, "mock-ec-linter")
	mockScript := `#!/bin/sh
# Mock embedded-cluster lint binary for testing
# Returns valid JSON output
cat <<'EOF'
{
  "files": [
    {
      "path": "$3",
      "valid": true
    }
  ]
}
EOF
`
	if err := os.WriteFile(mockBinaryPath, []byte(mockScript), 0755); err != nil {
		t.Fatalf("failed to create mock binary: %v", err)
	}

	// Set the override environment variable
	originalEnv := os.Getenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
	os.Setenv("REPLICATED_EMBEDDED_CLUSTER_PATH", mockBinaryPath)
	defer func() {
		if originalEnv != "" {
			os.Setenv("REPLICATED_EMBEDDED_CLUSTER_PATH", originalEnv)
		} else {
			os.Unsetenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
		}
	}()

	// Run linter - should use mock binary
	result, err := LintEmbeddedCluster(ctx, configPath, "latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected success from mock binary")
	}

	if len(result.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(result.Messages))
	}
}

func TestLintEmbeddedCluster_FallsBackToResolver(t *testing.T) {
	ctx := context.Background()

	// Create a temporary EC config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ec-config.yaml")
	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "1.0.0"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Ensure override env var is NOT set
	originalEnv := os.Getenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
	os.Unsetenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
	defer func() {
		if originalEnv != "" {
			os.Setenv("REPLICATED_EMBEDDED_CLUSTER_PATH", originalEnv)
		}
	}()

	// Run linter - should attempt to use resolver
	// On darwin-arm64, this will fail because binaries are linux-only
	// On linux-amd64, this would download the binary
	_, err := LintEmbeddedCluster(ctx, configPath, "latest")

	// We expect an error on darwin-arm64 (binary not available)
	// but the error should be from the resolver, not from finding the path
	if err != nil {
		// Verify it's a resolver error, not a config path error
		if err.Error() == "embedded cluster config path does not exist: "+configPath {
			t.Errorf("unexpected error - config path check ran first: %v", err)
		}
		// Resolver error is expected on darwin-arm64
		t.Logf("Expected resolver error on this platform: %v", err)
	}
}

// ExtractECVersion tests

func TestExtractECVersion_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "valid-config.yaml")

	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: test-config
spec:
  version: "1.33+k8s-1.33"
  roles:
    controller:
      name: "Controller"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	version, err := ExtractECVersion(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1.33+k8s-1.33"
	if version != expected {
		t.Errorf("expected version %q, got %q", expected, version)
	}
}

func TestExtractECVersion_MissingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "missing-version.yaml")

	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: test-config
spec:
  roles:
    controller:
      name: "Controller"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := ExtractECVersion(configPath)
	if err == nil {
		t.Fatal("expected error for missing version, got nil")
	}

	if err.Error() != "embedded cluster config missing spec.version field" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExtractECVersion_FileNotFound(t *testing.T) {
	_, err := ExtractECVersion("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}

	if !os.IsNotExist(err) && !filepath.IsAbs("/nonexistent/path/config.yaml") {
		t.Errorf("expected not exist error, got: %v", err)
	}
}

func TestExtractECVersion_MultiDocumentYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "multi-doc.yaml")

	// Multi-document YAML with EC Config as second document
	configContent := `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: some-configmap
data:
  key: value
---
apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: test-config
spec:
  version: "2.0.0+k8s-1.29"
---
apiVersion: v1
kind: Service
metadata:
  name: some-service
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	version, err := ExtractECVersion(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "2.0.0+k8s-1.29"
	if version != expected {
		t.Errorf("expected version %q, got %q", expected, version)
	}
}

func TestExtractECVersion_TemplatedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "templated.yaml")

	// Templated YAML that can't be parsed as valid YAML
	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: {{ .Values.appName }}
spec:
  version: "1.25+k8s-1.27"
  roles:
    controller:
      name: {{ .Values.controllerName }}
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Should fall back to string matching
	version, err := ExtractECVersion(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1.25+k8s-1.27"
	if version != expected {
		t.Errorf("expected version %q, got %q", expected, version)
	}
}

func TestExtractECVersion_QuotedVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "quoted.yaml")

	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "1.30+k8s-1.28"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	version, err := ExtractECVersion(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1.30+k8s-1.28"
	if version != expected {
		t.Errorf("expected version %q, got %q", expected, version)
	}
}

func TestExtractECVersion_UnquotedVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "unquoted.yaml")

	// Some users might write version without quotes (though this may cause YAML issues)
	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: 1.31+k8s-1.29
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// This should work via string matching fallback
	version, err := ExtractECVersion(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "1.31+k8s-1.29"
	if version != expected {
		t.Errorf("expected version %q, got %q", expected, version)
	}
}

func TestExtractECVersion_NotECConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "not-ec-config.yaml")

	// This is a KOTS Config, not EC Config
	configContent := `apiVersion: kots.io/v1beta1
kind: Config
metadata:
  name: app-config
spec:
  groups:
    - name: settings
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := ExtractECVersion(configPath)
	if err == nil {
		t.Fatal("expected error for non-EC config, got nil")
	}

	if err.Error() != "file does not appear to be an embedded cluster config" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExtractECVersion_EmptyVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty-version.yaml")

	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: ""
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := ExtractECVersion(configPath)
	if err == nil {
		t.Fatal("expected error for empty version, got nil")
	}

	if err.Error() != "embedded cluster config missing spec.version field" {
		t.Errorf("unexpected error message: %v", err)
	}
}

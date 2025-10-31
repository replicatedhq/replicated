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

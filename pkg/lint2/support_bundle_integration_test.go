//go:build integration
// +build integration

package lint2

import (
	"context"
	"testing"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// TestLintSupportBundle_Integration tests the full support bundle linting flow
// with actual support-bundle binary execution. This test requires the support-bundle
// tool to be downloadable and should be run with: go test -tags=integration
//
// NOTE: The support-bundle lint command may be in active development and could
// be temporarily broken. If these tests fail, it may be due to the command itself
// rather than the implementation. The infrastructure is ready for when the command
// is stabilized.
func TestLintSupportBundle_Integration(t *testing.T) {
	ctx := context.Background()

	t.Run("valid support bundle spec", func(t *testing.T) {
		result, err := LintSupportBundle(ctx, "testdata/support-bundles/valid.yaml", tools.DefaultSupportBundleVersion)
		if err != nil {
			t.Fatalf("LintSupportBundle() error = %v, want nil", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true for valid spec, got false")
		}

		// Valid spec may have warnings (e.g., missing docStrings)
		// but should not have errors
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR in valid spec: %s", msg.Message)
			}
		}
	})

	t.Run("invalid yaml support bundle spec", func(t *testing.T) {
		result, err := LintSupportBundle(ctx, "testdata/support-bundles/invalid-yaml.yaml", tools.DefaultSupportBundleVersion)
		if err != nil {
			t.Fatalf("LintSupportBundle() error = %v, want nil", err)
		}

		if result.Success {
			t.Errorf("Expected success=false for invalid YAML spec, got true")
		}

		// Should have at least one error message
		hasError := false
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				hasError = true
				// Verify error message is not empty
				if msg.Message == "" {
					t.Errorf("Error message should not be empty")
				}
			}
		}

		if !hasError {
			t.Errorf("Expected at least one ERROR message for invalid YAML spec")
		}
	})

	t.Run("missing collectors support bundle spec", func(t *testing.T) {
		result, err := LintSupportBundle(ctx, "testdata/support-bundles/missing-collectors.yaml", tools.DefaultSupportBundleVersion)
		if err != nil {
			t.Fatalf("LintSupportBundle() error = %v, want nil", err)
		}

		if result.Success {
			t.Errorf("Expected success=false for spec missing collectors, got true")
		}

		// Should have error about missing collectors
		hasCollectorError := false
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" && contains(msg.Message, "collector") {
				hasCollectorError = true
				break
			}
		}

		if !hasCollectorError {
			t.Errorf("Expected ERROR message about missing collectors")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := LintSupportBundle(ctx, "testdata/support-bundles/does-not-exist.yaml", tools.DefaultSupportBundleVersion)
		if err == nil {
			t.Errorf("Expected error for non-existent file, got nil")
		}

		// Error should mention the file doesn't exist
		if err != nil && !contains(err.Error(), "does not exist") {
			t.Errorf("Error should mention file doesn't exist, got: %v", err)
		}
	})
}

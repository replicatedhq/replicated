//go:build integration
// +build integration

package lint2

import (
	"context"
	"testing"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// TestLintChart_Integration tests the full helm chart linting flow
// with actual helm binary execution. This test requires the helm
// tool to be downloadable and should be run with: go test -tags=integration
func TestLintChart_Integration(t *testing.T) {
	ctx := context.Background()

	t.Run("valid helm chart", func(t *testing.T) {
		result, err := LintChart(ctx, "testdata/charts/valid-chart", tools.DefaultHelmVersion)
		if err != nil {
			t.Fatalf("LintChart() error = %v, want nil", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true for valid chart, got false")
		}

		// Valid chart may have INFO or WARNING messages
		// but should not have errors
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR in valid chart: %s", msg.Message)
			}
		}
	})

	t.Run("invalid yaml helm chart", func(t *testing.T) {
		result, err := LintChart(ctx, "testdata/charts/invalid-yaml", tools.DefaultHelmVersion)
		if err != nil {
			t.Fatalf("LintChart() error = %v, want nil", err)
		}

		if result.Success {
			t.Errorf("Expected success=false for invalid YAML chart, got true")
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
			t.Errorf("Expected at least one ERROR message for invalid YAML chart")
		}
	})

	t.Run("non-existent chart path", func(t *testing.T) {
		_, err := LintChart(ctx, "testdata/charts/does-not-exist", tools.DefaultHelmVersion)
		if err == nil {
			t.Errorf("Expected error for non-existent chart path, got nil")
		}

		// Error should mention the path doesn't exist
		if err != nil && !contains(err.Error(), "does not exist") {
			t.Errorf("Error should mention path doesn't exist, got: %v", err)
		}
	})
}

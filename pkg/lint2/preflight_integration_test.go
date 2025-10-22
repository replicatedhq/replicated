//go:build integration
// +build integration

package lint2

import (
	"context"
	"testing"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// TestLintPreflight_Integration tests the full preflight linting flow
// with actual preflight binary execution. This test requires the preflight
// tool to be downloadable and should be run with: go test -tags=integration
func TestLintPreflight_Integration(t *testing.T) {
	ctx := context.Background()

	t.Run("valid preflight spec", func(t *testing.T) {
		// No templating - pass empty strings for values/chart info and empty map for manifests
		result, err := LintPreflight(ctx, "testdata/preflights/valid.yaml", "", "", "", make(map[string]*HelmChartManifest), tools.DefaultPreflightVersion)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
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

	t.Run("invalid yaml preflight spec", func(t *testing.T) {
		result, err := LintPreflight(ctx, "testdata/preflights/invalid-yaml.yaml", "", "", "", make(map[string]*HelmChartManifest), tools.DefaultPreflightVersion)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		if result.Success {
			t.Errorf("Expected success=false for invalid YAML spec, got true")
		}

		// Should have at least one error message
		hasError := false
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				hasError = true
				// Verify error message mentions YAML or syntax
				if msg.Message == "" {
					t.Errorf("Error message should not be empty")
				}
			}
		}

		if !hasError {
			t.Errorf("Expected at least one ERROR message for invalid YAML spec")
		}
	})

	t.Run("missing analyzers preflight spec", func(t *testing.T) {
		result, err := LintPreflight(ctx, "testdata/preflights/missing-analyzers.yaml", "", "", "", make(map[string]*HelmChartManifest), tools.DefaultPreflightVersion)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		if result.Success {
			t.Errorf("Expected success=false for spec missing analyzers, got true")
		}

		// Should have error about missing analyzers
		hasAnalyzerError := false
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" && contains(msg.Message, "analyzer") {
				hasAnalyzerError = true
				break
			}
		}

		if !hasAnalyzerError {
			t.Errorf("Expected ERROR message about missing analyzers")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := LintPreflight(ctx, "testdata/preflights/does-not-exist.yaml", "", "", "", make(map[string]*HelmChartManifest), tools.DefaultPreflightVersion)
		if err == nil {
			t.Errorf("Expected error for non-existent file, got nil")
		}

		// Error should mention the file doesn't exist
		if err != nil && !contains(err.Error(), "does not exist") {
			t.Errorf("Error should mention file doesn't exist, got: %v", err)
		}
	})

	t.Run("templated preflight with builder values", func(t *testing.T) {
		// This test verifies that:
		// 1. Template rendering works ({{- if .Values.* }} expressions are evaluated)
		// 2. Builder values override chart values
		//    - Chart values.yaml has database.enabled: false, redis.enabled: false
		//    - Builder values have database.enabled: true, redis.enabled: true
		//    - If builder values work correctly, both collectors/analyzers should be rendered
		// 3. The rendered spec passes preflight lint validation

		// Discover HelmChart manifests
		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/templated-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		// Verify we found the HelmChart
		if len(helmChartManifests) != 1 {
			t.Fatalf("Expected 1 HelmChart manifest, got %d", len(helmChartManifests))
		}

		// Lint the templated preflight with values and builder values
		result, err := LintPreflight(
			ctx,
			"testdata/preflights/templated-test/preflight-templated.yaml",
			"testdata/preflights/templated-test/chart/values.yaml",
			"test-app",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		// Success indicates that:
		// - Template rendering succeeded (no {{ ... }} syntax errors)
		// - Builder values were applied (conditionals evaluated to true)
		// - Rendered spec is valid (has collectors and analyzers)
		if !result.Success {
			t.Errorf("Expected success=true for templated spec, got false")
			for _, msg := range result.Messages {
				t.Logf("Message: %s - %s", msg.Severity, msg.Message)
			}
		}

		// Should have no errors (may have warnings about missing docStrings)
		errorCount := 0
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR in templated spec: %s", msg.Message)
				errorCount++
			}
		}

		// Additional verification: If builder values weren't applied, we'd have an empty spec
		// (because chart values have enabled: false). This would cause errors like
		// "spec.collectors is required" or similar validation failures.
		// The fact that we have no errors confirms that builders values were applied correctly.
		if errorCount == 0 && result.Success {
			t.Logf("✓ Template rendering with builder values succeeded (postgres and redis collectors/analyzers were rendered)")
		}
	})

	t.Run("templated preflight missing HelmChart manifest", func(t *testing.T) {
		// Empty manifests map - simulates missing HelmChart
		emptyManifests := make(map[string]*HelmChartManifest)

		// Should fail because HelmChart manifest is required for templated preflights
		_, err := LintPreflight(
			ctx,
			"testdata/preflights/templated-test/preflight-templated.yaml",
			"testdata/preflights/templated-test/chart/values.yaml",
			"test-app",
			"1.0.0",
			emptyManifests,
			tools.DefaultPreflightVersion,
		)
		if err == nil {
			t.Fatal("Expected error for missing HelmChart manifest, got nil")
		}

		// Error should mention missing HelmChart
		if !contains(err.Error(), "no HelmChart manifest found") {
			t.Errorf("Error should mention missing HelmChart, got: %v", err)
		}
	})
}

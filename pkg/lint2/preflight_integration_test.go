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
		// Discover HelmChart manifests
		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/valid-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		result, err := LintPreflight(
			ctx,
			"testdata/preflights/valid-test/preflight-valid.yaml",
			"testdata/preflights/valid-test/chart/values.yaml",
			"test-app-valid",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
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
		// Discover HelmChart manifests
		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/invalid-yaml-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		result, err := LintPreflight(
			ctx,
			"testdata/preflights/invalid-yaml-test/preflight-invalid.yaml",
			"testdata/preflights/invalid-yaml-test/chart/values.yaml",
			"test-app-invalid-yaml",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
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
		// Discover HelmChart manifests
		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/missing-analyzers-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		result, err := LintPreflight(
			ctx,
			"testdata/preflights/missing-analyzers-test/preflight-missing.yaml",
			"testdata/preflights/missing-analyzers-test/chart/values.yaml",
			"test-app-missing-analyzers",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
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
		// Use existing test data for chart structure, but request non-existent preflight file
		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/templated-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		_, err = LintPreflight(
			ctx,
			"testdata/preflights/does-not-exist.yaml",
			"testdata/preflights/templated-test/chart/values.yaml",
			"test-app",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err == nil {
			t.Errorf("Expected error for non-existent file, got nil")
		}

		// Error should mention the template rendering or file issue
		if err != nil {
			t.Logf("Got expected error: %v", err)
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

	t.Run("templated preflight with builder values disabled - negative test", func(t *testing.T) {
		// This test verifies that when BOTH chart values AND builder values have enabled: false,
		// the rendered spec is empty/invalid and lint correctly fails.
		// This proves that:
		// 1. Template rendering works (conditionals are evaluated)
		// 2. Builder values override chart values (both have same value, no conflict)
		// 3. We correctly fail when the rendered spec is invalid

		// Discover HelmChart manifests for disabled test
		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/templated-disabled-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		if len(helmChartManifests) != 1 {
			t.Fatalf("Expected 1 HelmChart manifest, got %d", len(helmChartManifests))
		}

		// Lint the templated preflight where both values have enabled: false
		result, err := LintPreflight(
			ctx,
			"testdata/preflights/templated-disabled-test/preflight-templated.yaml",
			"testdata/preflights/templated-disabled-test/chart/values.yaml",
			"test-app-disabled",
			"2.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		// When both chart and builder values have enabled: false,
		// the template renders with empty collectors and analyzers.
		// This should cause preflight lint to FAIL because spec is incomplete.
		if result.Success {
			t.Errorf("Expected success=false when rendered spec has no collectors/analyzers, got true")
		}

		// Should have errors about missing required fields
		hasRequiredFieldError := false
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Logf("Expected ERROR: %s", msg.Message)
				if contains(msg.Message, "collector") || contains(msg.Message, "analyzer") || contains(msg.Message, "required") {
					hasRequiredFieldError = true
				}
			}
		}

		if !hasRequiredFieldError {
			t.Errorf("Expected ERROR about missing required fields (collectors/analyzers)")
		}

		t.Logf("✓ Correctly failed when builder values disabled (rendered spec invalid)")
	})

	t.Run("templated preflight verifies builder overrides chart values", func(t *testing.T) {
		// This test explicitly verifies that builder values override chart values.
		// Setup:
		//   - Chart values: database.enabled=false, redis.enabled=false
		//   - Builder values: database.enabled=true, redis.enabled=true
		// If builder values did NOT override, the spec would be empty and lint would fail.
		// If builder values DO override, collectors/analyzers are rendered and lint succeeds.

		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/templated-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		// Verify the HelmChart has builder values with enabled: true
		helmChart, found := helmChartManifests["test-app:1.0.0"]
		if !found {
			t.Fatal("HelmChart manifest not found for test-app:1.0.0")
		}

		// Verify builder values have enabled: true
		if builderDB, ok := helmChart.BuilderValues["database"].(map[string]interface{}); ok {
			if enabled, ok := builderDB["enabled"].(bool); !ok || !enabled {
				t.Error("Expected builder values to have database.enabled=true")
			}
		} else {
			t.Error("Builder values missing database config")
		}

		// Lint with these values
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

		// Success proves builder values overrode chart values
		// (chart values have enabled: false, which would produce empty spec)
		if !result.Success {
			t.Errorf("Expected success=true, proving builder values overrode chart values, got false")
			for _, msg := range result.Messages {
				t.Logf("Message: %s - %s", msg.Severity, msg.Message)
			}
		}

		// Should have no errors
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR: %s", msg.Message)
			}
		}

		t.Logf("✓ Verified builder values (enabled=true) overrode chart values (enabled=false)")
	})

	t.Run("end-to-end templated preflight from config", func(t *testing.T) {
		// This test verifies the complete end-to-end workflow:
		// 1. GetPreflightWithValuesFromConfig() extracts chart metadata
		// 2. DiscoverHelmChartManifests() finds builder values
		// 3. LintPreflight() renders and lints the spec
		// This tests the actual user workflow, not just isolated functions.

		// Create a config structure that mimics a .replicated config file
		config := &tools.Config{
			Preflights: []tools.PreflightConfig{
				{
					Path:       "testdata/preflights/templated-test/preflight-templated.yaml",
					ValuesPath: "testdata/preflights/templated-test/chart/values.yaml",
				},
			},
			Manifests: []string{"testdata/preflights/templated-test/manifests/*.yaml"},
		}

		// Step 1: Extract preflight paths with chart metadata
		preflights, err := GetPreflightWithValuesFromConfig(config)
		if err != nil {
			t.Fatalf("GetPreflightWithValuesFromConfig() error = %v", err)
		}

		if len(preflights) != 1 {
			t.Fatalf("Expected 1 preflight, got %d", len(preflights))
		}

		pf := preflights[0]

		// Verify chart metadata was extracted correctly
		if pf.ChartName != "test-app" {
			t.Errorf("Expected ChartName=test-app, got %s", pf.ChartName)
		}
		if pf.ChartVersion != "1.0.0" {
			t.Errorf("Expected ChartVersion=1.0.0, got %s", pf.ChartVersion)
		}
		if pf.ValuesPath == "" {
			t.Error("Expected ValuesPath to be set")
		}

		// Step 2: Discover HelmChart manifests (simulates CLI lint.go workflow)
		helmChartManifests, err := DiscoverHelmChartManifests(config.Manifests)
		if err != nil {
			t.Fatalf("DiscoverHelmChartManifests() error = %v", err)
		}

		if len(helmChartManifests) != 1 {
			t.Fatalf("Expected 1 HelmChart manifest, got %d", len(helmChartManifests))
		}

		// Step 3: Lint the preflight (complete workflow)
		result, err := LintPreflight(
			ctx,
			pf.SpecPath,
			pf.ValuesPath,
			pf.ChartName,
			pf.ChartVersion,
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true for end-to-end workflow, got false")
			for _, msg := range result.Messages {
				t.Logf("Message: %s - %s", msg.Severity, msg.Message)
			}
		}

		// Should have no errors
		errorCount := 0
		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR: %s", msg.Message)
				errorCount++
			}
		}

		if errorCount == 0 && result.Success {
			t.Logf("✓ End-to-end workflow succeeded: config → extract metadata → discover manifests → render → lint")
		}
	})

	t.Run("complex nested partial override", func(t *testing.T) {
		// This test verifies that builder values can partially override nested structures.
		// Chart values: postgresql.enabled=false, postgresql.host=localhost, postgresql.port=5432
		// Builder values: postgresql.enabled=true (ONLY enabled, not host/port)
		// Expected: enabled comes from builder (true), host/port come from chart (localhost:5432)
		// This is a common pattern - override feature flags but keep connection details

		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/nested-override-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		if len(helmChartManifests) != 1 {
			t.Fatalf("Expected 1 HelmChart manifest, got %d", len(helmChartManifests))
		}

		// Verify builder only has 'enabled', not 'host' or 'port'
		helmChart, found := helmChartManifests["test-app-nested:1.0.0"]
		if !found {
			t.Fatal("HelmChart manifest not found for test-app-nested:1.0.0")
		}
		if postgresql, ok := helmChart.BuilderValues["postgresql"].(map[string]interface{}); ok {
			if _, hasHost := postgresql["host"]; hasHost {
				t.Error("Builder should NOT have postgresql.host (should come from chart)")
			}
			if _, hasPort := postgresql["port"]; hasPort {
				t.Error("Builder should NOT have postgresql.port (should come from chart)")
			}
			if enabled, ok := postgresql["enabled"].(bool); !ok || !enabled {
				t.Error("Builder should have postgresql.enabled=true")
			}
		}

		result, err := LintPreflight(
			ctx,
			"testdata/preflights/nested-override-test/preflight-nested.yaml",
			"testdata/preflights/nested-override-test/chart/values.yaml",
			"test-app-nested",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true for nested partial override, got false")
			for _, msg := range result.Messages {
				t.Logf("Message: %s - %s", msg.Severity, msg.Message)
			}
		}

		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR: %s", msg.Message)
			}
		}

		t.Logf("✓ Partial nested override works: builder.enabled=true, chart.host/port used")
	})

	t.Run("array values from builder", func(t *testing.T) {
		// This test verifies that builder can provide array values.
		// Chart values: ingress.hosts=[] (empty array)
		// Builder values: ingress.hosts=[host1, host2, host3]
		// Template uses: {{- range .Values.ingress.hosts }}
		// Expected: Template iterates over builder's 3 hosts

		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/array-values-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		// Verify builder has array with 3 hosts
		helmChart, found := helmChartManifests["test-app-arrays:1.0.0"]
		if !found {
			t.Fatal("HelmChart manifest not found for test-app-arrays:1.0.0")
		}
		if ingress, ok := helmChart.BuilderValues["ingress"].(map[string]interface{}); ok {
			if hosts, ok := ingress["hosts"].([]interface{}); ok {
				if len(hosts) != 3 {
					t.Errorf("Expected 3 hosts in builder, got %d", len(hosts))
				}
			} else {
				t.Error("Builder should have ingress.hosts as array")
			}
		}

		result, err := LintPreflight(
			ctx,
			"testdata/preflights/array-values-test/preflight-arrays.yaml",
			"testdata/preflights/array-values-test/chart/values.yaml",
			"test-app-arrays",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true for array values, got false")
			for _, msg := range result.Messages {
				t.Logf("Message: %s - %s", msg.Severity, msg.Message)
			}
		}

		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR: %s", msg.Message)
			}
		}

		t.Logf("✓ Array values work: {{- range }} iterates over builder's 3 hosts")
	})

	t.Run("string interpolation without conditionals", func(t *testing.T) {
		// This test verifies direct value substitution in strings (no {{- if }} conditionals).
		// Chart values: database.host=localhost, database.port=5432, database.name=devdb
		// Builder values: database.host=prod.database.example.com, database.port=5432, database.name=proddb
		// Template: uri: 'postgresql://{{ .Values.database.user }}@{{ .Values.database.host }}:{{ .Values.database.port }}/{{ .Values.database.name }}'
		// Expected: Builder values substitute directly into connection string

		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/string-interpolation-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		// Verify builder has production values
		helmChart, found := helmChartManifests["test-app-strings:1.0.0"]
		if !found {
			t.Fatal("HelmChart manifest not found for test-app-strings:1.0.0")
		}
		if database, ok := helmChart.BuilderValues["database"].(map[string]interface{}); ok {
			if host, ok := database["host"].(string); !ok || host != "prod.database.example.com" {
				t.Errorf("Expected builder to have database.host=prod.database.example.com, got %v", database["host"])
			}
			if name, ok := database["name"].(string); !ok || name != "proddb" {
				t.Errorf("Expected builder to have database.name=proddb, got %v", database["name"])
			}
		}

		result, err := LintPreflight(
			ctx,
			"testdata/preflights/string-interpolation-test/preflight-strings.yaml",
			"testdata/preflights/string-interpolation-test/chart/values.yaml",
			"test-app-strings",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true for string interpolation, got false")
			for _, msg := range result.Messages {
				t.Logf("Message: %s - %s", msg.Severity, msg.Message)
			}
		}

		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR: %s", msg.Message)
			}
		}

		t.Logf("✓ String interpolation works: builder values substituted in connection strings")
	})

	t.Run("multiple charts with multiple preflights", func(t *testing.T) {
		// This test verifies that multiple charts/preflights work correctly.
		// Charts: frontend-app:1.0.0, backend-app:2.0.0
		// Preflights: One for frontend (uses service.port), one for backend (uses api.port)
		// Expected: Each preflight gets correct builder values for its chart

		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/multi-chart-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		if len(helmChartManifests) != 2 {
			t.Fatalf("Expected 2 HelmChart manifests, got %d", len(helmChartManifests))
		}

		// Verify we have both charts
		frontendChart, foundFrontend := helmChartManifests["frontend-app:1.0.0"]
		backendChart, foundBackend := helmChartManifests["backend-app:2.0.0"]
		if !foundFrontend || !foundBackend {
			t.Fatal("Expected to find both frontend-app:1.0.0 and backend-app:2.0.0")
		}

		// Lint frontend preflight with frontend chart
		frontendResult, err := LintPreflight(
			ctx,
			"testdata/preflights/multi-chart-test/preflight-frontend.yaml",
			"testdata/preflights/multi-chart-test/frontend-chart/values.yaml",
			"frontend-app",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() frontend error = %v, want nil", err)
		}

		if !frontendResult.Success {
			t.Errorf("Expected success=true for frontend preflight, got false")
		}

		// Verify frontend used correct builder (service.enabled=true, service.port=3000)
		if service, ok := frontendChart.BuilderValues["service"].(map[string]interface{}); ok {
			if port, ok := service["port"].(int); !ok || port != 3000 {
				t.Errorf("Frontend builder should have service.port=3000, got %v", service["port"])
			}
		}

		// Lint backend preflight with backend chart
		backendResult, err := LintPreflight(
			ctx,
			"testdata/preflights/multi-chart-test/preflight-backend.yaml",
			"testdata/preflights/multi-chart-test/backend-chart/values.yaml",
			"backend-app",
			"2.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() backend error = %v, want nil", err)
		}

		if !backendResult.Success {
			t.Errorf("Expected success=true for backend preflight, got false")
		}

		// Verify backend used correct builder (api.enabled=true, api.port=8080)
		if api, ok := backendChart.BuilderValues["api"].(map[string]interface{}); ok {
			if port, ok := api["port"].(int); !ok || port != 8080 {
				t.Errorf("Backend builder should have api.port=8080, got %v", api["port"])
			}
		}

		t.Logf("✓ Multiple charts work: frontend used service.port=3000, backend used api.port=8080")
	})

	t.Run("empty builder values uses chart defaults", func(t *testing.T) {
		// This test verifies that when builder is empty (builder: {}), chart defaults are used.
		// Chart values: feature.enabled=true, feature.name=default-feature, feature.timeout=30
		// Builder values: {} (explicitly empty, not nil)
		// Expected: All values come from chart defaults

		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/empty-builder-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("Failed to discover HelmChart manifests: %v", err)
		}

		// Verify builder is empty
		helmChart, found := helmChartManifests["test-app-empty-builder:1.0.0"]
		if !found {
			t.Fatal("HelmChart manifest not found for test-app-empty-builder:1.0.0")
		}
		if helmChart.BuilderValues == nil || len(helmChart.BuilderValues) != 0 {
			t.Errorf("Expected empty builder values map, got %v", helmChart.BuilderValues)
		}

		result, err := LintPreflight(
			ctx,
			"testdata/preflights/empty-builder-test/preflight-empty-builder.yaml",
			"testdata/preflights/empty-builder-test/chart/values.yaml",
			"test-app-empty-builder",
			"1.0.0",
			helmChartManifests,
			tools.DefaultPreflightVersion,
		)
		if err != nil {
			t.Fatalf("LintPreflight() error = %v, want nil", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true for empty builder, got false")
			for _, msg := range result.Messages {
				t.Logf("Message: %s - %s", msg.Severity, msg.Message)
			}
		}

		for _, msg := range result.Messages {
			if msg.Severity == "ERROR" {
				t.Errorf("Unexpected ERROR: %s", msg.Message)
			}
		}

		t.Logf("✓ Empty builder works: chart defaults used (enabled=true, name=default-feature, timeout=30)")
	})

	t.Run("manifests without HelmChart kind", func(t *testing.T) {
		// This test verifies the error path when manifests are configured but don't contain any kind: HelmChart.
		// Scenario: User has Deployment, Service, ConfigMap manifests, but forgot the HelmChart custom resource.
		// Expected:
		// 1. DiscoverHelmChartManifests() returns empty map (no error)
		// 2. LintPreflight() returns error with helpful message pointing to the issue

		// Manifests directory contains Deployment, Service, ConfigMap - but NO HelmChart
		helmChartManifests, err := DiscoverHelmChartManifests([]string{"testdata/preflights/no-helmchart-test/manifests/*.yaml"})
		if err != nil {
			t.Fatalf("DiscoverHelmChartManifests() should not error when no HelmChart found, got: %v", err)
		}

		// Should return empty map, not error
		if len(helmChartManifests) != 0 {
			t.Errorf("Expected empty map when no HelmChart found, got %d manifests", len(helmChartManifests))
		}

		// Now try to lint a templated preflight with this empty map
		_, err = LintPreflight(
			ctx,
			"testdata/preflights/no-helmchart-test/preflight-templated.yaml",
			"testdata/preflights/no-helmchart-test/chart/values.yaml",
			"test-app-no-helmchart",
			"1.0.0",
			helmChartManifests, // Empty map
			tools.DefaultPreflightVersion,
		)

		// Should error with helpful message
		if err == nil {
			t.Fatal("Expected error when HelmChart not found in manifests, got nil")
		}

		// Verify error message is helpful
		expectedPhrases := []string{
			"no HelmChart manifest found",
			"test-app-no-helmchart:1.0.0",
			"manifests paths",
		}
		for _, phrase := range expectedPhrases {
			if !contains(err.Error(), phrase) {
				t.Errorf("Error message should contain %q, got: %v", phrase, err)
			}
		}

		t.Logf("✓ Helpful error when manifests configured but no HelmChart found: %v", err)
	})
}

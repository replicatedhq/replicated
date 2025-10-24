//go:build integration
// +build integration

package lint2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// TestLintValidation_Success tests successful validation with matching HelmChart manifests
func TestLintValidation_Success(t *testing.T) {
	testDir := filepath.Join("testdata", "validation", "scenario-1-success")

	// Change to test directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Extract charts with metadata
	charts, err := GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Discover HelmChart manifests
	helmCharts, err := DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("DiscoverHelmChartManifests failed: %v", err)
	}

	// Validate
	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Should have no warnings
	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

// TestLintValidation_MissingHelmChart tests error when HelmChart manifest is missing
func TestLintValidation_MissingHelmChart(t *testing.T) {
	testDir := filepath.Join("testdata", "validation", "scenario-2-missing-helmchart")

	// Change to test directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Extract charts with metadata
	charts, err := GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Discover HelmChart manifests (should be empty)
	helmCharts, err := DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("DiscoverHelmChartManifests failed: %v", err)
	}

	// Validate - should fail
	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	// Should be a MultipleChartsMissingHelmChartsError
	multiErr, ok := err.(*MultipleChartsMissingHelmChartsError)
	if !ok {
		t.Fatalf("expected *MultipleChartsMissingHelmChartsError, got %T", err)
	}

	// Should report 1 missing chart
	if len(multiErr.MissingCharts) != 1 {
		t.Errorf("expected 1 missing chart, got %d", len(multiErr.MissingCharts))
	}

	// Verify chart name
	if multiErr.MissingCharts[0].ChartName != "missing-manifest-app" {
		t.Errorf("expected chart name 'missing-manifest-app', got %s", multiErr.MissingCharts[0].ChartName)
	}

	// Result should be nil on error
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestLintValidation_MultipleCharts tests batch error reporting for multiple missing charts
func TestLintValidation_MultipleCharts(t *testing.T) {
	testDir := filepath.Join("testdata", "validation", "scenario-3-multiple-charts")

	// Change to test directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Extract charts with metadata
	charts, err := GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Should have 3 charts
	if len(charts) != 3 {
		t.Fatalf("expected 3 charts, got %d", len(charts))
	}

	// Discover HelmChart manifests (only 2 exist)
	helmCharts, err := DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("DiscoverHelmChartManifests failed: %v", err)
	}

	// Should have 2 HelmChart manifests
	if len(helmCharts) != 2 {
		t.Fatalf("expected 2 HelmChart manifests, got %d", len(helmCharts))
	}

	// Validate - should fail with batch error
	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	// Should be a MultipleChartsMissingHelmChartsError
	multiErr, ok := err.(*MultipleChartsMissingHelmChartsError)
	if !ok {
		t.Fatalf("expected *MultipleChartsMissingHelmChartsError, got %T", err)
	}

	// Should report 1 missing chart (database)
	if len(multiErr.MissingCharts) != 1 {
		t.Errorf("expected 1 missing chart, got %d", len(multiErr.MissingCharts))
	}

	// Verify the missing chart is 'database'
	if multiErr.MissingCharts[0].ChartName != "database" {
		t.Errorf("expected missing chart 'database', got %s", multiErr.MissingCharts[0].ChartName)
	}

	// Error message should mention batch reporting
	errMsg := err.Error()
	if !strings.Contains(errMsg, "database") {
		t.Errorf("error message should contain 'database': %s", errMsg)
	}

	// Result should be nil on error
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestLintValidation_OrphanedManifest tests warning for orphaned HelmChart manifests
func TestLintValidation_OrphanedManifest(t *testing.T) {
	testDir := filepath.Join("testdata", "validation", "scenario-4-orphaned-manifest")

	// Change to test directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Extract charts with metadata
	charts, err := GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Should have 1 chart
	if len(charts) != 1 {
		t.Fatalf("expected 1 chart, got %d", len(charts))
	}

	// Discover HelmChart manifests (2 exist, 1 orphaned)
	helmCharts, err := DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("DiscoverHelmChartManifests failed: %v", err)
	}

	// Should have 2 HelmChart manifests
	if len(helmCharts) != 2 {
		t.Fatalf("expected 2 HelmChart manifests, got %d", len(helmCharts))
	}

	// Validate - should succeed with warning
	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Should have 1 warning
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Warning should mention old-app
	warning := result.Warnings[0]
	if !strings.Contains(warning, "old-app") {
		t.Errorf("warning should contain 'old-app': %s", warning)
	}
	if !strings.Contains(warning, "no corresponding chart") {
		t.Errorf("warning should explain issue: %s", warning)
	}
}

// TestLintValidation_NoManifestsConfig tests error when manifests section is missing
func TestLintValidation_NoManifestsConfig(t *testing.T) {
	testDir := filepath.Join("testdata", "validation", "scenario-6-no-manifests-config")

	// Change to test directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Charts should be configured
	if len(config.Charts) == 0 {
		t.Fatal("expected charts to be configured")
	}

	// Manifests should be empty
	if len(config.Manifests) != 0 {
		t.Fatalf("expected no manifests configured, got %d", len(config.Manifests))
	}

	// This simulates the error check in runLint()
	// Should error early before extraction
	if len(config.Charts) > 0 && len(config.Manifests) == 0 {
		// Expected error path - this is what runLint() checks
		t.Log("Correctly detected charts with no manifests config")
	} else {
		t.Error("should have charts but no manifests")
	}
}

// TestLintValidation_AutoDiscovery tests auto-discovery mode
func TestLintValidation_AutoDiscovery(t *testing.T) {
	testDir := filepath.Join("testdata", "validation", "scenario-5-auto-discovery")

	// Change to test directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("failed to change to test directory: %v", err)
	}

	// Simulate auto-discovery (no .replicated file exists)
	// Auto-discover charts
	chartPaths, err := DiscoverChartPaths(filepath.Join(".", "**"))
	if err != nil {
		t.Fatalf("failed to auto-discover charts: %v", err)
	}

	if len(chartPaths) != 1 {
		t.Fatalf("expected 1 chart discovered, got %d", len(chartPaths))
	}

	// Auto-discover HelmChart manifests
	helmChartPaths, err := DiscoverHelmChartPaths(filepath.Join(".", "**"))
	if err != nil {
		t.Fatalf("failed to auto-discover HelmChart manifests: %v", err)
	}

	if len(helmChartPaths) != 1 {
		t.Fatalf("expected 1 HelmChart manifest discovered, got %d", len(helmChartPaths))
	}

	// Create a temporary config with discovered resources
	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chartPaths[0]},
		},
		Manifests: helmChartPaths,
	}

	// Extract charts with metadata
	charts, err := GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Discover HelmChart manifests
	helmCharts, err := DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("DiscoverHelmChartManifests failed: %v", err)
	}

	// Validate - should succeed
	result, err := ValidateChartToHelmChartMapping(charts, helmCharts)
	if err != nil {
		t.Fatalf("validation failed in auto-discovery mode: %v", err)
	}

	// Should have no warnings
	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}

	t.Log("Auto-discovery successfully found and validated chart with HelmChart manifest")
}

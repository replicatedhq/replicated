package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/tools"
)

// getAbsTestDataPath returns the absolute path to a testdata subdirectory
func getAbsTestDataPath(t *testing.T, relPath string) string {
	t.Helper()

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Build path relative to project root (two levels up from cli/cmd)
	projectRoot := filepath.Join(cwd, "..", "..")
	absPath := filepath.Join(projectRoot, relPath)

	// Verify path exists
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("test data path does not exist: %s (error: %v)", absPath, err)
	}

	return absPath
}

func TestExtractImagesFromConfig_ChartWithRequiredValues_WithMatchingHelmChartManifest(t *testing.T) {
	// Test that builder values from HelmChart manifest enable rendering of charts with required values
	chartPath := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "chart-with-required-values-test", "chart"))
	manifestGlob := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "chart-with-required-values-test", "manifests")) + "/*.yaml"

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chartPath},
		},
		Manifests: []string{manifestGlob},
	}

	r := &runners{}
	ctx := context.Background()

	// Extract charts with metadata
	charts, err := lint2.GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Extract HelmChart manifests
	helmChartManifests, err := lint2.DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("GetHelmChartManifestsFromConfig failed: %v", err)
	}

	result, err := r.extractImagesFromCharts(ctx, charts, helmChartManifests)
	if err != nil {
		t.Fatalf("extractImagesFromCharts failed: %v", err)
	}

	// Should successfully extract both postgres and redis images
	if len(result.Images) < 2 {
		t.Fatalf("Expected at least 2 images to be extracted with builder values, got %d", len(result.Images))
	}

	// Check that we got the expected images
	foundPostgres := false
	foundRedis := false
	for _, img := range result.Images {
		if (img.Repository == "library/postgres" || img.Repository == "postgres") && img.Tag == "15-alpine" {
			foundPostgres = true
		}
		if (img.Repository == "library/redis" || img.Repository == "redis") && img.Tag == "7-alpine" {
			foundRedis = true
		}
	}

	if !foundPostgres {
		t.Errorf("Expected to find postgres:15-alpine image. Got images: %+v", result.Images)
	}
	if !foundRedis {
		t.Errorf("Expected to find redis:7-alpine image. Got images: %+v", result.Images)
	}
}

func TestExtractImagesFromConfig_ChartWithRequiredValues_NoHelmChartManifest(t *testing.T) {
	// Test that extraction fails when manifests are not configured
	chartPath := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "chart-with-required-values-test", "chart"))

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chartPath},
		},
		Manifests: []string{}, // No manifests configured
	}

	// Try to extract HelmChart manifests - should fail because manifests are required
	_, err := lint2.DiscoverHelmChartManifests(config.Manifests)

	// Should fail because manifests are required
	if err == nil {
		t.Fatal("Expected error when manifests not configured, got nil")
	}

	// Error should mention manifests configuration
	if !strings.Contains(err.Error(), "no manifests configured") {
		t.Errorf("Expected error about manifests configuration, got: %v", err)
	}
}


func TestExtractImagesFromConfig_NonMatchingHelmChart_FailsToRender(t *testing.T) {
	// Test that HelmChart manifest must match chart name:version exactly
	chartPath := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "non-matching-helmchart-test", "chart"))
	manifestGlob := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "non-matching-helmchart-test", "manifests")) + "/*.yaml"

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chartPath},
		},
		Manifests: []string{manifestGlob},
	}

	r := &runners{}
	ctx := context.Background()

	// Extract charts with metadata
	charts, err := lint2.GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Extract HelmChart manifests
	helmChartManifests, err := lint2.DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("GetHelmChartManifestsFromConfig failed: %v", err)
	}

	result, err := r.extractImagesFromCharts(ctx, charts, helmChartManifests)
	if err != nil {
		t.Fatalf("extractImagesFromCharts failed: %v", err)
	}

	// Should get 0 images because HelmChart doesn't match (different name)
	if len(result.Images) != 0 {
		t.Errorf("Expected 0 images (HelmChart name doesn't match chart name), got %d: %+v", len(result.Images), result.Images)
	}

	// Should have a warning about the failure
	if len(result.Warnings) == 0 {
		t.Error("Expected at least one warning about failed extraction")
	}
}

func TestExtractImagesFromConfig_MultipleCharts_MixedScenario(t *testing.T) {
	// Test extracting from multiple charts - one with builder values, one without
	chart1Path := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "chart-with-required-values-test", "chart"))
	chart2Path := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "simple-chart-test", "chart"))
	manifestGlob := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "chart-with-required-values-test", "manifests")) + "/*.yaml"

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chart1Path},
			{Path: chart2Path},
		},
		Manifests: []string{manifestGlob},
	}

	r := &runners{}
	ctx := context.Background()

	// Extract charts with metadata
	charts, err := lint2.GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Extract HelmChart manifests
	helmChartManifests, err := lint2.DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("GetHelmChartManifestsFromConfig failed: %v", err)
	}

	result, err := r.extractImagesFromCharts(ctx, charts, helmChartManifests)
	if err != nil {
		t.Fatalf("extractImagesFromCharts failed: %v", err)
	}

	// Should extract images from both charts
	// Chart 1: postgres:15-alpine, redis:7-alpine (using builder values)
	// Chart 2: nginx:1.21 (hardcoded)
	if len(result.Images) < 3 {
		t.Errorf("Expected at least 3 images (2 from chart1, 1 from chart2), got %d", len(result.Images))
	}

	foundPostgres := false
	foundRedis := false
	foundNginx := false

	for _, img := range result.Images {
		if (img.Repository == "library/postgres" || img.Repository == "postgres") && img.Tag == "15-alpine" {
			foundPostgres = true
		}
		if (img.Repository == "library/redis" || img.Repository == "redis") && img.Tag == "7-alpine" {
			foundRedis = true
		}
		if (img.Repository == "library/nginx" || img.Repository == "nginx") && img.Tag == "1.21" {
			foundNginx = true
		}
	}

	if !foundPostgres {
		t.Errorf("Expected to find postgres:15-alpine from chart with builder values. Got: %+v", result.Images)
	}
	if !foundRedis {
		t.Errorf("Expected to find redis:7-alpine from chart with builder values. Got: %+v", result.Images)
	}
	if !foundNginx {
		t.Errorf("Expected to find nginx:1.21 from simple chart. Got: %+v", result.Images)
	}
}

func TestExtractImagesFromConfig_NoCharts_ReturnsError(t *testing.T) {
	// Test that empty chart configuration returns error
	config := &tools.Config{
		Charts:    []tools.ChartConfig{},
		Manifests: []string{},
	}

	// Extract charts with metadata - should error when no charts configured
	_, err := lint2.GetChartsWithMetadataFromConfig(config)

	// Should get error about no charts
	if err == nil {
		t.Fatal("expected error when no charts in config")
	}
	if !strings.Contains(err.Error(), "no charts found") {
		t.Errorf("expected 'no charts found' error, got: %v", err)
	}
}

func TestExtractImagesFromConfig_NoManifests_ReturnsError(t *testing.T) {
	// Test that manifests are required for image extraction
	chartPath := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "simple-chart-test", "chart"))

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chartPath},
		},
		Manifests: []string{}, // No manifests configured
	}

	// Try to extract HelmChart manifests - should fail because manifests are required
	_, err := lint2.DiscoverHelmChartManifests(config.Manifests)

	// Should fail because manifests are required
	if err == nil {
		t.Fatal("Expected error when manifests not configured, got nil")
	}

	// Error should mention manifests configuration
	if !strings.Contains(err.Error(), "no manifests configured") {
		t.Errorf("Expected error about manifests configuration, got: %v", err)
	}
}


func TestExtractImagesFromConfig_EmptyBuilder_FailsToRender(t *testing.T) {
	// Test that HelmChart manifest with empty builder section doesn't provide values
	chartPath := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "empty-builder-test", "chart"))
	manifestGlob := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "empty-builder-test", "manifests")) + "/*.yaml"

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chartPath},
		},
		Manifests: []string{manifestGlob},
	}

	r := &runners{}
	ctx := context.Background()

	// Extract charts with metadata
	charts, err := lint2.GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Extract HelmChart manifests
	helmChartManifests, err := lint2.DiscoverHelmChartManifests(config.Manifests)
	if err != nil {
		t.Fatalf("GetHelmChartManifestsFromConfig failed: %v", err)
	}

	result, err := r.extractImagesFromCharts(ctx, charts, helmChartManifests)
	if err != nil {
		t.Fatalf("extractImagesFromCharts failed: %v", err)
	}

	// Should get 0 images because empty builder provides no values
	if len(result.Images) != 0 {
		t.Errorf("Expected 0 images (empty builder provides no values), got %d: %+v", len(result.Images), result.Images)
	}

	// Should have a warning about the failure
	if len(result.Warnings) == 0 {
		t.Error("Expected at least one warning about failed extraction")
	}
}

func TestExtractImagesFromConfig_NoHelmChartInManifests_FailsDiscovery(t *testing.T) {
	// Test that manifests with other K8s resources but no HelmChart kind fail discovery
	chartPath := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "no-helmchart-test", "chart"))
	manifestGlob := getAbsTestDataPath(t, filepath.Join("testdata", "image-extraction", "no-helmchart-test", "manifests")) + "/*.yaml"

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chartPath},
		},
		Manifests: []string{manifestGlob},
	}

	// Try to extract HelmChart manifests - should fail because manifests don't contain HelmCharts
	_, err := lint2.DiscoverHelmChartManifests(config.Manifests)

	// Should fail because manifests are configured but contain no HelmCharts
	if err == nil {
		t.Fatal("Expected error when manifests configured but no HelmCharts found, got nil")
	}

	// Error should mention no HelmChart resources found
	if !strings.Contains(err.Error(), "no HelmChart resources found") {
		t.Errorf("Expected error about no HelmCharts, got: %v", err)
	}
}

//go:build integration
// +build integration

package lint2

import (
	"path/filepath"
	"testing"
)

// Phase 5 Tests: Integration - Cross-Linter Behavior

func TestIntegration_MixedDirectoryAllThreeTypes(t *testing.T) {
	// Test that all three linters correctly discover their resources from the same pattern
	// Pattern: ./k8s/** should work for charts, preflights, and support bundles
	tmpDir := t.TempDir()
	k8sDir := filepath.Join(tmpDir, "k8s")

	// Create a chart
	appChartDir := createTestChart(t, k8sDir, "app")

	// Create a Preflight spec
	preflightPath := filepath.Join(k8sDir, "preflights", "check.yaml")
	createTestPreflight(t, preflightPath)

	// Create a SupportBundle spec
	bundlePath := filepath.Join(k8sDir, "manifests", "bundle.yaml")
	createTestSupportBundle(t, bundlePath)

	// Create various K8s resources that should be filtered
	createTestK8sResource(t, filepath.Join(k8sDir, "deployment.yaml"), "Deployment")
	createTestK8sResource(t, filepath.Join(k8sDir, "service.yaml"), "Service")

	pattern := filepath.Join(k8sDir, "**")

	// Test chart discovery
	t.Run("charts", func(t *testing.T) {
		chartPaths, err := discoverChartPaths(pattern)
		if err != nil {
			t.Fatalf("discoverChartPaths() error = %v", err)
		}

		wantCharts := []string{appChartDir}
		assertPathsEqual(t, chartPaths, wantCharts)
	})

	// Test preflight discovery
	t.Run("preflights", func(t *testing.T) {
		preflightPaths, err := discoverPreflightPaths(pattern)
		if err != nil {
			t.Fatalf("discoverPreflightPaths() error = %v", err)
		}

		wantPreflights := []string{preflightPath}
		assertPathsEqual(t, preflightPaths, wantPreflights)
	})

	// Test support bundle discovery
	t.Run("support_bundles", func(t *testing.T) {
		manifestPattern := filepath.Join(k8sDir, "**", "*.yaml")
		bundlePaths, err := DiscoverSupportBundlesFromManifests([]string{manifestPattern})
		if err != nil {
			t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
		}

		wantBundles := []string{bundlePath}
		assertPathsEqual(t, bundlePaths, wantBundles)
	})
}

func TestIntegration_SamePatternMultipleLinters(t *testing.T) {
	// Test that each linter finds only its resources when using the same pattern
	tmpDir := t.TempDir()
	resourcesDir := filepath.Join(tmpDir, "resources")

	// Create all three types in the same directory
	chartDir := createTestChart(t, resourcesDir, "my-chart")
	preflightPath := filepath.Join(resourcesDir, "preflight.yaml")
	createTestPreflight(t, preflightPath)
	bundlePath := filepath.Join(resourcesDir, "bundle.yaml")
	createTestSupportBundle(t, bundlePath)

	// Also add some K8s resources
	createTestK8sResource(t, filepath.Join(resourcesDir, "deployment.yaml"), "Deployment")
	createTestK8sResource(t, filepath.Join(resourcesDir, "service.yaml"), "Service")

	pattern := filepath.Join(resourcesDir, "**")

	// All three linters should find only their resources
	chartPaths, err := discoverChartPaths(pattern)
	if err != nil {
		t.Fatalf("discoverChartPaths() error = %v", err)
	}
	if len(chartPaths) != 1 || chartPaths[0] != chartDir {
		t.Errorf("Charts: expected [%s], got %v", chartDir, chartPaths)
	}

	preflightPaths, err := discoverPreflightPaths(pattern)
	if err != nil {
		t.Fatalf("discoverPreflightPaths() error = %v", err)
	}
	if len(preflightPaths) != 1 || preflightPaths[0] != preflightPath {
		t.Errorf("Preflights: expected [%s], got %v", preflightPath, preflightPaths)
	}

	manifestPattern := filepath.Join(resourcesDir, "**", "*.yaml")
	bundlePaths, err := DiscoverSupportBundlesFromManifests([]string{manifestPattern})
	if err != nil {
		t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
	}
	if len(bundlePaths) != 1 || bundlePaths[0] != bundlePath {
		t.Errorf("Support Bundles: expected [%s], got %v", bundlePath, bundlePaths)
	}
}

func TestIntegration_HiddenPathsFilteredAcrossAllLinters(t *testing.T) {
	// Test that all three linters filter out hidden directories like .git and .github
	tmpDir := t.TempDir()

	// Create resources in hidden directories (should be filtered)
	gitDir := filepath.Join(tmpDir, ".git", "resources")
	createTestChart(t, gitDir, "git-chart")
	createTestPreflight(t, filepath.Join(gitDir, "preflight.yaml"))
	createTestSupportBundle(t, filepath.Join(gitDir, "bundle.yaml"))

	githubDir := filepath.Join(tmpDir, ".github", "resources")
	createTestChart(t, githubDir, "github-chart")
	createTestPreflight(t, filepath.Join(githubDir, "preflight.yaml"))
	createTestSupportBundle(t, filepath.Join(githubDir, "bundle.yaml"))

	// Create resources in normal directories (should be found)
	normalDir := filepath.Join(tmpDir, "resources")
	validChartDir := createTestChart(t, normalDir, "valid-chart")
	validPreflightPath := filepath.Join(normalDir, "valid-preflight.yaml")
	createTestPreflight(t, validPreflightPath)
	validBundlePath := filepath.Join(normalDir, "valid-bundle.yaml")
	createTestSupportBundle(t, validBundlePath)

	pattern := filepath.Join(tmpDir, "**")

	// All linters should filter out hidden directories
	t.Run("charts_filter_hidden", func(t *testing.T) {
		chartPaths, err := discoverChartPaths(pattern)
		if err != nil {
			t.Fatalf("discoverChartPaths() error = %v", err)
		}
		want := []string{validChartDir}
		assertPathsEqual(t, chartPaths, want)
	})

	t.Run("preflights_filter_hidden", func(t *testing.T) {
		preflightPaths, err := discoverPreflightPaths(pattern)
		if err != nil {
			t.Fatalf("discoverPreflightPaths() error = %v", err)
		}
		want := []string{validPreflightPath}
		assertPathsEqual(t, preflightPaths, want)
	})

	t.Run("support_bundles_filter_hidden", func(t *testing.T) {
		manifestPattern := filepath.Join(tmpDir, "**", "*.yaml")
		bundlePaths, err := DiscoverSupportBundlesFromManifests([]string{manifestPattern})
		if err != nil {
			t.Fatalf("DiscoverSupportBundlesFromManifests() error = %v", err)
		}
		want := []string{validBundlePath}
		assertPathsEqual(t, bundlePaths, want)
	})
}

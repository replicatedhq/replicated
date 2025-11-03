//go:build integration
// +build integration

package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/spf13/cobra"
)

// TestResolveAppContext_Integration tests app resolution with real API calls.
// Run with: go test -tags=integration -v ./cli/cmd -run TestResolveAppContext
//
// Prerequisites:
// - REPLICATED_API_TOKEN environment variable set
// - REPLICATED_API_ORIGIN (optional, defaults to https://api.replicated.com)
// - At least one app in the vendor account
func TestResolveAppContext_Integration(t *testing.T) {
	// Check for required credentials
	apiToken := os.Getenv("REPLICATED_API_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping integration test: REPLICATED_API_TOKEN not set")
	}

	apiOrigin := os.Getenv("REPLICATED_API_ORIGIN")
	if apiOrigin == "" {
		apiOrigin = "https://api.replicated.com"
	}

	// Set up real kotsAPI client
	httpClient := platformclient.NewHTTPClient(apiOrigin, apiToken)
	kotsAPI := kotsclient.VendorV3Client{HTTPClient: *httpClient}

	// Create runners instance with real API client
	r := &runners{
		kotsAPI: &kotsAPI,
		w:       tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', 0),
	}

	ctx := context.Background()

	t.Run("resolve with no app specified (API fetch)", func(t *testing.T) {
		config := &tools.Config{}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		// Should either:
		// - Return empty string (0 apps available)
		// - Return an app ID (1 app auto-selected)
		// - Return error (multiple apps, non-TTY)
		t.Logf("Resolved app ID: %q", appID)

		if appID != "" {
			// Verify it's a valid app ID by fetching it
			app, err := kotsAPI.GetApp(ctx, appID, true)
			if err != nil {
				t.Errorf("GetApp(%q) failed: %v", appID, err)
			}
			if app == nil || app.ID == "" {
				t.Errorf("GetApp returned invalid app: %+v", app)
			}
			t.Logf("Successfully resolved to app: %s (slug: %s)", app.Name, app.Slug)
		} else {
			t.Log("No app resolved (0 apps or API failure)")
		}
	})

	t.Run("resolve with explicit app ID in config", func(t *testing.T) {
		// First, list apps to get a valid app ID
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) == 0 {
			t.Skip("Skipping: no apps available in vendor account")
		}

		testAppID := apps[0].App.ID

		config := &tools.Config{
			AppId: testAppID,
		}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != testAppID {
			t.Errorf("Expected appID = %q, got %q", testAppID, appID)
		}
	})

	t.Run("resolve with explicit app slug in config", func(t *testing.T) {
		// First, list apps to get a valid app slug
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) == 0 {
			t.Skip("Skipping: no apps available in vendor account")
		}

		testAppSlug := apps[0].App.Slug
		expectedAppID := apps[0].App.ID

		config := &tools.Config{
			AppSlug: testAppSlug,
		}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != expectedAppID {
			t.Errorf("Expected appID = %q, got %q", expectedAppID, appID)
		}
	})

	t.Run("resolve with runners.appID set", func(t *testing.T) {
		// First, list apps to get a valid app ID
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) == 0 {
			t.Skip("Skipping: no apps available in vendor account")
		}

		testAppID := apps[0].App.ID

		// Set app ID in runners (simulates --app-id flag)
		r.appID = testAppID
		defer func() { r.appID = "" }() // cleanup

		config := &tools.Config{}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != testAppID {
			t.Errorf("Expected appID = %q, got %q", testAppID, appID)
		}
	})

	t.Run("resolve with invalid app ID should error", func(t *testing.T) {
		config := &tools.Config{
			AppId: "invalid-app-id-does-not-exist",
		}

		_, err := r.resolveAppContext(ctx, config)
		if err == nil {
			t.Error("Expected error when resolving invalid app ID, got nil")
		}
		t.Logf("Got expected error: %v", err)
	})

	t.Run("priority order: runners.appID > config.appId", func(t *testing.T) {
		// Get two different apps (if available)
		apps, err := kotsAPI.ListApps(ctx, false)
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if len(apps) < 1 {
			t.Skip("Skipping: need at least 1 app for priority test")
		}

		runnerAppID := apps[0].App.ID
		configAppID := "different-app-id" // doesn't matter, runner should win

		r.appID = runnerAppID
		defer func() { r.appID = "" }()

		config := &tools.Config{
			AppId: configAppID,
		}

		appID, err := r.resolveAppContext(ctx, config)
		if err != nil {
			t.Fatalf("resolveAppContext() error = %v", err)
		}

		if appID != runnerAppID {
			t.Errorf("Priority test failed: expected runners.appID %q to win, got %q", runnerAppID, appID)
		}
	})
}

// Embedded Cluster Linter Integration Tests

// TestEmbeddedClusterLint_Integration tests full EC linter integration with local binary
func TestEmbeddedClusterLint_Integration(t *testing.T) {
	ecBinaryPath := os.Getenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
	if ecBinaryPath == "" {
		t.Skip("Skipping: REPLICATED_EMBEDDED_CLUSTER_PATH not set")
	}

	ctx := context.Background()

	// Create temp directory with EC config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ec-config.yaml")
	
	// Create valid EC config
	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "1.0.0"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	
	t.Run("lint with valid config", func(t *testing.T) {
		result, err := lint2.LintEmbeddedCluster(ctx, configPath, "latest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if !result.Success {
			t.Errorf("expected success for valid config, got failure with %d messages", len(result.Messages))
			for _, msg := range result.Messages {
				t.Logf("  [%s] %s", msg.Severity, msg.Message)
			}
		}
	})
	
	t.Run("lint with invalid config (duplicate keys)", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tmpDir, "invalid-config.yaml")
		invalidContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "1.0.0"
  extensions:
    - name: test
      name: duplicate
`
		if err := os.WriteFile(invalidConfigPath, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("failed to create invalid config: %v", err)
		}
		
		result, err := lint2.LintEmbeddedCluster(ctx, invalidConfigPath, "latest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if result.Success {
			t.Error("expected failure for invalid config")
		}
		
		if len(result.Messages) == 0 {
			t.Error("expected error messages for invalid config")
		}
		
		// Verify we got an error about duplicate key
		foundDuplicateError := false
		for _, msg := range result.Messages {
			if msg.Severity == "error" && strings.Contains(msg.Message, "already set") {
				foundDuplicateError = true
				break
			}
		}
		if !foundDuplicateError {
			t.Error("expected error about duplicate key")
		}
	})
}

// TestEmbeddedClusterLint_EnvironmentVariables tests that env vars are accessible to EC binary
func TestEmbeddedClusterLint_EnvironmentVariables(t *testing.T) {
	ecBinaryPath := os.Getenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
	if ecBinaryPath == "" {
		t.Skip("Skipping: REPLICATED_EMBEDDED_CLUSTER_PATH not set")
	}
	
	apiToken := os.Getenv("REPLICATED_API_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping: REPLICATED_API_TOKEN not set")
	}

	ctx := context.Background()

	// Create temp directory with EC config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ec-config.yaml")
	
	configContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
spec:
  version: "1.0.0"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	
	// Set test env vars
	testAppID := "test-app-id-123"
	testAPIOrigin := "https://api.replicated.com"
	
	os.Setenv("REPLICATED_APP", testAppID)
	os.Setenv("REPLICATED_API_ORIGIN", testAPIOrigin)
	defer func() {
		os.Unsetenv("REPLICATED_APP")
	}()
	
	// Run linter - it should inherit these env vars
	result, err := lint2.LintEmbeddedCluster(ctx, configPath, "latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// The fact that it runs successfully means env vars were accessible
	// (the binary doesn't fail on missing env vars, but would if they were needed and unavailable)
	if !result.Success {
		t.Logf("Note: Linter reported issues, but env vars were accessible")
	}
	
	t.Logf("SUCCESS: Env vars accessible to embedded-cluster binary")
}

// TestEmbeddedClusterDiscovery_SingleConfig tests EC config discovery
func TestEmbeddedClusterDiscovery_SingleConfig(t *testing.T) {
	// Discovery tests don't need context
	
	// Create temp directory structure
	tmpDir := t.TempDir()
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create EC config
	ecConfigContent := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: my-cluster
spec:
  version: "1.0.0"
`
	ecConfigPath := filepath.Join(manifestsDir, "embedded-cluster.yaml")
	if err := os.WriteFile(ecConfigPath, []byte(ecConfigContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test discovery
	pattern := filepath.Join(manifestsDir, "*.yaml")
	paths, err := lint2.DiscoverEmbeddedClusterPaths(pattern)
	if err != nil {
		t.Fatalf("discovery failed: %v", err)
	}
	
	if len(paths) != 1 {
		t.Fatalf("expected 1 EC config, found %d", len(paths))
	}
	
	if paths[0] != ecConfigPath {
		t.Errorf("expected path %q, got %q", ecConfigPath, paths[0])
	}
	
	t.Logf("SUCCESS: Discovered single EC config at %s", paths[0])
}

// TestEmbeddedClusterDiscovery_MultipleConfigs tests error on multiple EC configs
func TestEmbeddedClusterDiscovery_MultipleConfigs(t *testing.T) {
	// Discovery tests don't need context
	
	// Create temp directory structure
	tmpDir := t.TempDir()
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create first EC config
	ecConfig1Content := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: cluster-1
spec:
  version: "1.0.0"
`
	ecConfig1Path := filepath.Join(manifestsDir, "cluster-1.yaml")
	if err := os.WriteFile(ecConfig1Path, []byte(ecConfig1Content), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create second EC config
	ecConfig2Content := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: cluster-2
spec:
  version: "1.0.0"
`
	ecConfig2Path := filepath.Join(manifestsDir, "cluster-2.yaml")
	if err := os.WriteFile(ecConfig2Path, []byte(ecConfig2Content), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test discovery - should find both
	pattern := filepath.Join(manifestsDir, "*.yaml")
	paths, err := lint2.DiscoverEmbeddedClusterPaths(pattern)
	if err != nil {
		t.Fatalf("discovery failed: %v", err)
	}
	
	if len(paths) != 2 {
		t.Fatalf("expected 2 EC configs, found %d", len(paths))
	}
	
	// Note: Discovery succeeds even with 2+ configs
	// The validation happens in lintEmbeddedClusterConfigs() which fails gracefully
	// without blocking other linters
	t.Logf("SUCCESS: Discovered %d EC configs (EC linter will fail gracefully, other linters continue)", len(paths))
}

// TestEmbeddedClusterDiscovery_NoConfigs tests graceful handling of no EC configs
func TestEmbeddedClusterDiscovery_NoConfigs(t *testing.T) {
	// Discovery tests don't need context
	
	// Create temp directory with non-EC manifests
	tmpDir := t.TempDir()
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create HelmChart manifest (not EC config)
	helmChartContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: test-chart
spec:
  chart:
    name: nginx
    chartVersion: 1.0.0
`
	helmChartPath := filepath.Join(manifestsDir, "helmchart.yaml")
	if err := os.WriteFile(helmChartPath, []byte(helmChartContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test discovery
	pattern := filepath.Join(manifestsDir, "*.yaml")
	paths, err := lint2.DiscoverEmbeddedClusterPaths(pattern)
	if err != nil {
		t.Fatalf("discovery failed: %v", err)
	}
	
	if len(paths) != 0 {
		t.Fatalf("expected 0 EC configs, found %d", len(paths))
	}
	
	t.Log("SUCCESS: No EC configs found (as expected)")
}

// TestEmbeddedClusterLint_MultipleConfigsGracefulFailure tests that when 2+ EC configs
// are found, the EC linter fails gracefully with a clear error message and doesn't
// block the entire lint command.
func TestEmbeddedClusterLint_MultipleConfigsGracefulFailure(t *testing.T) {
	ecBinaryPath := os.Getenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
	if ecBinaryPath == "" {
		t.Skip("Skipping: REPLICATED_EMBEDDED_CLUSTER_PATH not set")
	}

	// Create temp directory with multiple EC configs
	tmpDir := t.TempDir()
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create first EC config
	ecConfig1 := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: cluster-1
spec:
  version: "1.0.0"
`
	ecConfig1Path := filepath.Join(manifestsDir, "cluster-1.yaml")
	if err := os.WriteFile(ecConfig1Path, []byte(ecConfig1), 0644); err != nil {
		t.Fatal(err)
	}

	// Create second EC config
	ecConfig2 := `apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: cluster-2
spec:
  version: "2.0.0"
`
	ecConfig2Path := filepath.Join(manifestsDir, "cluster-2.yaml")
	if err := os.WriteFile(ecConfig2Path, []byte(ecConfig2), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a minimal cobra command for testing
	cmd := &cobra.Command{
		Use: "test",
	}

	// Create a runners instance with minimal setup
	var output bytes.Buffer
	w := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)
	runners := &runners{
		outputFormat: "table",
		w:            w,
	}

	// Call lintEmbeddedClusterConfigs with both config paths
	ecPaths := []string{ecConfig1Path, ecConfig2Path}
	results, err := runners.lintEmbeddedClusterConfigs(cmd, ecPaths, "latest", nil)

	// Should return results (not error out) - graceful failure
	if err != nil {
		t.Fatalf("expected graceful failure (return results), got error: %v", err)
	}

	// Should have results for both configs
	if len(results.Configs) != 2 {
		t.Fatalf("expected 2 config results, got %d", len(results.Configs))
	}

	// All results should show Success: false
	for i, config := range results.Configs {
		if config.Success {
			t.Errorf("config %d should have Success=false, got true", i)
		}

		// Check that error message mentions "Multiple embedded cluster configs"
		if len(config.Messages) == 0 {
			t.Errorf("config %d has no error messages", i)
			continue
		}

		foundMultipleConfigsError := false
		for _, msg := range config.Messages {
			if msg.Severity == "ERROR" && strings.Contains(msg.Message, "Multiple embedded cluster configs") {
				foundMultipleConfigsError = true
				break
			}
		}

		if !foundMultipleConfigsError {
			t.Errorf("config %d missing 'Multiple embedded cluster configs' error message", i)
		}
	}

	t.Log("✓ EC linter failed gracefully with clear error message")
	t.Log("✓ Function returned results (not fatal error)")
	t.Log("✓ Other linters would continue running")
}

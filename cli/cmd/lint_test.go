package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/spf13/cobra"
)

// getTestDataPath returns the absolute path to a test data file or directory.
// This helper is used to locate fixture files in the testdata/ directory.
func getTestDataPath(t *testing.T, relPath string) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Navigate from cli/cmd/ to project root
	projectRoot := filepath.Join(cwd, "..", "..")
	absPath := filepath.Join(projectRoot, relPath)

	// Verify the path exists
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("test data path does not exist: %s (error: %v)", absPath, err)
	}

	return absPath
}

// copyFixtureToTemp copies a fixture directory to a temporary directory for test isolation.
// This ensures tests can modify files without affecting the original fixtures.
func copyFixtureToTemp(t *testing.T, fixturePath string) string {
	t.Helper()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, filepath.Base(fixturePath))

	// Copy the fixture directory recursively
	if err := filepath.Walk(fixturePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from fixture root
		relPath, err := filepath.Rel(fixturePath, path)
		if err != nil {
			return err
		}

		destFilePath := filepath.Join(destPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(destFilePath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destFilePath, data, info.Mode())
	}); err != nil {
		t.Fatalf("failed to copy fixture from %s to %s: %v", fixturePath, destPath, err)
	}

	return destPath
}

func TestLint_VerboseFlag(t *testing.T) {
	// Use fixture for test data
	fixturePath := getTestDataPath(t, "testdata/lint/simple-chart")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name              string
		verbose           bool
		expectImageOutput bool
	}{
		{
			name:              "with verbose flag",
			verbose:           true,
			expectImageOutput: true,
		},
		{
			name:              "without verbose flag",
			verbose:           false,
			expectImageOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

			r := &runners{
				w: w,
				args: runnerArgs{
					lintVerbose: tt.verbose,
				},
			}

			// Load config
			parser := tools.NewConfigParser()
			config, err := parser.FindAndParseConfig(".")
			if err != nil {
				t.Fatalf("failed to load config: %v", err)
			}

			// Test extractImagesFromCharts
			// Extract charts with metadata
			charts, err := lint2.GetChartsWithMetadataFromConfig(config)
			if err != nil {
				t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
			}

			// Extract HelmChart manifests (if manifests configured)
			var helmChartManifests map[string]*lint2.HelmChartManifest
			if len(config.Manifests) > 0 {
				helmChartManifests, err = lint2.DiscoverHelmChartManifests(config.Manifests)
				if err != nil {
					// Only fail if error is not "no HelmChart resources found"
					if !strings.Contains(err.Error(), "no HelmChart resources found") {
						t.Fatalf("GetHelmChartManifestsFromConfig failed: %v", err)
					}
				}
			}

			imageResults, err := r.extractImagesFromCharts(context.Background(), charts, helmChartManifests)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if imageResults != nil {
				r.displayImages(imageResults)
			}

			w.Flush()
			output := buf.String()

			if tt.expectImageOutput {
				// Should contain image extraction output
				if !strings.Contains(output, "IMAGE EXTRACTION") {
					t.Error("expected 'IMAGE EXTRACTION' section header in verbose output")
				}
				if !strings.Contains(output, "nginx") {
					t.Error("expected to find nginx image in output")
				}
				if !strings.Contains(output, "Found") && !strings.Contains(output, "unique images") {
					t.Error("expected image count message in output")
				}
			}
		})
	}
}

func TestExtractAndDisplayImagesFromConfig_NoCharts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .replicated config with no charts section
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `repl-lint:
  linters:
    helm: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Should get error about no charts
	_, err = lint2.GetChartsWithMetadataFromConfig(config)
	if err == nil {
		t.Error("expected error when no charts in config")
	}
	if err != nil && !strings.Contains(err.Error(), "no charts found") {
		t.Errorf("expected 'no charts found' error, got: %v", err)
	}
}

func TestExtractAndDisplayImagesFromConfig_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .replicated config with non-existent chart
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: /nonexistent/chart/path
repl-lint:
  linters:
    helm: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Should get an error for non-existent chart path (validated by GetChartsWithMetadataFromConfig)
	_, err = lint2.GetChartsWithMetadataFromConfig(config)

	// We expect an error because the chart path doesn't exist
	if err == nil {
		t.Error("expected error for non-existent chart path")
	}

	// Since we got an error, we don't display anything
	// This is the correct behavior - fail fast on invalid paths
	// The test verified that we correctly return an error for non-existent paths
}

func TestExtractAndDisplayImagesFromConfig_MultipleCharts(t *testing.T) {
	// Use multi-chart fixture
	fixturePath := getTestDataPath(t, "testdata/lint/multi-chart-project")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w: w,
		args: runnerArgs{
			lintVerbose: true,
		},
	}

	// Load config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Extract images
	// Extract charts with metadata
	charts, err := lint2.GetChartsWithMetadataFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartsWithMetadataFromConfig failed: %v", err)
	}

	// Extract HelmChart manifests (if manifests configured)
	var helmChartManifests map[string]*lint2.HelmChartManifest
	if len(config.Manifests) > 0 {
		helmChartManifests, err = lint2.DiscoverHelmChartManifests(config.Manifests)
		if err != nil {
			// Only fail if error is not "no HelmChart resources found"
			if !strings.Contains(err.Error(), "no HelmChart resources found") {
				t.Fatalf("GetHelmChartManifestsFromConfig failed: %v", err)
			}
		}
	}

	imageResults, err := r.extractImagesFromCharts(context.Background(), charts, helmChartManifests)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if imageResults != nil {
		r.displayImages(imageResults)
	}

	w.Flush()
	output := buf.String()

	// Should find images from both charts
	if !strings.Contains(output, "nginx") {
		t.Error("expected to find nginx image from chart1")
	}
	if !strings.Contains(output, "redis") {
		t.Error("expected to find redis image from chart2")
	}
	// The new implementation shows total unique images instead of chart count
	if !strings.Contains(output, "unique images") {
		t.Error("expected message about unique images")
	}
}

// TestJSONOutputContainsAllToolVersions tests that JSON output includes all tool versions
func TestJSONOutputContainsAllToolVersions(t *testing.T) {
	// Use fixture and customize config with tool versions
	fixturePath := getTestDataPath(t, "testdata/lint/simple-chart")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Update .replicated config to add tool versions and all linters
	configPath := filepath.Join(testDir, ".replicated")
	configContent := `charts:
  - path: chart
repl-lint:
  version: 1
  linters:
    helm: {}
    preflight: {}
    support-bundle: {}
  tools:
    helm: "3.14.4"
    preflight: "0.123.9"
    support-bundle: "0.123.9"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "json",
		args: runnerArgs{
			lintVerbose: false, // Test without verbose - versions should still be in JSON
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	// We might get lint errors, but we should still get output
	// Ignore the error and check the output

	w.Flush()
	jsonOutput := buf.String()

	// Parse the JSON output
	var output JSONLintOutput
	if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
		// If we can't parse, check if there's output at all
		if jsonOutput == "" {
			t.Skip("No JSON output produced (likely due to missing tools)")
		}
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, jsonOutput)
	}

	// Check that all three tool versions are present in metadata
	if output.Metadata.HelmVersion == "" {
		t.Error("HelmVersion missing from JSON metadata")
	}
	if output.Metadata.PreflightVersion == "" {
		t.Error("PreflightVersion missing from JSON metadata")
	}
	if output.Metadata.SupportBundleVersion == "" {
		t.Error("SupportBundleVersion missing from JSON metadata")
	}

	// Check that versions match what was in config (not "latest")
	if output.Metadata.HelmVersion != "3.14.4" {
		t.Errorf("Expected HelmVersion to be '3.14.4', got '%s'", output.Metadata.HelmVersion)
	}
	if output.Metadata.PreflightVersion != "0.123.9" {
		t.Errorf("Expected PreflightVersion to be '0.123.9', got '%s'", output.Metadata.PreflightVersion)
	}
	if output.Metadata.SupportBundleVersion != "0.123.9" {
		t.Errorf("Expected SupportBundleVersion to be '0.123.9', got '%s'", output.Metadata.SupportBundleVersion)
	}

	t.Logf("JSON metadata contains all tool versions: Helm=%s, Preflight=%s, SupportBundle=%s",
		output.Metadata.HelmVersion,
		output.Metadata.PreflightVersion,
		output.Metadata.SupportBundleVersion)
}

// TestJSONOutputWithLatestVersions tests that "latest" in config resolves to actual versions
func TestJSONOutputWithLatestVersions(t *testing.T) {
	// This test may require network access to resolve "latest"
	if testing.Short() {
		t.Skip("Skipping test that requires network access in short mode")
	}

	// Create a temporary directory with a test chart
	tmpDir := t.TempDir()
	chartDir := filepath.Join(tmpDir, "test-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create minimal Chart.yaml
	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: test-chart
version: 1.0.0
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .replicated config with "latest" for all tools
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chartDir + `
repl-lint:
  version: 1
  linters:
    helm: {}
    preflight: {}
    support-bundle: {}
  tools:
    helm: "latest"
    preflight: "latest"
    support-bundle: "latest"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory for test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "json",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	_ = r.runLint(cmd, []string{}) // Ignore error, we care about the output

	w.Flush()
	jsonOutput := buf.String()

	// Parse the JSON output
	var output JSONLintOutput
	if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
		if jsonOutput == "" {
			t.Skip("No JSON output produced (likely network issue resolving latest versions)")
		}
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check that versions are resolved (not "latest")
	if output.Metadata.HelmVersion == "latest" {
		t.Error("HelmVersion should be resolved to actual version, not 'latest'")
	}
	if output.Metadata.PreflightVersion == "latest" {
		t.Error("PreflightVersion should be resolved to actual version, not 'latest'")
	}
	if output.Metadata.SupportBundleVersion == "latest" {
		t.Error("SupportBundleVersion should be resolved to actual version, not 'latest'")
	}

	// Check that versions look like semantic versions (x.y.z)
	if !isValidSemVer(output.Metadata.HelmVersion) {
		t.Errorf("HelmVersion doesn't look like a semantic version: %s", output.Metadata.HelmVersion)
	}
	if !isValidSemVer(output.Metadata.PreflightVersion) {
		t.Errorf("PreflightVersion doesn't look like a semantic version: %s", output.Metadata.PreflightVersion)
	}
	if !isValidSemVer(output.Metadata.SupportBundleVersion) {
		t.Errorf("SupportBundleVersion doesn't look like a semantic version: %s", output.Metadata.SupportBundleVersion)
	}

	t.Logf("'latest' resolved to: Helm=%s, Preflight=%s, SupportBundle=%s",
		output.Metadata.HelmVersion,
		output.Metadata.PreflightVersion,
		output.Metadata.SupportBundleVersion)
}

// TestConfigMissingToolVersions tests that missing tool versions default to "latest"
func TestConfigMissingToolVersions(t *testing.T) {
	// Use fixture and customize config without tool versions
	fixturePath := getTestDataPath(t, "testdata/lint/simple-chart")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Update .replicated config WITHOUT tool versions
	configPath := filepath.Join(testDir, ".replicated")
	configContent := `charts:
  - path: chart
repl-lint:
  version: 1
  linters:
    helm: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Load and parse config
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Check that ReplLint exists
	if config.ReplLint == nil {
		t.Fatal("ReplLint should be initialized")
	}

	// Debug to see what we have
	t.Logf("ReplLint: %+v", config.ReplLint)
	t.Logf("Tools is nil: %v", config.ReplLint.Tools == nil)

	// Check that tools map was initialized with "latest" defaults
	if config.ReplLint.Tools == nil {
		t.Fatal("Tools map should be initialized")
	}

	// Debug: print what's in the tools map
	t.Logf("Tools map contents: %+v", config.ReplLint.Tools)
	t.Logf("Number of tools in map: %d", len(config.ReplLint.Tools))

	// All tools should default to "latest"
	if v, ok := config.ReplLint.Tools[tools.ToolHelm]; !ok {
		t.Error("Helm tool not found in config")
	} else if v != "latest" {
		t.Errorf("Expected Helm to default to 'latest', got '%s'", v)
	}
	if v, ok := config.ReplLint.Tools[tools.ToolPreflight]; !ok {
		t.Error("Preflight tool not found in config")
	} else if v != "latest" {
		t.Errorf("Expected Preflight to default to 'latest', got '%s'", v)
	}
	if v, ok := config.ReplLint.Tools[tools.ToolSupportBundle]; !ok {
		t.Error("SupportBundle tool not found in config")
	} else if v != "latest" {
		t.Errorf("Expected SupportBundle to default to 'latest', got '%s'", v)
	}
}

// Helper function to check if a string looks like a semantic version
func isValidSemVer(version string) bool {
	// Basic check: should contain at least one dot and start with a digit
	// Examples: "3.14.4", "0.123.9", "v3.14.4"
	if version == "" {
		return false
	}
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Should have format x.y.z or x.y
	parts := strings.Split(version, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}

	// Each part should be numeric
	for _, part := range parts {
		if part == "" {
			return false
		}
		// Check first character is a digit
		if part[0] < '0' || part[0] > '9' {
			return false
		}
	}

	return true
}

// TestLint_ChartValidationError tests that lint fails when a chart is missing its HelmChart manifest
func TestLint_ChartValidationError(t *testing.T) {
	// Use fixture with chart but no HelmChart manifest
	fixturePath := getTestDataPath(t, "testdata/lint/chart-missing-helmchart")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command - should fail
	err = r.runLint(cmd, []string{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	// Error should mention the missing HelmChart manifest
	errMsg := err.Error()
	if !strings.Contains(errMsg, "chart validation failed") {
		t.Errorf("error should mention 'chart validation failed': %s", errMsg)
	}
	if !strings.Contains(errMsg, "test-app") {
		t.Errorf("error should contain chart name 'test-app': %s", errMsg)
	}
	if !strings.Contains(errMsg, "1.0.0") {
		t.Errorf("error should contain version '1.0.0': %s", errMsg)
	}
	if !strings.Contains(errMsg, "HelmChart manifest") {
		t.Errorf("error should mention HelmChart manifest: %s", errMsg)
	}
}

// TestLint_ChartValidationWarning tests that lint succeeds but shows warning for orphaned HelmChart manifest
func TestLint_ChartValidationWarning(t *testing.T) {
	// Use fixture with matching HelmChart + orphaned HelmChart manifest
	fixturePath := getTestDataPath(t, "testdata/lint/orphaned-helmchart")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command - should succeed
	err = r.runLint(cmd, []string{})
	// Note: err might be non-nil due to disabled linters not running
	// We care about the output showing the warning

	w.Flush()
	output := buf.String()

	// Output should contain warning about orphaned manifest
	if !strings.Contains(output, "Warning") {
		t.Error("expected warning message in output")
	}
	if !strings.Contains(output, "old-app") {
		t.Errorf("warning should mention orphaned chart 'old-app': %s", output)
	}
	if !strings.Contains(output, "no corresponding chart") {
		t.Errorf("warning should explain the issue: %s", output)
	}
}

// TestLint_NoManifestsConfig tests error when charts configured but manifests section missing
func TestLint_NoManifestsConfig(t *testing.T) {
	// Use fixture and modify config to remove manifests section
	fixturePath := getTestDataPath(t, "testdata/lint/chart-missing-helmchart")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Update config to remove manifests section
	configPath := filepath.Join(testDir, ".replicated")
	configContent := `charts:
  - path: chart
repl-lint:
  linters:
    helm:
      disabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command - should fail early
	err = r.runLint(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when charts configured but no manifests, got nil")
	}

	// Error should explain the problem clearly
	errMsg := err.Error()
	if !strings.Contains(errMsg, "charts are configured") {
		t.Errorf("error should mention charts are configured: %s", errMsg)
	}
	if !strings.Contains(errMsg, "no manifests") {
		t.Errorf("error should mention missing manifests: %s", errMsg)
	}
	if !strings.Contains(errMsg, "HelmChart manifest") {
		t.Errorf("error should mention HelmChart manifest requirement: %s", errMsg)
	}
	if !strings.Contains(errMsg, ".replicated") || !strings.Contains(errMsg, "manifests:") {
		t.Errorf("error should provide actionable guidance: %s", errMsg)
	}
}

// TestLint_AutodiscoveryWithMixedManifests tests that autodiscovery works
// when manifests directory contains BOTH HelmChart manifests and Support Bundle specs,
// and also includes Preflight specs.
func TestLint_AutodiscoveryWithMixedManifests(t *testing.T) {
	// Use fixture with mixed manifests (no .replicated to trigger autodiscovery)
	fixturePath := getTestDataPath(t, "testdata/lint/mixed-manifests-autodiscovery")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})

	// EXPECTED (after fix): Should succeed
	// ACTUAL (before fix): Fails with "file does not contain kind: SupportBundle"
	if err != nil {
		errMsg := err.Error()
		// Check if it's the expected bug
		if strings.Contains(errMsg, "file does not contain kind: SupportBundle") {
			t.Fatalf("BUG REPRODUCED: autodiscovery fails with mixed manifests: %v\n\n"+
				"This is the bug we're fixing. Autodiscovery stores explicit paths which triggers\n"+
				"strict validation that fails on mixed resource types.", err)
		}
		if strings.Contains(errMsg, "file does not contain kind: Preflight") {
			t.Fatalf("BUG REPRODUCED: autodiscovery fails when preflight in manifests dir: %v\n\n"+
				"This is the bug we're fixing. Autodiscovery stores explicit paths which triggers\n"+
				"strict validation that fails on mixed resource types.", err)
		}
		// If it's a different error, report it
		t.Fatalf("unexpected error (not the bug we're testing): %v", err)
	}

	// Verify autodiscovery found resources
	output := buf.String()
	if !strings.Contains(output, "Auto-discovering lintable resources") {
		t.Error("expected autodiscovery message in output")
	}
	if !strings.Contains(output, "1 Helm chart(s)") {
		t.Error("expected to discover 1 chart")
	}
	if !strings.Contains(output, "1 Preflight spec(s)") {
		t.Error("expected to discover 1 preflight spec")
	}
	if !strings.Contains(output, "1 Support Bundle spec(s)") {
		t.Error("expected to discover 1 support bundle")
	}
	if !strings.Contains(output, "1 HelmChart manifest(s)") {
		t.Error("expected to discover 1 HelmChart manifest")
	}

	t.Log("SUCCESS: Autodiscovery correctly handled mixed HelmChart, Support Bundle, and Preflight manifests")
}

// TestLint_AutodiscoveryEmptyProject tests that autodiscovery handles empty directories gracefully
func TestLint_AutodiscoveryEmptyProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty directory - no resources

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})

	// Should succeed with message about no resources
	if err != nil {
		t.Fatalf("expected success with empty project, got error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No lintable resources found") {
		t.Error("expected 'No lintable resources found' message in output")
	}

	t.Log("SUCCESS: Autodiscovery correctly handled empty project")
}

// TestLint_AutodiscoveryChartWithoutHelmChart tests that autodiscovery fails
// when a chart is discovered but no corresponding HelmChart manifest exists
func TestLint_AutodiscoveryChartWithoutHelmChart(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a chart
	chartDir := filepath.Join(tmpDir, "charts", "my-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: missing-helmchart
version: 1.0.0
description: Chart without HelmChart manifest
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	valuesYaml := filepath.Join(chartDir, "values.yaml")
	if err := os.WriteFile(valuesYaml, []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create manifests directory with a Support Bundle (but no HelmChart for the chart)
	// This ensures config.Manifests gets populated, triggering validation
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Add a Support Bundle so manifests discovery happens
	sbFile := filepath.Join(manifestsDir, "support-bundle.yaml")
	sbContent := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: some-bundle
spec:
  collectors: []
`
	if err := os.WriteFile(sbFile, []byte(sbContent), 0644); err != nil {
		t.Fatal(err)
	}

	// NO .replicated config - trigger autodiscovery

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command - should fail validation
	err = r.runLint(cmd, []string{})
	if err == nil {
		t.Fatal("expected validation error for chart without HelmChart manifest, got nil")
	}

	// Error should mention validation failure
	errMsg := err.Error()
	if !strings.Contains(errMsg, "chart validation failed") {
		t.Errorf("error should mention 'chart validation failed': %s", errMsg)
	}
	if !strings.Contains(errMsg, "missing-helmchart") {
		t.Errorf("error should contain chart name: %s", errMsg)
	}

	t.Log("SUCCESS: Autodiscovery correctly detected missing HelmChart manifest")
}

// TestLint_AutodiscoveryMultipleCharts tests autodiscovery with multiple charts
func TestLint_AutodiscoveryMultipleCharts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create three charts
	for i := 1; i <= 3; i++ {
		chartDir := filepath.Join(tmpDir, "charts", fmt.Sprintf("chart%d", i))
		if err := os.MkdirAll(chartDir, 0755); err != nil {
			t.Fatal(err)
		}

		chartYaml := filepath.Join(chartDir, "Chart.yaml")
		chartContent := fmt.Sprintf(`apiVersion: v2
name: app%d
version: 1.0.0
description: Chart %d
`, i, i)
		if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
			t.Fatal(err)
		}

		valuesYaml := filepath.Join(chartDir, "values.yaml")
		if err := os.WriteFile(valuesYaml, []byte("replicaCount: 1\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create manifests with HelmCharts for all three
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 3; i++ {
		helmChartFile := filepath.Join(manifestsDir, fmt.Sprintf("helmchart%d.yaml", i))
		helmChartContent := fmt.Sprintf(`apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: app%d-chart
spec:
  chart:
    name: app%d
    chartVersion: 1.0.0
  builder: {}
`, i, i)
		if err := os.WriteFile(helmChartFile, []byte(helmChartContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// NO .replicated config - trigger autodiscovery

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error with multiple charts: %v", err)
	}

	// Verify all charts were discovered
	output := buf.String()
	if !strings.Contains(output, "3 Helm chart(s)") {
		t.Error("expected to discover 3 charts")
	}
	if !strings.Contains(output, "3 HelmChart manifest(s)") {
		t.Error("expected to discover 3 HelmChart manifests")
	}

	t.Log("SUCCESS: Autodiscovery correctly handled multiple charts")
}

// TestLint_AutodiscoveryHiddenDirectories tests that hidden directories (.git, .github) are ignored
func TestLint_AutodiscoveryHiddenDirectories(t *testing.T) {
	// Use fixture with hidden .github directory and real chart
	fixturePath := getTestDataPath(t, "testdata/lint/hidden-dirs-test")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Create .git directory with fake chart (can't commit .git dirs to git, so create at runtime)
	gitDir := filepath.Join(testDir, ".git", "charts")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	gitChartYaml := filepath.Join(gitDir, "Chart.yaml")
	if err := os.WriteFile(gitChartYaml, []byte("apiVersion: v2\nname: fake\nversion: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify only 1 chart discovered (hidden dirs ignored)
	output := buf.String()
	t.Logf("Output:\n%s", output)
	if !strings.Contains(output, "1 Helm chart(s)") {
		t.Error("expected to discover exactly 1 chart (hidden dirs should be ignored)")
	}
	if !strings.Contains(output, "1 HelmChart manifest(s)") {
		t.Error("expected to discover exactly 1 HelmChart manifest")
	}
	// Check for multiple charts/manifests being discovered (would indicate hidden dirs not ignored)
	if strings.Contains(output, "2 Helm chart(s)") || strings.Contains(output, "2 HelmChart manifest(s)") {
		t.Error("should not discover resources from hidden directories")
	}

	t.Log("SUCCESS: Autodiscovery correctly ignored hidden directories")
}

// TestLint_AutodiscoveryBothYamlExtensions tests that both .yaml and .yml files are discovered
func TestLint_AutodiscoveryBothYamlExtensions(t *testing.T) {
	// Use fixture with mixed .yaml and .yml extensions
	fixturePath := getTestDataPath(t, "testdata/lint/mixed-manifests-yaml-yml")
	testDir := copyFixtureToTemp(t, fixturePath)

	// Change to test directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all resources discovered with different extensions
	output := buf.String()
	if !strings.Contains(output, "1 Helm chart(s)") {
		t.Error("expected to discover chart")
	}
	if !strings.Contains(output, "1 Preflight spec(s)") {
		t.Error("expected to discover preflight (.yaml)")
	}
	if !strings.Contains(output, "1 Support Bundle spec(s)") {
		t.Error("expected to discover support bundle (.yml)")
	}
	if !strings.Contains(output, "1 HelmChart manifest(s)") {
		t.Error("expected to discover HelmChart (.yaml)")
	}

	t.Log("SUCCESS: Autodiscovery correctly handled both .yaml and .yml extensions for manifests")
}

// Tests for config discovery vs autodiscovery behavior

// TestLint_SubdirectoryWithParentConfig tests that running lint from a subdirectory
// uses the parent's .replicated config and does NOT trigger autodiscovery.
func TestLint_SubdirectoryWithParentConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create chart1 in parent directory
	chart1Dir := filepath.Join(tmpDir, "chart1")
	if err := os.MkdirAll(chart1Dir, 0755); err != nil {
		t.Fatal(err)
	}

	chart1Yaml := filepath.Join(chart1Dir, "Chart.yaml")
	chart1Content := `apiVersion: v2
name: app1
version: 1.0.0
description: Chart in parent config
`
	if err := os.WriteFile(chart1Yaml, []byte(chart1Content), 0644); err != nil {
		t.Fatal(err)
	}

	valuesYaml := filepath.Join(chart1Dir, "values.yaml")
	if err := os.WriteFile(valuesYaml, []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create HelmChart manifest for chart1
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	helmChartFile := filepath.Join(manifestsDir, "helmchart1.yaml")
	helmChartContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: app1-chart
spec:
  chart:
    name: app1
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(helmChartFile, []byte(helmChartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .replicated config in parent with chart1
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chart1Dir + `
manifests:
  - ` + manifestsDir + `/**
repl-lint:
  linters:
    helm: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with chart2 (NOT in parent config)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	chart2Dir := filepath.Join(subDir, "chart2")
	if err := os.MkdirAll(chart2Dir, 0755); err != nil {
		t.Fatal(err)
	}

	chart2Yaml := filepath.Join(chart2Dir, "Chart.yaml")
	chart2Content := `apiVersion: v2
name: app2
version: 1.0.0
description: Chart NOT in parent config
`
	if err := os.WriteFile(chart2Yaml, []byte(chart2Content), 0644); err != nil {
		t.Fatal(err)
	}

	chart2Values := filepath.Join(chart2Dir, "values.yaml")
	if err := os.WriteFile(chart2Values, []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to subdirectory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that autodiscovery did NOT trigger
	output := buf.String()
	if strings.Contains(output, "Auto-discovering lintable resources") {
		t.Error("autodiscovery should NOT trigger when parent config exists with resources")
	}

	// Verify chart1 from parent config was linted
	if !strings.Contains(output, chart1Dir) {
		t.Errorf("expected chart1 from parent config to be linted, got output:\n%s", output)
	}

	// Verify chart2 was NOT linted (not in parent config)
	if strings.Contains(output, "app2") {
		t.Error("chart2 should NOT be linted (not in parent config)")
	}

	t.Log("SUCCESS: Subdirectory correctly used parent config instead of autodiscovery")
}

// TestLint_SubdirectoryWithEmptyParentConfig tests that running lint from a subdirectory
// with an empty parent config DOES trigger autodiscovery.
func TestLint_SubdirectoryWithEmptyParentConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty .replicated config in parent (no charts/preflights/manifests)
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `repl-lint:
  linters:
    helm: {}
    preflight: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with resources
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create chart in subdirectory
	chartDir := filepath.Join(subDir, "my-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: discovered-app
version: 1.0.0
description: Chart discovered by autodiscovery
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	valuesYaml := filepath.Join(chartDir, "values.yaml")
	if err := os.WriteFile(valuesYaml, []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create HelmChart manifest for the discovered chart
	helmChartFile := filepath.Join(subDir, "helmchart.yaml")
	helmChartContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: discovered-app-chart
spec:
  chart:
    name: discovered-app
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(helmChartFile, []byte(helmChartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create preflight in subdirectory
	preflightFile := filepath.Join(subDir, "preflight.yaml")
	preflightContent := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: discovered-preflight
spec:
  analyzers:
    - clusterVersion:
        outcomes:
          - pass:
              message: Valid
`
	if err := os.WriteFile(preflightFile, []byte(preflightContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to subdirectory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify autodiscovery DID trigger
	output := buf.String()
	if !strings.Contains(output, "Auto-discovering lintable resources") {
		t.Error("autodiscovery SHOULD trigger when parent config is empty")
	}

	// Verify resources were discovered
	if !strings.Contains(output, "1 Helm chart(s)") {
		t.Error("expected to discover 1 chart")
	}
	if !strings.Contains(output, "1 Preflight spec(s)") {
		t.Error("expected to discover 1 preflight")
	}

	t.Log("SUCCESS: Empty parent config correctly triggered autodiscovery")
}

// TestLint_MonorepoMultipleConfigs tests that multiple .replicated configs
// are merged correctly and autodiscovery does NOT trigger.
func TestLint_MonorepoMultipleConfigs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create app1 chart at grandparent level
	app1Dir := filepath.Join(tmpDir, "app1")
	if err := os.MkdirAll(app1Dir, 0755); err != nil {
		t.Fatal(err)
	}

	app1Chart := filepath.Join(app1Dir, "Chart.yaml")
	app1Content := `apiVersion: v2
name: app1
version: 1.0.0
description: Grandparent app
`
	if err := os.WriteFile(app1Chart, []byte(app1Content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(app1Dir, "values.yaml"), []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create HelmChart manifest for app1
	app1HelmChart := filepath.Join(tmpDir, "app1-helmchart.yaml")
	app1HelmChartContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: app1-chart
spec:
  chart:
    name: app1
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(app1HelmChart, []byte(app1HelmChartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create grandparent .replicated with app1
	grandparentConfig := filepath.Join(tmpDir, ".replicated")
	grandparentContent := `charts:
  - path: ` + app1Dir + `
manifests:
  - ` + tmpDir + `/*.yaml
repl-lint:
  linters:
    helm: {}
`
	if err := os.WriteFile(grandparentConfig, []byte(grandparentContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create parent directory
	parentDir := filepath.Join(tmpDir, "parent")
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create app2 chart at parent level
	app2Dir := filepath.Join(parentDir, "app2")
	if err := os.MkdirAll(app2Dir, 0755); err != nil {
		t.Fatal(err)
	}

	app2Chart := filepath.Join(app2Dir, "Chart.yaml")
	app2Content := `apiVersion: v2
name: app2
version: 1.0.0
description: Parent app
`
	if err := os.WriteFile(app2Chart, []byte(app2Content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(app2Dir, "values.yaml"), []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create HelmChart manifest for app2
	app2HelmChart := filepath.Join(parentDir, "app2-helmchart.yaml")
	app2HelmChartContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: app2-chart
spec:
  chart:
    name: app2
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(app2HelmChart, []byte(app2HelmChartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create parent .replicated with app2
	parentConfig := filepath.Join(parentDir, ".replicated")
	parentContent := `charts:
  - path: ` + app2Dir + `
manifests:
  - ` + parentDir + `/*.yaml
`
	if err := os.WriteFile(parentConfig, []byte(parentContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create child directory
	childDir := filepath.Join(parentDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to child directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(childDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify autodiscovery did NOT trigger
	output := buf.String()
	if strings.Contains(output, "Auto-discovering lintable resources") {
		t.Error("autodiscovery should NOT trigger when merged configs have resources")
	}

	// Verify BOTH apps were linted (merged from grandparent and parent configs)
	if !strings.Contains(output, "app1") {
		t.Error("expected app1 from grandparent config to be linted")
	}
	if !strings.Contains(output, "app2") {
		t.Error("expected app2 from parent config to be linted")
	}

	t.Log("SUCCESS: Multiple configs correctly merged and both apps linted")
}

// TestLint_AutodiscoveryOnlyWhenAllArraysEmpty tests that autodiscovery
// only triggers when ALL resource arrays are empty.
func TestLint_AutodiscoveryOnlyWhenAllArraysEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	// Create support bundle manifest
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	sbFile := filepath.Join(manifestsDir, "support-bundle.yaml")
	sbContent := `apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: configured-sb
spec:
  collectors:
    - logs:
        selector:
          - app=test
`
	if err := os.WriteFile(sbFile, []byte(sbContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .replicated config with ONLY manifests (no charts/preflights)
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `manifests:
  - ` + manifestsDir + `/**
repl-lint:
  linters:
    support-bundle: {}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with a chart (NOT in config)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(subDir, "chart1")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: undiscovered-app
version: 1.0.0
description: Should NOT be discovered
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to subdirectory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify autodiscovery did NOT trigger
	output := buf.String()
	if strings.Contains(output, "Auto-discovering lintable resources") {
		t.Error("autodiscovery should NOT trigger when manifests array is non-empty")
	}

	// Verify chart was NOT discovered
	if strings.Contains(output, "undiscovered-app") {
		t.Error("chart should NOT be discovered when config has manifests")
	}

	// Verify support bundle from config was used
	if !strings.Contains(output, sbFile) {
		t.Errorf("expected support bundle from parent config to be linted, got output:\n%s", output)
	}

	t.Log("SUCCESS: Autodiscovery correctly did NOT trigger with non-empty manifests array")
}

// TestLint_NoConfigAnywhere tests autodiscovery when no .replicated config
// exists anywhere in the directory tree.
func TestLint_NoConfigAnywhere(t *testing.T) {
	tmpDir := t.TempDir()

	// Create deep nested directory structure with NO .replicated anywhere
	deepDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create chart in deep directory
	chartDir := filepath.Join(deepDir, "my-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: deep-app
version: 1.0.0
description: Chart in deep directory
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte("replicaCount: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create HelmChart manifest
	helmChartFile := filepath.Join(deepDir, "helmchart.yaml")
	helmChartContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: deep-app-chart
spec:
  chart:
    name: deep-app
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(helmChartFile, []byte(helmChartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to deep directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(deepDir); err != nil {
		t.Fatal(err)
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)

	r := &runners{
		w:            w,
		outputFormat: "table",
		args: runnerArgs{
			lintVerbose: false,
		},
	}

	// Create a mock command with context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Run the lint command
	err = r.runLint(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify autodiscovery triggered
	output := buf.String()
	if !strings.Contains(output, "Auto-discovering lintable resources") {
		t.Error("autodiscovery SHOULD trigger when no config exists anywhere")
	}

	// Verify chart was discovered
	if !strings.Contains(output, "1 Helm chart(s)") {
		t.Error("expected to discover 1 chart")
	}

	t.Log("SUCCESS: Autodiscovery correctly triggered with no config in directory tree")
}

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/spf13/cobra"
)

func TestLint_VerboseFlag(t *testing.T) {
	// Create a temporary directory with a test chart
	tmpDir := t.TempDir()
	chartDir := filepath.Join(tmpDir, "test-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create Chart.yaml
	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: test-chart
version: 1.0.0
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create templates directory with a deployment
	templatesDir := filepath.Join(chartDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	deploymentYaml := filepath.Join(templatesDir, "deployment.yaml")
	deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
`
	if err := os.WriteFile(deploymentYaml, []byte(deploymentContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create manifests directory with HelmChart
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	helmChartManifest := filepath.Join(manifestsDir, "helmchart.yaml")
	helmChartContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: test-chart
spec:
  chart:
    name: test-chart
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(helmChartManifest, []byte(helmChartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .replicated config
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chartDir + `
manifests:
  - ` + manifestsDir + `/*.yaml
repl-lint:
  linters:
    helm: {}
    preflight:
      disabled: true
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

	// Create .replicated config with no charts
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `repl-lint:
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
	// Create a temporary directory with multiple test charts
	tmpDir := t.TempDir()

	// Create first chart
	chart1Dir := filepath.Join(tmpDir, "chart1")
	if err := os.MkdirAll(filepath.Join(chart1Dir, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	chart1Yaml := filepath.Join(chart1Dir, "Chart.yaml")
	if err := os.WriteFile(chart1Yaml, []byte("apiVersion: v2\nname: chart1\nversion: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	dep1Yaml := filepath.Join(chart1Dir, "templates", "deployment.yaml")
	dep1Content := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test1
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
`
	if err := os.WriteFile(dep1Yaml, []byte(dep1Content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create second chart with different image
	chart2Dir := filepath.Join(tmpDir, "chart2")
	if err := os.MkdirAll(filepath.Join(chart2Dir, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	chart2Yaml := filepath.Join(chart2Dir, "Chart.yaml")
	if err := os.WriteFile(chart2Yaml, []byte("apiVersion: v2\nname: chart2\nversion: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	dep2Yaml := filepath.Join(chart2Dir, "templates", "deployment.yaml")
	dep2Content := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test2
spec:
  template:
    spec:
      containers:
      - name: redis
        image: redis:7.0
`
	if err := os.WriteFile(dep2Yaml, []byte(dep2Content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create manifests directory with HelmChart manifests for both charts
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	helmChart1 := filepath.Join(manifestsDir, "chart1-helmchart.yaml")
	helmChart1Content := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: chart1
spec:
  chart:
    name: chart1
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(helmChart1, []byte(helmChart1Content), 0644); err != nil {
		t.Fatal(err)
	}

	helmChart2 := filepath.Join(manifestsDir, "chart2-helmchart.yaml")
	helmChart2Content := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: chart2
spec:
  chart:
    name: chart2
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(helmChart2, []byte(helmChart2Content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .replicated config with both charts
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chart1Dir + `
  - path: ` + chart2Dir + `
manifests:
  - ` + manifestsDir + `/*.yaml
repl-lint:
  linters:
    helm: {}
    preflight:
      disabled: true
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

	// Create .replicated config with specific tool versions
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
    helm: "3.14.4"
    preflight: "0.123.9"
    support-bundle: "0.123.9"
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

	// Create .replicated config WITHOUT tool versions
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chartDir + `
repl-lint:
  version: 1
  linters:
    helm: {}
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
	tmpDir := t.TempDir()

	// Create a chart
	chartDir := filepath.Join(tmpDir, "test-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: test-app
version: 1.0.0
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create empty manifests directory (no HelmChart manifest)
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with chart but no matching HelmChart manifest
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chartDir + `
manifests:
  - ` + manifestsDir + `/*.yaml
repl-lint:
  linters:
    helm:
      disabled: true
    preflight:
      disabled: true
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
	tmpDir := t.TempDir()

	// Create a chart
	chartDir := filepath.Join(tmpDir, "test-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: current-app
version: 1.0.0
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create manifests directory with matching HelmChart + orphaned one
	manifestsDir := filepath.Join(tmpDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Matching HelmChart manifest
	currentHelmChart := filepath.Join(manifestsDir, "current-app.yaml")
	currentContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: current-app
spec:
  chart:
    name: current-app
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(currentHelmChart, []byte(currentContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Orphaned HelmChart manifest
	oldHelmChart := filepath.Join(manifestsDir, "old-app.yaml")
	oldContent := `apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: old-app
spec:
  chart:
    name: old-app
    chartVersion: 1.0.0
  builder: {}
`
	if err := os.WriteFile(oldHelmChart, []byte(oldContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chartDir + `
manifests:
  - ` + manifestsDir + `/*.yaml
repl-lint:
  linters:
    helm:
      disabled: true
    preflight:
      disabled: true
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
	tmpDir := t.TempDir()

	// Create a chart
	chartDir := filepath.Join(tmpDir, "test-chart")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := `apiVersion: v2
name: test-app
version: 1.0.0
`
	if err := os.WriteFile(chartYaml, []byte(chartContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config WITHOUT manifests section
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chartDir + `
repl-lint:
  linters:
    helm:
      disabled: true
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

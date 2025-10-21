package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/tools"
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

	// Create .replicated config
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chartDir + `
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

			// Test extractImagesFromConfig
			imageResults, err := r.extractImagesFromConfig(context.Background(), config)

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
				if !strings.Contains(output, "Extracting images") {
					t.Error("expected 'Extracting images' message in verbose output")
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

	// Should handle no charts gracefully
	imageResults, err := r.extractImagesFromConfig(context.Background(), config)
	if err == nil && imageResults != nil {
		r.displayImages(imageResults)
	}

	// Should get error about no charts
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

	// Should get an error for non-existent chart path (validated by GetChartPathsFromConfig)
	_, err = r.extractImagesFromConfig(context.Background(), config)

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

	// Create .replicated config with both charts
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `charts:
  - path: ` + chart1Dir + `
  - path: ` + chart2Dir + `
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
	imageResults, err := r.extractImagesFromConfig(context.Background(), config)
	if err == nil && imageResults != nil {
		r.displayImages(imageResults)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
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

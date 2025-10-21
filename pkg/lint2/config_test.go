package lint2

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/replicatedhq/replicated/pkg/tools"
)

func TestGetChartPathsFromConfig(t *testing.T) {
	// Create a test chart directory
	tmpDir := t.TempDir()
	validChartDir := filepath.Join(tmpDir, "valid-chart")
	if err := os.MkdirAll(validChartDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create Chart.yaml
	chartYaml := filepath.Join(validChartDir, "Chart.yaml")
	if err := os.WriteFile(chartYaml, []byte("name: test\nversion: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		config    *tools.Config
		wantPaths []string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "no charts in config",
			config: &tools.Config{
				Charts: []tools.ChartConfig{},
			},
			wantErr: true,
			errMsg:  "no charts found",
		},
		{
			name: "single chart path",
			config: &tools.Config{
				Charts: []tools.ChartConfig{
					{Path: validChartDir},
				},
			},
			wantPaths: []string{validChartDir},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := GetChartPathsFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChartPathsFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetChartPathsFromConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			// Validate actual paths match expected
			if !tt.wantErr {
				if len(paths) != len(tt.wantPaths) {
					t.Errorf("GetChartPathsFromConfig() returned %d paths, want %d", len(paths), len(tt.wantPaths))
					return
				}
				for i, path := range paths {
					if path != tt.wantPaths[i] {
						t.Errorf("GetChartPathsFromConfig() path[%d] = %q, want %q", i, path, tt.wantPaths[i])
					}
				}
			}
		})
	}
}

func TestGetChartPathsFromConfig_GlobExpansion(t *testing.T) {
	// Create test directory structure with multiple charts
	tmpDir := t.TempDir()

	// Create charts directory with multiple charts
	chartsDir := filepath.Join(tmpDir, "charts")
	chart1Dir := filepath.Join(chartsDir, "chart1")
	chart2Dir := filepath.Join(chartsDir, "chart2")
	chart3Dir := filepath.Join(tmpDir, "standalone-chart")

	for _, dir := range []string{chart1Dir, chart2Dir, chart3Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		chartYaml := filepath.Join(dir, "Chart.yaml")
		if err := os.WriteFile(chartYaml, []byte("name: test\nversion: 1.0.0\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		config    *tools.Config
		wantPaths []string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "glob pattern expansion",
			config: &tools.Config{
				Charts: []tools.ChartConfig{
					{Path: filepath.Join(chartsDir, "*")},
				},
			},
			wantPaths: []string{chart1Dir, chart2Dir},
			wantErr:   false,
		},
		{
			name: "multiple charts - mixed glob and direct",
			config: &tools.Config{
				Charts: []tools.ChartConfig{
					{Path: filepath.Join(chartsDir, "*")},
					{Path: chart3Dir},
				},
			},
			wantPaths: []string{chart1Dir, chart2Dir, chart3Dir},
			wantErr:   false,
		},
		{
			name: "glob with no matches",
			config: &tools.Config{
				Charts: []tools.ChartConfig{
					{Path: filepath.Join(tmpDir, "nonexistent", "*")},
				},
			},
			wantErr: true,
			errMsg:  "no directories found matching pattern",
		},
		{
			name: "glob pattern in current directory",
			config: &tools.Config{
				Charts: []tools.ChartConfig{
					{Path: filepath.Join(chartsDir, "chart*")},
				},
			},
			wantPaths: []string{chart1Dir, chart2Dir},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := GetChartPathsFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChartPathsFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetChartPathsFromConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			// Validate actual paths match expected (for success cases)
			if !tt.wantErr {
				if len(paths) != len(tt.wantPaths) {
					t.Errorf("GetChartPathsFromConfig() returned %d paths, want %d", len(paths), len(tt.wantPaths))
					return
				}
				// Build map of expected paths for order-independent comparison
				expectedPaths := make(map[string]bool)
				for _, p := range tt.wantPaths {
					expectedPaths[p] = false
				}
				// Mark found paths
				for _, path := range paths {
					if _, ok := expectedPaths[path]; ok {
						expectedPaths[path] = true
					} else {
						t.Errorf("GetChartPathsFromConfig() returned unexpected path: %q", path)
					}
				}
				// Check all expected paths were found
				for path, found := range expectedPaths {
					if !found {
						t.Errorf("GetChartPathsFromConfig() missing expected path: %q", path)
					}
				}
			}
		})
	}
}

func TestGetChartPathsFromConfig_InvalidChartsInGlob(t *testing.T) {
	// Create directory with mix of valid and invalid charts
	tmpDir := t.TempDir()
	chartsDir := filepath.Join(tmpDir, "charts")

	// Valid chart
	validChartDir := filepath.Join(chartsDir, "valid-chart")
	if err := os.MkdirAll(validChartDir, 0755); err != nil {
		t.Fatal(err)
	}
	chartYaml := filepath.Join(validChartDir, "Chart.yaml")
	if err := os.WriteFile(chartYaml, []byte("name: test\nversion: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Invalid chart (no Chart.yaml)
	invalidChartDir := filepath.Join(chartsDir, "invalid-chart")
	if err := os.MkdirAll(invalidChartDir, 0755); err != nil {
		t.Fatal(err)
	}

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: filepath.Join(chartsDir, "*")},
		},
	}

	// With the new behavior, we filter to only valid charts instead of failing
	paths, err := GetChartPathsFromConfig(config)
	if err != nil {
		t.Errorf("GetChartPathsFromConfig() should succeed and filter to valid charts, got error: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("GetChartPathsFromConfig() returned %d paths, want 1 (only the valid chart)", len(paths))
	}
	if len(paths) > 0 && paths[0] != validChartDir {
		t.Errorf("GetChartPathsFromConfig() returned path %s, want %s", paths[0], validChartDir)
	}
}

func TestGetChartPathsFromConfig_MultipleCharts(t *testing.T) {
	// Create multiple valid charts
	tmpDir := t.TempDir()
	chart1Dir := filepath.Join(tmpDir, "chart1")
	chart2Dir := filepath.Join(tmpDir, "chart2")
	chart3Dir := filepath.Join(tmpDir, "chart3")

	for _, dir := range []string{chart1Dir, chart2Dir, chart3Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		chartYaml := filepath.Join(dir, "Chart.yaml")
		if err := os.WriteFile(chartYaml, []byte("name: test\nversion: 1.0.0\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := &tools.Config{
		Charts: []tools.ChartConfig{
			{Path: chart1Dir},
			{Path: chart2Dir},
			{Path: chart3Dir},
		},
	}

	paths, err := GetChartPathsFromConfig(config)
	if err != nil {
		t.Fatalf("GetChartPathsFromConfig() unexpected error = %v", err)
	}
	if len(paths) != 3 {
		t.Errorf("GetChartPathsFromConfig() returned %d paths, want 3", len(paths))
	}

	// Verify all paths are present
	expectedPaths := map[string]bool{
		chart1Dir: false,
		chart2Dir: false,
		chart3Dir: false,
	}
	for _, path := range paths {
		if _, ok := expectedPaths[path]; ok {
			expectedPaths[path] = true
		}
	}
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected path %s not found in results", path)
		}
	}
}

func TestValidateChartPath(t *testing.T) {
	// Create a temporary valid chart directory
	tmpDir := t.TempDir()
	validChartDir := filepath.Join(tmpDir, "valid-chart")
	if err := os.MkdirAll(validChartDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create Chart.yaml
	chartYaml := filepath.Join(validChartDir, "Chart.yaml")
	if err := os.WriteFile(chartYaml, []byte("name: test\nversion: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an invalid chart directory (no Chart.yaml)
	invalidChartDir := filepath.Join(tmpDir, "invalid-chart")
	if err := os.MkdirAll(invalidChartDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file (not a directory)
	notADir := filepath.Join(tmpDir, "not-a-dir.txt")
	if err := os.WriteFile(notADir, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid chart directory",
			path:    validChartDir,
			wantErr: false,
		},
		{
			name:    "non-existent path",
			path:    filepath.Join(tmpDir, "does-not-exist"),
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name:    "path is not a directory",
			path:    notADir,
			wantErr: true,
			errMsg:  "not a directory",
		},
		{
			name:    "directory without Chart.yaml",
			path:    invalidChartDir,
			wantErr: true,
			errMsg:  "Chart.yaml or Chart.yml not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChartPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateChartPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateChartPath() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateChartPath_WithChartYml(t *testing.T) {
	// Test that Chart.yml (alternative spelling) is also accepted
	tmpDir := t.TempDir()
	chartDir := filepath.Join(tmpDir, "chart-with-yml")
	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create Chart.yml (not Chart.yaml)
	chartYml := filepath.Join(chartDir, "Chart.yml")
	if err := os.WriteFile(chartYml, []byte("name: test\nversion: 1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := validateChartPath(chartDir)
	if err != nil {
		t.Errorf("validateChartPath() with Chart.yml should succeed, got error: %v", err)
	}
}

func TestContainsGlob(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"./charts/*", true},
		{"./charts/**/*.yaml", true},
		{"./charts/[abc]", true},
		{"./charts/foo?bar", true},
		{"./charts/simple", false},
		{"./charts/simple-path", false},
		{"simple", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := containsGlob(tt.path)
			if got != tt.want {
				t.Errorf("containsGlob(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetPreflightPathsFromConfig(t *testing.T) {
	// Create a test preflight spec file
	tmpDir := t.TempDir()
	validPreflightSpec := filepath.Join(tmpDir, "preflight.yaml")
	preflightContent := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test
spec:
  collectors: []
`
	if err := os.WriteFile(validPreflightSpec, []byte(preflightContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		config    *tools.Config
		wantPaths []string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "no preflights in config",
			config: &tools.Config{
				Preflights: []tools.PreflightConfig{},
			},
			wantErr: true,
			errMsg:  "no preflights found",
		},
		{
			name: "single preflight path",
			config: &tools.Config{
				Preflights: []tools.PreflightConfig{
					{Path: validPreflightSpec},
				},
			},
			wantPaths: []string{validPreflightSpec},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := GetPreflightPathsFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPreflightPathsFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetPreflightPathsFromConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if !tt.wantErr {
				if len(paths) != len(tt.wantPaths) {
					t.Errorf("GetPreflightPathsFromConfig() returned %d paths, want %d", len(paths), len(tt.wantPaths))
					return
				}
				for i, path := range paths {
					if path != tt.wantPaths[i] {
						t.Errorf("GetPreflightPathsFromConfig() path[%d] = %q, want %q", i, path, tt.wantPaths[i])
					}
				}
			}
		})
	}
}

func TestGetPreflightPathsFromConfig_GlobExpansion(t *testing.T) {
	// Create test directory structure with multiple preflight specs
	tmpDir := t.TempDir()

	// Create preflights directory with multiple specs
	preflightsDir := filepath.Join(tmpDir, "preflights")
	if err := os.MkdirAll(preflightsDir, 0755); err != nil {
		t.Fatal(err)
	}

	preflight1 := filepath.Join(preflightsDir, "preflight1.yaml")
	preflight2 := filepath.Join(preflightsDir, "preflight2.yaml")
	preflight3 := filepath.Join(tmpDir, "standalone-preflight.yaml")

	preflightContent := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test
spec:
  collectors: []
`

	for _, file := range []string{preflight1, preflight2, preflight3} {
		if err := os.WriteFile(file, []byte(preflightContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		config    *tools.Config
		wantPaths []string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "glob pattern expansion",
			config: &tools.Config{
				Preflights: []tools.PreflightConfig{
					{Path: filepath.Join(preflightsDir, "*.yaml")},
				},
			},
			wantPaths: []string{preflight1, preflight2},
			wantErr:   false,
		},
		{
			name: "multiple preflights - mixed glob and direct",
			config: &tools.Config{
				Preflights: []tools.PreflightConfig{
					{Path: filepath.Join(preflightsDir, "*.yaml")},
					{Path: preflight3},
				},
			},
			wantPaths: []string{preflight1, preflight2, preflight3},
			wantErr:   false,
		},
		{
			name: "glob with no matches",
			config: &tools.Config{
				Preflights: []tools.PreflightConfig{
					{Path: filepath.Join(tmpDir, "nonexistent", "*.yaml")},
				},
			},
			wantErr: true,
			errMsg:  "no preflight specs found matching pattern",
		},
		{
			name: "glob pattern with prefix",
			config: &tools.Config{
				Preflights: []tools.PreflightConfig{
					{Path: filepath.Join(preflightsDir, "preflight*.yaml")},
				},
			},
			wantPaths: []string{preflight1, preflight2},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := GetPreflightPathsFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPreflightPathsFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetPreflightPathsFromConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if !tt.wantErr {
				if len(paths) != len(tt.wantPaths) {
					t.Errorf("GetPreflightPathsFromConfig() returned %d paths, want %d", len(paths), len(tt.wantPaths))
					return
				}
				// Build map of expected paths for order-independent comparison
				expectedPaths := make(map[string]bool)
				for _, p := range tt.wantPaths {
					expectedPaths[p] = false
				}
				// Mark found paths
				for _, path := range paths {
					if _, ok := expectedPaths[path]; ok {
						expectedPaths[path] = true
					} else {
						t.Errorf("GetPreflightPathsFromConfig() returned unexpected path: %q", path)
					}
				}
				// Check all expected paths were found
				for path, found := range expectedPaths {
					if !found {
						t.Errorf("GetPreflightPathsFromConfig() missing expected path: %q", path)
					}
				}
			}
		})
	}
}

func TestGetPreflightPathsFromConfig_InvalidPreflightsInGlob(t *testing.T) {
	// Create directory with mix of valid and invalid preflight specs
	tmpDir := t.TempDir()
	preflightsDir := filepath.Join(tmpDir, "preflights")
	if err := os.MkdirAll(preflightsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Valid preflight spec
	validPreflight := filepath.Join(preflightsDir, "valid.yaml")
	preflightContent := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test
spec:
  collectors: []
`
	if err := os.WriteFile(validPreflight, []byte(preflightContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Invalid preflight spec (non-existent file that glob might match)
	// For preflight, we test the case where one of the matched files doesn't exist
	invalidPreflight := filepath.Join(preflightsDir, "nonexistent.yaml")

	config := &tools.Config{
		Preflights: []tools.PreflightConfig{
			{Path: validPreflight},
			{Path: invalidPreflight},
		},
	}

	_, err := GetPreflightPathsFromConfig(config)
	if err == nil {
		t.Error("GetPreflightPathsFromConfig() should fail when spec file doesn't exist, got nil error")
	}
	if !contains(err.Error(), "does not exist") {
		t.Errorf("GetPreflightPathsFromConfig() error = %v, want error about file not existing", err)
	}
}

func TestGetPreflightPathsFromConfig_MultiplePreflights(t *testing.T) {
	// Create multiple valid preflight specs
	tmpDir := t.TempDir()
	preflight1 := filepath.Join(tmpDir, "preflight1.yaml")
	preflight2 := filepath.Join(tmpDir, "preflight2.yaml")
	preflight3 := filepath.Join(tmpDir, "preflight3.yaml")

	preflightContent := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test
spec:
  collectors: []
`

	for _, file := range []string{preflight1, preflight2, preflight3} {
		if err := os.WriteFile(file, []byte(preflightContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := &tools.Config{
		Preflights: []tools.PreflightConfig{
			{Path: preflight1},
			{Path: preflight2},
			{Path: preflight3},
		},
	}

	paths, err := GetPreflightPathsFromConfig(config)
	if err != nil {
		t.Fatalf("GetPreflightPathsFromConfig() unexpected error = %v", err)
	}
	if len(paths) != 3 {
		t.Errorf("GetPreflightPathsFromConfig() returned %d paths, want 3", len(paths))
	}

	// Verify all paths are present
	expectedPaths := map[string]bool{
		preflight1: false,
		preflight2: false,
		preflight3: false,
	}
	for _, path := range paths {
		if _, ok := expectedPaths[path]; ok {
			expectedPaths[path] = true
		}
	}
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected path %s not found in results", path)
		}
	}
}

func TestValidatePreflightPath(t *testing.T) {
	// Create a temporary valid preflight spec file
	tmpDir := t.TempDir()
	validPreflight := filepath.Join(tmpDir, "valid-preflight.yaml")
	preflightContent := `apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: test
spec:
  collectors: []
`
	if err := os.WriteFile(validPreflight, []byte(preflightContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a directory (not a file)
	notAFile := filepath.Join(tmpDir, "not-a-file")
	if err := os.MkdirAll(notAFile, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid preflight spec file",
			path:    validPreflight,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    filepath.Join(tmpDir, "does-not-exist.yaml"),
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name:    "path is a directory",
			path:    notAFile,
			wantErr: true,
			errMsg:  "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePreflightPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePreflightPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("validatePreflightPath() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

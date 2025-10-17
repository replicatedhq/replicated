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
			errMsg:  "no charts found matching pattern",
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

	_, err := GetChartPathsFromConfig(config)
	if err == nil {
		t.Error("GetChartPathsFromConfig() should fail when glob matches invalid chart, got nil error")
	}
	if !contains(err.Error(), "Chart.yaml or Chart.yml not found") {
		t.Errorf("GetChartPathsFromConfig() error = %v, want error about Chart.yaml not found", err)
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

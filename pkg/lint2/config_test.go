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
		name    string
		config  *tools.Config
		wantErr bool
		errMsg  string
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
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetChartPathsFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChartPathsFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetChartPathsFromConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
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

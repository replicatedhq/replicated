package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigParser_ParseConfig(t *testing.T) {
	parser := NewConfigParser()

	tests := []struct {
		name        string
		fixture     string
		wantErr     bool
		checkConfig func(*testing.T, *Config)
	}{
		{
			name:    "valid YAML with all fields",
			fixture: "valid-full.yaml",
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.ReplLint.Version != 1 {
					t.Errorf("version = %d, want 1", cfg.ReplLint.Version)
				}
				if !cfg.ReplLint.Enabled {
					t.Error("enabled = false, want true")
				}
				if !cfg.ReplLint.Linters.Helm.IsEnabled() {
					t.Error("helm is disabled, want enabled")
				}
				if cfg.ReplLint.Linters.Helm.Disabled {
					t.Error("helm.disabled = true, want false")
				}
				if !cfg.ReplLint.Linters.Preflight.Strict {
					t.Error("preflight.strict = false, want true")
				}
				if cfg.ReplLint.Tools[ToolHelm] != "3.14.4" {
					t.Errorf("helm version = %q, want 3.14.4", cfg.ReplLint.Tools[ToolHelm])
				}
			},
		},
		{
			name:    "valid JSON with all fields",
			fixture: "valid-full.json",
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.ReplLint.Version != 1 {
					t.Errorf("version = %d, want 1", cfg.ReplLint.Version)
				}
				if cfg.ReplLint.Tools[ToolHelm] != "3.14.4" {
					t.Errorf("helm version = %q, want 3.14.4", cfg.ReplLint.Tools[ToolHelm])
				}
			},
		},
		{
			name:    "minimal config with defaults",
			fixture: "minimal.yaml",
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				// Version should default to 1
				if cfg.ReplLint.Version != 1 {
					t.Errorf("version = %d, want 1 (default)", cfg.ReplLint.Version)
				}
				// Tools should be populated with defaults
				if cfg.ReplLint.Tools[ToolHelm] != DefaultHelmVersion {
					t.Errorf("helm version = %q, want default %q", cfg.ReplLint.Tools[ToolHelm], DefaultHelmVersion)
				}
				if cfg.ReplLint.Tools[ToolPreflight] != DefaultPreflightVersion {
					t.Errorf("preflight version = %q, want default %q", cfg.ReplLint.Tools[ToolPreflight], DefaultPreflightVersion)
				}
				if cfg.ReplLint.Tools[ToolSupportBundle] != DefaultSupportBundleVersion {
					t.Errorf("support-bundle version = %q, want default %q", cfg.ReplLint.Tools[ToolSupportBundle], DefaultSupportBundleVersion)
				}
			},
		},
		{
			name:    "invalid version string",
			fixture: "invalid-version.yaml",
			wantErr: true,
		},
		{
			name:    "malformed YAML",
			fixture: "malformed.yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.fixture)
			// Use FindAndParseConfig with file path to get defaults applied
			config, err := parser.FindAndParseConfig(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindAndParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkConfig != nil {
				tt.checkConfig(t, config)
			}
		})
	}
}

func TestConfigParser_DefaultConfig(t *testing.T) {
	parser := NewConfigParser()
	config := parser.DefaultConfig()

	if config.ReplLint.Version != 1 {
		t.Errorf("version = %d, want 1", config.ReplLint.Version)
	}

	if !config.ReplLint.Enabled {
		t.Error("enabled should default to true")
	}

	// Check default tool versions
	if config.ReplLint.Tools[ToolHelm] != DefaultHelmVersion {
		t.Errorf("helm version = %q, want %q", config.ReplLint.Tools[ToolHelm], DefaultHelmVersion)
	}
	if config.ReplLint.Tools[ToolPreflight] != DefaultPreflightVersion {
		t.Errorf("preflight version = %q, want %q", config.ReplLint.Tools[ToolPreflight], DefaultPreflightVersion)
	}
	if config.ReplLint.Tools[ToolSupportBundle] != DefaultSupportBundleVersion {
		t.Errorf("support-bundle version = %q, want %q", config.ReplLint.Tools[ToolSupportBundle], DefaultSupportBundleVersion)
	}
}

func TestConfigParser_FindAndParseConfig(t *testing.T) {
	parser := NewConfigParser()

	// Test with direct file path
	t.Run("direct file path", func(t *testing.T) {
		path := filepath.Join("testdata", "valid-full.yaml")
		config, err := parser.FindAndParseConfig(path)
		if err != nil {
			t.Fatalf("FindAndParseConfig() error = %v", err)
		}
		if config.ReplLint.Tools[ToolHelm] != "3.14.4" {
			t.Errorf("helm version = %q, want 3.14.4", config.ReplLint.Tools[ToolHelm])
		}
	})

	// Test with directory containing .replicated
	t.Run("directory walk up", func(t *testing.T) {
		// Create a temporary directory structure
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")
		subDir := filepath.Join(tmpDir, "subdir", "nested")

		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("creating test dirs: %v", err)
		}

		// Write a config file at the root
		configData := []byte(`repl-lint:
  enabled: true
  tools:
    helm: "3.14.4"
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		// Try to find config from nested subdirectory
		config, err := parser.FindAndParseConfig(subDir)
		if err != nil {
			t.Fatalf("FindAndParseConfig() error = %v", err)
		}
		if config.ReplLint.Tools[ToolHelm] != "3.14.4" {
			t.Errorf("helm version = %q, want 3.14.4", config.ReplLint.Tools[ToolHelm])
		}
	})

	// Test when no config found (should return error)
	t.Run("no config found returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := parser.FindAndParseConfig(tmpDir)
		if err == nil {
			t.Error("FindAndParseConfig() expected error when no config found, got nil")
		}
		if !strings.Contains(err.Error(), "no .replicated config file found") {
			t.Errorf("Expected 'no .replicated config file found' error, got: %v", err)
		}
	})
}

func TestGetToolVersions(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   map[string]string
	}{
		{
			name: "valid config",
			config: &Config{
				ReplLint: &ReplLintConfig{
					Tools: map[string]string{
						"helm":      "3.14.4",
						"preflight": "0.123.9",
					},
				},
			},
			want: map[string]string{
				"helm":      "3.14.4",
				"preflight": "0.123.9",
			},
		},
		{
			name:   "nil config",
			config: nil,
			want:   map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetToolVersions(tt.config)
			if len(got) != len(tt.want) {
				t.Errorf("GetToolVersions() returned %d items, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("GetToolVersions()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.2.3", true},
		{"v1.2.3", true},
		{"0.0.0", true},
		{"3.14.4", true},
		{"0.123.9", true},
		{"1.2.3-beta", true},
		{"1.2.3-alpha.1", true},
		{"1.2.3+build.123", true},
		{"1.2.3-beta+build", true},
		{"not-a-version", false},
		{"1.2", false},
		{"1", false},
		{"1.2.3.4", false},
		{"", false},
		{"v", false},
		{"latest", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := isValidSemver(tt.version)
			if got != tt.want {
				t.Errorf("isValidSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestConfigParser_ResolveChartPaths(t *testing.T) {
	parser := NewConfigParser()

	t.Run("relative paths resolved to absolute", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		// Write a config file with relative chart paths
		configData := []byte(`charts:
  - path: ./charts/chart1
  - path: ./charts/chart2
  - path: charts/chart3
repl-lint:
  enabled: true
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		// Parse the config
		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		// Verify all chart paths are absolute and relative to config file directory
		if len(config.Charts) != 3 {
			t.Fatalf("expected 3 charts, got %d", len(config.Charts))
		}

		expectedPaths := []string{
			filepath.Join(tmpDir, "charts/chart1"),
			filepath.Join(tmpDir, "charts/chart2"),
			filepath.Join(tmpDir, "charts/chart3"),
		}

		for i, chart := range config.Charts {
			if !filepath.IsAbs(chart.Path) {
				t.Errorf("chart[%d].Path = %q, expected absolute path", i, chart.Path)
			}
			if chart.Path != expectedPaths[i] {
				t.Errorf("chart[%d].Path = %q, want %q", i, chart.Path, expectedPaths[i])
			}
		}
	})

	t.Run("absolute paths preserved", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		absolutePath := "/absolute/path/to/chart"
		configData := []byte(`charts:
  - path: ` + absolutePath + `
repl-lint:
  enabled: true
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		// Parse the config
		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		// Verify absolute path is preserved
		if len(config.Charts) != 1 {
			t.Fatalf("expected 1 chart, got %d", len(config.Charts))
		}

		if config.Charts[0].Path != absolutePath {
			t.Errorf("chart.Path = %q, want %q (absolute path should be preserved)", config.Charts[0].Path, absolutePath)
		}
	})

	t.Run("mixed relative and absolute paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "subdir", ".replicated")

		// Create subdirectory
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			t.Fatalf("creating test dirs: %v", err)
		}

		absolutePath := "/absolute/path/to/chart"
		configData := []byte(`charts:
  - path: ./charts/relative-chart
  - path: ` + absolutePath + `
  - path: ../parent-chart
repl-lint:
  enabled: true
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		// Parse the config
		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		// Verify paths
		if len(config.Charts) != 3 {
			t.Fatalf("expected 3 charts, got %d", len(config.Charts))
		}

		configDir := filepath.Dir(configPath)
		expectedPaths := []string{
			filepath.Join(configDir, "charts/relative-chart"),
			absolutePath, // preserved
			filepath.Join(configDir, "../parent-chart"),
		}

		for i, chart := range config.Charts {
			if !filepath.IsAbs(chart.Path) {
				t.Errorf("chart[%d].Path = %q, expected absolute path", i, chart.Path)
			}
			if chart.Path != expectedPaths[i] {
				t.Errorf("chart[%d].Path = %q, want %q", i, chart.Path, expectedPaths[i])
			}
		}
	})

	t.Run("no charts in config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`repl-lint:
  enabled: true
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		// Parse the config - should not error even with no charts
		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		if len(config.Charts) != 0 {
			t.Errorf("expected 0 charts, got %d", len(config.Charts))
		}
	})
}

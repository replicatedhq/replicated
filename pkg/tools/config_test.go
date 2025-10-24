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
				if !cfg.ReplLint.Linters.Helm.IsEnabled() {
					t.Error("helm is disabled, want enabled")
				}
				if cfg.ReplLint.Linters.Helm.Disabled != nil && *cfg.ReplLint.Linters.Helm.Disabled {
					t.Error("helm.disabled = true, want false")
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
				// Tools should be populated with "latest" as defaults
				if cfg.ReplLint.Tools[ToolHelm] != "latest" {
					t.Errorf("helm version = %q, want %q", cfg.ReplLint.Tools[ToolHelm], "latest")
				}
				if cfg.ReplLint.Tools[ToolPreflight] != "latest" {
					t.Errorf("preflight version = %q, want %q", cfg.ReplLint.Tools[ToolPreflight], "latest")
				}
				if cfg.ReplLint.Tools[ToolSupportBundle] != "latest" {
					t.Errorf("support-bundle version = %q, want %q", cfg.ReplLint.Tools[ToolSupportBundle], "latest")
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

	// Check default tool versions - now defaults to "latest" which resolves at runtime
	if config.ReplLint.Tools[ToolHelm] != "latest" {
		t.Errorf("helm version = %q, want %q", config.ReplLint.Tools[ToolHelm], "latest")
	}
	if config.ReplLint.Tools[ToolPreflight] != "latest" {
		t.Errorf("preflight version = %q, want %q", config.ReplLint.Tools[ToolPreflight], "latest")
	}
	if config.ReplLint.Tools[ToolSupportBundle] != "latest" {
		t.Errorf("support-bundle version = %q, want %q", config.ReplLint.Tools[ToolSupportBundle], "latest")
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

	// Test when no config found (should return default config for auto-discovery mode)
	t.Run("no config found returns default config", func(t *testing.T) {
		tmpDir := t.TempDir()
		config, err := parser.FindAndParseConfig(tmpDir)
		if err != nil {
			t.Errorf("FindAndParseConfig() unexpected error = %v", err)
		}
		if config == nil {
			t.Error("FindAndParseConfig() returned nil config, expected default config")
		}
		// Verify it's a valid default config
		if config.ReplLint == nil {
			t.Error("Default config should have ReplLint section")
		}
		if config.ReplLint.Version != 1 {
			t.Errorf("Default config version = %d, want 1", config.ReplLint.Version)
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

func TestConfigParser_MergeConfigs(t *testing.T) {
	parser := NewConfigParser()

	t.Run("scalar fields override", func(t *testing.T) {
		parent := &Config{
			AppId:        "parent-app",
			AppSlug:      "parent-slug",
			ReleaseLabel: "parent-label",
		}
		child := &Config{
			AppId:        "child-app",
			AppSlug:      "child-slug",
			ReleaseLabel: "child-label",
		}

		merged := parser.mergeConfigs([]*Config{parent, child})

		if merged.AppId != "child-app" {
			t.Errorf("AppId = %q, want %q", merged.AppId, "child-app")
		}
		if merged.AppSlug != "child-slug" {
			t.Errorf("AppSlug = %q, want %q", merged.AppSlug, "child-slug")
		}
		if merged.ReleaseLabel != "child-label" {
			t.Errorf("ReleaseLabel = %q, want %q", merged.ReleaseLabel, "child-label")
		}
	})

	t.Run("channel arrays override", func(t *testing.T) {
		parent := &Config{
			PromoteToChannelIds:   []string{"channel-1", "channel-2"},
			PromoteToChannelNames: []string{"stable", "beta"},
		}
		child := &Config{
			PromoteToChannelIds:   []string{"channel-3"},
			PromoteToChannelNames: []string{"alpha"},
		}

		merged := parser.mergeConfigs([]*Config{parent, child})

		if len(merged.PromoteToChannelIds) != 1 || merged.PromoteToChannelIds[0] != "channel-3" {
			t.Errorf("PromoteToChannelIds = %v, want [channel-3]", merged.PromoteToChannelIds)
		}
		if len(merged.PromoteToChannelNames) != 1 || merged.PromoteToChannelNames[0] != "alpha" {
			t.Errorf("PromoteToChannelNames = %v, want [alpha]", merged.PromoteToChannelNames)
		}
	})

	t.Run("charts append", func(t *testing.T) {
		parent := &Config{
			Charts: []ChartConfig{
				{Path: "/parent/chart1"},
				{Path: "/parent/chart2"},
			},
		}
		child := &Config{
			Charts: []ChartConfig{
				{Path: "/child/chart3"},
			},
		}

		merged := parser.mergeConfigs([]*Config{parent, child})

		if len(merged.Charts) != 3 {
			t.Fatalf("len(Charts) = %d, want 3", len(merged.Charts))
		}
		if merged.Charts[0].Path != "/parent/chart1" {
			t.Errorf("Charts[0].Path = %q, want %q", merged.Charts[0].Path, "/parent/chart1")
		}
		if merged.Charts[1].Path != "/parent/chart2" {
			t.Errorf("Charts[1].Path = %q, want %q", merged.Charts[1].Path, "/parent/chart2")
		}
		if merged.Charts[2].Path != "/child/chart3" {
			t.Errorf("Charts[2].Path = %q, want %q", merged.Charts[2].Path, "/child/chart3")
		}
	})

	t.Run("preflights append", func(t *testing.T) {
		parent := &Config{
			Preflights: []PreflightConfig{
				{Path: "/parent/preflight1"},
			},
		}
		child := &Config{
			Preflights: []PreflightConfig{
				{Path: "/child/preflight2"},
			},
		}

		merged := parser.mergeConfigs([]*Config{parent, child})

		if len(merged.Preflights) != 2 {
			t.Fatalf("len(Preflights) = %d, want 2", len(merged.Preflights))
		}
		if merged.Preflights[0].Path != "/parent/preflight1" {
			t.Errorf("Preflights[0].Path = %q, want %q", merged.Preflights[0].Path, "/parent/preflight1")
		}
		if merged.Preflights[1].Path != "/child/preflight2" {
			t.Errorf("Preflights[1].Path = %q, want %q", merged.Preflights[1].Path, "/child/preflight2")
		}
	})

	t.Run("manifests append", func(t *testing.T) {
		parent := &Config{
			Manifests: []string{"/parent/**/*.yaml"},
		}
		child := &Config{
			Manifests: []string{"/child/**/*.yaml"},
		}

		merged := parser.mergeConfigs([]*Config{parent, child})

		if len(merged.Manifests) != 2 {
			t.Fatalf("len(Manifests) = %d, want 2", len(merged.Manifests))
		}
		if merged.Manifests[0] != "/parent/**/*.yaml" {
			t.Errorf("Manifests[0] = %q, want %q", merged.Manifests[0], "/parent/**/*.yaml")
		}
		if merged.Manifests[1] != "/child/**/*.yaml" {
			t.Errorf("Manifests[1] = %q, want %q", merged.Manifests[1], "/child/**/*.yaml")
		}
	})

	t.Run("empty values dont override", func(t *testing.T) {
		parent := &Config{
			AppId:                 "parent-app",
			PromoteToChannelIds:   []string{"channel-1"},
			PromoteToChannelNames: []string{"stable"},
		}
		child := &Config{
			AppId:                 "", // Empty - should not override
			PromoteToChannelIds:   nil, // Nil - should not override
			PromoteToChannelNames: []string{}, // Empty slice - should not override
		}

		merged := parser.mergeConfigs([]*Config{parent, child})

		if merged.AppId != "parent-app" {
			t.Errorf("AppId = %q, want %q (empty child should not override)", merged.AppId, "parent-app")
		}
		if len(merged.PromoteToChannelIds) != 1 || merged.PromoteToChannelIds[0] != "channel-1" {
			t.Errorf("PromoteToChannelIds = %v, want [channel-1] (nil child should not override)", merged.PromoteToChannelIds)
		}
		if len(merged.PromoteToChannelNames) != 1 || merged.PromoteToChannelNames[0] != "stable" {
			t.Errorf("PromoteToChannelNames = %v, want [stable] (empty child should not override)", merged.PromoteToChannelNames)
		}
	})

	t.Run("three level merge", func(t *testing.T) {
		grandparent := &Config{
			AppId: "grandparent-app",
			Charts: []ChartConfig{
				{Path: "/gp/chart"},
			},
		}
		parent := &Config{
			AppSlug: "parent-slug",
			Charts: []ChartConfig{
				{Path: "/parent/chart"},
			},
		}
		child := &Config{
			ReleaseLabel: "child-label",
			Charts: []ChartConfig{
				{Path: "/child/chart"},
			},
		}

		merged := parser.mergeConfigs([]*Config{grandparent, parent, child})

		// Scalars - last non-empty wins
		if merged.AppId != "grandparent-app" {
			t.Errorf("AppId = %q, want %q", merged.AppId, "grandparent-app")
		}
		if merged.AppSlug != "parent-slug" {
			t.Errorf("AppSlug = %q, want %q", merged.AppSlug, "parent-slug")
		}
		if merged.ReleaseLabel != "child-label" {
			t.Errorf("ReleaseLabel = %q, want %q", merged.ReleaseLabel, "child-label")
		}

		// Charts - all accumulated
		if len(merged.Charts) != 3 {
			t.Fatalf("len(Charts) = %d, want 3", len(merged.Charts))
		}
	})

	t.Run("repl-lint merge preserved", func(t *testing.T) {
		// Helper to create bool pointers
		boolPtr := func(b bool) *bool { return &b }

		parent := &Config{
			ReplLint: &ReplLintConfig{
				Version: 1,
				Linters: LintersConfig{
					Helm: LinterConfig{Disabled: boolPtr(false)},
				},
				Tools: map[string]string{
					"helm": "3.14.4",
				},
			},
		}
		child := &Config{
			ReplLint: &ReplLintConfig{
				Linters: LintersConfig{
					Helm: LinterConfig{Disabled: boolPtr(true)},
				},
				Tools: map[string]string{
					"helm": "3.19.0",
				},
			},
		}

		merged := parser.mergeConfigs([]*Config{parent, child})

		if merged.ReplLint == nil {
			t.Fatal("ReplLint is nil")
		}
		// Verify child's disabled setting overrides parent
		if merged.ReplLint.Linters.Helm.Disabled == nil || !*merged.ReplLint.Linters.Helm.Disabled {
			t.Errorf("Helm.Disabled = %v, want true (child overrides parent)", merged.ReplLint.Linters.Helm.Disabled)
		}
		if merged.ReplLint.Tools["helm"] != "3.19.0" {
			t.Errorf("Tools[helm] = %q, want %q", merged.ReplLint.Tools["helm"], "3.19.0")
		}
	})
}

func TestConfigParser_ParseFullConfig(t *testing.T) {
	parser := NewConfigParser()

	t.Run("parse config with all fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`appId: "app-123"
appSlug: "my-app"
promoteToChannelIds: ["channel-1", "channel-2"]
promoteToChannelNames: ["stable", "beta"]
charts:
  - path: ./charts/chart1
    chartVersion: "1.0.0"
    appVersion: "1.0.0"
  - path: ./charts/chart2
preflights:
  - path: ./preflights/check1
    chartName: "chart1"
    chartVersion: "1.0.0"
  - path: ./preflights/check2
    chartName: "chart2"
    chartVersion: "1.0.0"
releaseLabel: "v{{.Version}}"
manifests:
  - "replicated/**/*.yaml"
  - "manifests/**/*.yaml"
repl-lint:
  linters:
    helm:
      disabled: false
  tools:
    helm: "3.14.4"
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		// Parse the config
		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		// Verify all fields are populated
		if config.AppId != "app-123" {
			t.Errorf("AppId = %q, want %q", config.AppId, "app-123")
		}
		if config.AppSlug != "my-app" {
			t.Errorf("AppSlug = %q, want %q", config.AppSlug, "my-app")
		}
		if len(config.PromoteToChannelIds) != 2 {
			t.Errorf("len(PromoteToChannelIds) = %d, want 2", len(config.PromoteToChannelIds))
		}
		if len(config.PromoteToChannelNames) != 2 {
			t.Errorf("len(PromoteToChannelNames) = %d, want 2", len(config.PromoteToChannelNames))
		}
		if len(config.Charts) != 2 {
			t.Errorf("len(Charts) = %d, want 2", len(config.Charts))
		}
		if len(config.Preflights) != 2 {
			t.Errorf("len(Preflights) = %d, want 2", len(config.Preflights))
		}
		if config.ReleaseLabel != "v{{.Version}}" {
			t.Errorf("ReleaseLabel = %q, want %q", config.ReleaseLabel, "v{{.Version}}")
		}
		if len(config.Manifests) != 2 {
			t.Errorf("len(Manifests) = %d, want 2", len(config.Manifests))
		}
		if config.ReplLint == nil {
			t.Fatal("ReplLint is nil")
		}
	})

	t.Run("parse config with missing fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		// Minimal config with only repl-lint
		configData := []byte(`repl-lint:
  version: 1
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		// Parse the config - should not error
		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		// Verify empty fields are empty, not nil
		if config.AppId != "" {
			t.Errorf("AppId should be empty, got %q", config.AppId)
		}
		if config.PromoteToChannelIds != nil {
			t.Errorf("PromoteToChannelIds should be nil, got %v", config.PromoteToChannelIds)
		}
		if config.Charts != nil {
			t.Errorf("Charts should be nil, got %v", config.Charts)
		}
	})
}

func TestConfigParser_ResolvePaths(t *testing.T) {
	parser := NewConfigParser()

	t.Run("relative chart paths resolved to absolute", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		// Write a config file with relative chart paths
		configData := []byte(`charts:
  - path: ./charts/chart1
  - path: ./charts/chart2
  - path: charts/chart3
repl-lint:
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

	t.Run("relative preflight paths resolved to absolute", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`preflights:
  - path: ./preflights/check1
    chartName: chart1
    chartVersion: "1.0.0"
  - path: preflights/check2
    chartName: chart2
    chartVersion: "2.0.0"
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		if len(config.Preflights) != 2 {
			t.Fatalf("expected 2 preflights, got %d", len(config.Preflights))
		}

		// Check first preflight - path is resolved, chartName/Version are not
		expectedPath := filepath.Join(tmpDir, "preflights/check1")
		if config.Preflights[0].Path != expectedPath {
			t.Errorf("preflights[0].Path = %q, want %q", config.Preflights[0].Path, expectedPath)
		}
		if config.Preflights[0].ChartName != "chart1" {
			t.Errorf("preflights[0].ChartName = %q, want %q", config.Preflights[0].ChartName, "chart1")
		}
		if config.Preflights[0].ChartVersion != "1.0.0" {
			t.Errorf("preflights[0].ChartVersion = %q, want %q", config.Preflights[0].ChartVersion, "1.0.0")
		}

		// Check second preflight
		expectedPath2 := filepath.Join(tmpDir, "preflights/check2")
		if config.Preflights[1].Path != expectedPath2 {
			t.Errorf("preflights[1].Path = %q, want %q", config.Preflights[1].Path, expectedPath2)
		}
		if config.Preflights[1].ChartName != "chart2" {
			t.Errorf("preflights[1].ChartName = %q, want %q", config.Preflights[1].ChartName, "chart2")
		}
		if config.Preflights[1].ChartVersion != "2.0.0" {
			t.Errorf("preflights[1].ChartVersion = %q, want %q", config.Preflights[1].ChartVersion, "2.0.0")
		}
	})

	t.Run("relative manifest paths resolved to absolute", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`manifests:
  - "replicated/**/*.yaml"
  - "./manifests/**/*.yaml"
  - "other/*.yaml"
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		config, err := parser.ParseConfigFile(configPath)
		if err != nil {
			t.Fatalf("ParseConfigFile() error = %v", err)
		}

		if len(config.Manifests) != 3 {
			t.Fatalf("expected 3 manifests, got %d", len(config.Manifests))
		}

		expectedManifests := []string{
			filepath.Join(tmpDir, "replicated/**/*.yaml"),
			filepath.Join(tmpDir, "manifests/**/*.yaml"),
			filepath.Join(tmpDir, "other/*.yaml"),
		}

		for i, manifest := range config.Manifests {
			if !filepath.IsAbs(manifest) {
				t.Errorf("manifests[%d] = %q, expected absolute path", i, manifest)
			}
			if manifest != expectedManifests[i] {
				t.Errorf("manifests[%d] = %q, want %q", i, manifest, expectedManifests[i])
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

func TestConfigParser_MonorepoEndToEnd(t *testing.T) {
	// End-to-end integration test for monorepo support
	// Tests the complete flow: discovery -> parsing -> path resolution -> merging
	parser := NewConfigParser()

	// Create monorepo directory structure
	tmpDir := t.TempDir()
	rootDir := tmpDir
	appDir := filepath.Join(rootDir, "apps", "app1")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("creating directories: %v", err)
	}

	// Root config: defines common chart and org-wide settings
	rootConfigPath := filepath.Join(rootDir, ".replicated")
	rootConfigData := []byte(`charts:
  - path: ./common/lib-chart
repl-lint:
  linters:
    helm:
      disabled: false
  tools:
    helm: "3.14.4"
`)
	if err := os.WriteFile(rootConfigPath, rootConfigData, 0644); err != nil {
		t.Fatalf("writing root config: %v", err)
	}

	// App config: defines app-specific resources and metadata
	appConfigPath := filepath.Join(appDir, ".replicated")
	appConfigData := []byte(`appId: "app-123"
appSlug: "my-app"
charts:
  - path: ./chart
manifests:
  - "manifests/**/*.yaml"
repl-lint:
  tools:
    helm: "3.19.0"
`)
	if err := os.WriteFile(appConfigPath, appConfigData, 0644); err != nil {
		t.Fatalf("writing app config: %v", err)
	}

	// Parse config from app directory (should discover and merge both configs)
	config, err := parser.FindAndParseConfig(appDir)
	if err != nil {
		t.Fatalf("FindAndParseConfig() error = %v", err)
	}

	// Verify: Charts from both configs are present (accumulated)
	if len(config.Charts) != 2 {
		t.Errorf("len(Charts) = %d, want 2 (root + app charts should accumulate)", len(config.Charts))
	}

	// Verify: Both chart paths are absolute and resolved relative to their config files
	expectedRootChart := filepath.Join(rootDir, "common/lib-chart")
	expectedAppChart := filepath.Join(appDir, "chart")

	chartPaths := make(map[string]bool)
	for _, chart := range config.Charts {
		if !filepath.IsAbs(chart.Path) {
			t.Errorf("Chart path %q is not absolute", chart.Path)
		}
		chartPaths[chart.Path] = true
	}

	if !chartPaths[expectedRootChart] {
		t.Errorf("Expected root chart %q not found in merged config", expectedRootChart)
	}
	if !chartPaths[expectedAppChart] {
		t.Errorf("Expected app chart %q not found in merged config", expectedAppChart)
	}

	// Verify: Manifests from app config are present and absolute
	if len(config.Manifests) != 1 {
		t.Errorf("len(Manifests) = %d, want 1", len(config.Manifests))
	} else {
		expectedManifest := filepath.Join(appDir, "manifests/**/*.yaml")
		if config.Manifests[0] != expectedManifest {
			t.Errorf("Manifests[0] = %q, want %q", config.Manifests[0], expectedManifest)
		}
		if !filepath.IsAbs(config.Manifests[0]) {
			t.Errorf("Manifest path %q is not absolute", config.Manifests[0])
		}
	}

	// Verify: AppId from child config (override)
	if config.AppId != "app-123" {
		t.Errorf("AppId = %q, want %q (from app config)", config.AppId, "app-123")
	}

	// Verify: AppSlug from child config (override)
	if config.AppSlug != "my-app" {
		t.Errorf("AppSlug = %q, want %q (from app config)", config.AppSlug, "my-app")
	}

	// Verify: ReplLint config present and valid
	if config.ReplLint == nil {
		t.Fatal("ReplLint is nil")
	}

	// Verify: Helm disabled setting inherited from root config
	// Child config doesn't specify disabled, so should inherit parent's value
	if config.ReplLint.Linters.Helm.Disabled == nil || *config.ReplLint.Linters.Helm.Disabled {
		t.Error("Helm.Disabled should be false (inherited from root config)")
	}

	// Verify: Tool version from app config (override)
	if config.ReplLint.Tools[ToolHelm] != "3.19.0" {
		t.Errorf("Tools[helm] = %q, want %q (from app config)", config.ReplLint.Tools[ToolHelm], "3.19.0")
	}
}

func TestConfigParser_PathValidation(t *testing.T) {
	parser := NewConfigParser()

	t.Run("empty chart path rejected", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`charts:
  - path: ""
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		_, err := parser.ParseConfigFile(configPath)
		if err == nil {
			t.Error("ParseConfigFile() expected error for empty chart path, got nil")
		}
		if !strings.Contains(err.Error(), "chart[0]: path is required") {
			t.Errorf("Expected 'chart[0]: path is required' error, got: %v", err)
		}
	})

	t.Run("empty preflight path rejected", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`preflights:
  - path: ""
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		_, err := parser.ParseConfigFile(configPath)
		if err == nil {
			t.Error("ParseConfigFile() expected error for empty preflight path, got nil")
		}
		if !strings.Contains(err.Error(), "preflight[0]: path is required") {
			t.Errorf("Expected 'preflight[0]: path is required' error, got: %v", err)
		}
	})

	t.Run("missing preflight chartName rejected", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`preflights:
  - path: "./preflight.yaml"
    chartVersion: "1.0.0"
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		_, err := parser.ParseConfigFile(configPath)
		if err == nil {
			t.Error("ParseConfigFile() expected error for missing chartName, got nil")
		}
		if !strings.Contains(err.Error(), "preflight[0]: chartName is required") {
			t.Errorf("Expected 'preflight[0]: chartName is required' error, got: %v", err)
		}
	})

	t.Run("missing preflight chartVersion rejected", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`preflights:
  - path: "./preflight.yaml"
    chartName: "my-chart"
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		_, err := parser.ParseConfigFile(configPath)
		if err == nil {
			t.Error("ParseConfigFile() expected error for missing chartVersion, got nil")
		}
		if !strings.Contains(err.Error(), "preflight[0]: chartVersion is required") {
			t.Errorf("Expected 'preflight[0]: chartVersion is required' error, got: %v", err)
		}
	})

	t.Run("empty manifest path rejected", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`manifests:
  - ""
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		_, err := parser.ParseConfigFile(configPath)
		if err == nil {
			t.Error("ParseConfigFile() expected error for empty manifest path, got nil")
		}
		if !strings.Contains(err.Error(), "manifest[0]: path is required") {
			t.Errorf("Expected 'manifest[0]: path is required' error, got: %v", err)
		}
	})

	t.Run("multiple empty paths all reported", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`charts:
  - path: "./chart1"
  - path: ""
  - path: "./chart3"
repl-lint:
`)
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			t.Fatalf("writing test config: %v", err)
		}

		_, err := parser.ParseConfigFile(configPath)
		if err == nil {
			t.Error("ParseConfigFile() expected error for empty chart path, got nil")
		}
		// Should report the first empty path (index 1)
		if !strings.Contains(err.Error(), "chart[1]: path is required") {
			t.Errorf("Expected 'chart[1]: path is required' error, got: %v", err)
		}
	})
}

func TestConfigParser_Deduplication(t *testing.T) {
	parser := NewConfigParser()

	t.Run("duplicate chart paths removed", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := tmpDir
		appDir := filepath.Join(rootDir, "app")

		if err := os.MkdirAll(appDir, 0755); err != nil {
			t.Fatalf("creating directories: %v", err)
		}

		// Root config: defines a chart
		rootConfigPath := filepath.Join(rootDir, ".replicated")
		rootConfigData := []byte(`charts:
  - path: ./common/chart1
repl-lint:
`)
		if err := os.WriteFile(rootConfigPath, rootConfigData, 0644); err != nil {
			t.Fatalf("writing root config: %v", err)
		}

		// App config: references the same chart (same absolute path after resolution)
		appConfigPath := filepath.Join(appDir, ".replicated")
		appConfigData := []byte(`charts:
  - path: ../common/chart1
repl-lint:
`)
		if err := os.WriteFile(appConfigPath, appConfigData, 0644); err != nil {
			t.Fatalf("writing app config: %v", err)
		}

		// Parse from app directory - should merge and deduplicate
		config, err := parser.FindAndParseConfig(appDir)
		if err != nil {
			t.Fatalf("FindAndParseConfig() error = %v", err)
		}

		// Should only have 1 chart after deduplication
		if len(config.Charts) != 1 {
			t.Errorf("len(Charts) = %d, want 1 (duplicate should be removed)", len(config.Charts))
		}

		expectedPath := filepath.Join(rootDir, "common/chart1")
		if config.Charts[0].Path != expectedPath {
			t.Errorf("Charts[0].Path = %q, want %q", config.Charts[0].Path, expectedPath)
		}
	})

	t.Run("duplicate preflight paths removed", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := tmpDir
		appDir := filepath.Join(rootDir, "app")

		if err := os.MkdirAll(appDir, 0755); err != nil {
			t.Fatalf("creating directories: %v", err)
		}

		rootConfigPath := filepath.Join(rootDir, ".replicated")
		rootConfigData := []byte(`preflights:
  - path: ./checks/preflight1
    chartName: "test-chart"
    chartVersion: "1.0.0"
repl-lint:
`)
		if err := os.WriteFile(rootConfigPath, rootConfigData, 0644); err != nil {
			t.Fatalf("writing root config: %v", err)
		}

		appConfigPath := filepath.Join(appDir, ".replicated")
		appConfigData := []byte(`preflights:
  - path: ../checks/preflight1
    chartName: "test-chart"
    chartVersion: "1.0.0"
repl-lint:
`)
		if err := os.WriteFile(appConfigPath, appConfigData, 0644); err != nil {
			t.Fatalf("writing app config: %v", err)
		}

		config, err := parser.FindAndParseConfig(appDir)
		if err != nil {
			t.Fatalf("FindAndParseConfig() error = %v", err)
		}

		// Should only have 1 preflight after deduplication
		if len(config.Preflights) != 1 {
			t.Errorf("len(Preflights) = %d, want 1 (duplicate should be removed)", len(config.Preflights))
		}
	})

	t.Run("duplicate manifest paths removed", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := tmpDir
		appDir := filepath.Join(rootDir, "app")

		if err := os.MkdirAll(appDir, 0755); err != nil {
			t.Fatalf("creating directories: %v", err)
		}

		rootConfigPath := filepath.Join(rootDir, ".replicated")
		rootConfigData := []byte(`manifests:
  - "./manifests/**/*.yaml"
repl-lint:
`)
		if err := os.WriteFile(rootConfigPath, rootConfigData, 0644); err != nil {
			t.Fatalf("writing root config: %v", err)
		}

		appConfigPath := filepath.Join(appDir, ".replicated")
		appConfigData := []byte(`manifests:
  - "../manifests/**/*.yaml"
repl-lint:
`)
		if err := os.WriteFile(appConfigPath, appConfigData, 0644); err != nil {
			t.Fatalf("writing app config: %v", err)
		}

		config, err := parser.FindAndParseConfig(appDir)
		if err != nil {
			t.Fatalf("FindAndParseConfig() error = %v", err)
		}

		// Should only have 1 manifest after deduplication
		if len(config.Manifests) != 1 {
			t.Errorf("len(Manifests) = %d, want 1 (duplicate should be removed)", len(config.Manifests))
		}
	})

	t.Run("unique paths preserved, duplicates removed", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootDir := tmpDir
		appDir := filepath.Join(rootDir, "app")

		if err := os.MkdirAll(appDir, 0755); err != nil {
			t.Fatalf("creating directories: %v", err)
		}

		rootConfigPath := filepath.Join(rootDir, ".replicated")
		rootConfigData := []byte(`charts:
  - path: ./common/chart1
  - path: ./common/chart2
repl-lint:
`)
		if err := os.WriteFile(rootConfigPath, rootConfigData, 0644); err != nil {
			t.Fatalf("writing root config: %v", err)
		}

		appConfigPath := filepath.Join(appDir, ".replicated")
		appConfigData := []byte(`charts:
  - path: ../common/chart1
  - path: ./app-chart
repl-lint:
`)
		if err := os.WriteFile(appConfigPath, appConfigData, 0644); err != nil {
			t.Fatalf("writing app config: %v", err)
		}

		config, err := parser.FindAndParseConfig(appDir)
		if err != nil {
			t.Fatalf("FindAndParseConfig() error = %v", err)
		}

		// Should have 3 charts: chart1 (deduped), chart2, app-chart
		if len(config.Charts) != 3 {
			t.Errorf("len(Charts) = %d, want 3", len(config.Charts))
		}

		chartPaths := make(map[string]bool)
		for _, chart := range config.Charts {
			chartPaths[chart.Path] = true
		}

		expectedPaths := []string{
			filepath.Join(rootDir, "common/chart1"),
			filepath.Join(rootDir, "common/chart2"),
			filepath.Join(appDir, "app-chart"),
		}

		for _, expected := range expectedPaths {
			if !chartPaths[expected] {
				t.Errorf("Expected chart path %q not found in merged config", expected)
			}
		}
	})
}

func TestParseConfig_InvalidGlobPatterns(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		wantErrMsg string
	}{
		{
			name: "invalid chart glob pattern - unclosed bracket",
			configYAML: `
charts:
  - path: "./charts/[invalid"
`,
			wantErrMsg: "invalid glob pattern in charts[0].path",
		},
		{
			name: "invalid preflight glob pattern - unclosed brace",
			configYAML: `
preflights:
  - path: "./preflights/{unclosed"
    chartName: "test"
    chartVersion: "1.0.0"
`,
			wantErrMsg: "invalid glob pattern in preflights[0].path",
		},
		{
			name: "invalid manifest glob pattern - unclosed bracket",
			configYAML: `
manifests:
  - "./manifests/[invalid/*.yaml"
`,
			wantErrMsg: "invalid glob pattern in manifests[0]",
		},
		{
			name: "multiple invalid patterns",
			configYAML: `
charts:
  - path: "./charts/*"
  - path: "./charts/[bad"
preflights:
  - path: "./preflights/{invalid"
    chartName: "test"
    chartVersion: "1.0.0"
`,
			wantErrMsg: "invalid glob pattern in charts[1].path",
		},
		{
			name: "valid patterns should not error",
			configYAML: `
charts:
  - path: "./charts/**"
preflights:
  - path: "./preflights/{dev,prod}/*.yaml"
    chartName: "test"
    chartVersion: "1.0.0"
manifests:
  - "./manifests/**/*.yaml"
`,
			wantErrMsg: "", // No error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Write config file
			configPath := filepath.Join(tmpDir, ".replicated.yaml")
			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Parse config
			parser := NewConfigParser()
			_, err = parser.FindAndParseConfig(tmpDir)

			// Check error expectations
			if tt.wantErrMsg == "" {
				// Should succeed
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				// Should fail with specific error
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.wantErrMsg)
				}

				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Error %q does not contain %q", err.Error(), tt.wantErrMsg)
				}

				// Verify it says "invalid glob syntax"
				if !strings.Contains(err.Error(), "invalid glob syntax") {
					t.Errorf("Error %q does not contain 'invalid glob syntax'", err.Error())
				}
			}
		})
	}
}

// TestApplyDefaultsWithNilTools tests that ApplyDefaults correctly initializes tools map
func TestApplyDefaultsWithNilTools(t *testing.T) {
	parser := NewConfigParser()

	// Create config with ReplLint but no tools
	config := &Config{
		ReplLint: &ReplLintConfig{
			Version: 1,
			Linters: LintersConfig{
				Helm: LinterConfig{},
			},
			// Tools is nil here
		},
	}

	// Apply defaults
	parser.ApplyDefaults(config)

	// Check that tools map was initialized
	if config.ReplLint.Tools == nil {
		t.Fatal("Tools map should be initialized after ApplyDefaults")
	}

	// Check that all tools have "latest" as default
	if v, ok := config.ReplLint.Tools[ToolHelm]; !ok || v != "latest" {
		t.Errorf("Expected Helm to default to 'latest', got '%s' (exists: %v)", v, ok)
	}
	if v, ok := config.ReplLint.Tools[ToolPreflight]; !ok || v != "latest" {
		t.Errorf("Expected Preflight to default to 'latest', got '%s' (exists: %v)", v, ok)
	}
	if v, ok := config.ReplLint.Tools[ToolSupportBundle]; !ok || v != "latest" {
		t.Errorf("Expected SupportBundle to default to 'latest', got '%s' (exists: %v)", v, ok)
	}
}

// TestFindAndParseConfigWithMinimalConfig tests that a minimal config gets defaults applied
func TestFindAndParseConfigWithMinimalConfig(t *testing.T) {
	// Create a temporary directory with minimal config
	tmpDir := t.TempDir()

	// Create minimal .replicated config WITHOUT tool versions
	configPath := filepath.Join(tmpDir, ".replicated")
	configContent := `repl-lint:
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
	parser := NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Check that ReplLint exists
	if config.ReplLint == nil {
		t.Fatal("ReplLint should be initialized")
	}

	// Check that tools map was initialized with "latest" defaults
	if config.ReplLint.Tools == nil {
		t.Logf("Tools is nil, full ReplLint: %+v", config.ReplLint)
		t.Fatal("Tools map should be initialized")
	}

	// Log the tools map content for debugging
	t.Logf("Tools map length: %d", len(config.ReplLint.Tools))
	for k, v := range config.ReplLint.Tools {
		t.Logf("Tool %s = %s", k, v)
	}

	// All tools should default to "latest"
	if v, ok := config.ReplLint.Tools[ToolHelm]; !ok || v != "latest" {
		t.Errorf("Expected Helm to default to 'latest', got '%s' (exists: %v)", v, ok)
	}
	if v, ok := config.ReplLint.Tools[ToolPreflight]; !ok || v != "latest" {
		t.Errorf("Expected Preflight to default to 'latest', got '%s' (exists: %v)", v, ok)
	}
	if v, ok := config.ReplLint.Tools[ToolSupportBundle]; !ok || v != "latest" {
		t.Errorf("Expected SupportBundle to default to 'latest', got '%s' (exists: %v)", v, ok)
	}
}

// TestValidateConfig_PreflightWithoutChart tests that preflights without chart references are valid
func TestValidateConfig_PreflightWithoutChart(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".replicated")

	// Config with preflight but NO chart reference (Branch 2: linter decides)
	configData := []byte(`preflights:
  - path: "./preflight.yaml"
repl-lint:
`)
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	parser := NewConfigParser()
	_, err := parser.ParseConfigFile(configPath)
	if err != nil {
		t.Errorf("ParseConfigFile() unexpected error for preflight without chart reference: %v", err)
	}
}

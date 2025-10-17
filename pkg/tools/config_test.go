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
		parent := &Config{
			ReplLint: &ReplLintConfig{
				Version: 1,
				Enabled: true,
				Linters: LintersConfig{
					Helm: LinterConfig{Disabled: false, Strict: false},
				},
				Tools: map[string]string{
					"helm": "3.14.4",
				},
			},
		}
		child := &Config{
			ReplLint: &ReplLintConfig{
				Linters: LintersConfig{
					Helm: LinterConfig{Disabled: false, Strict: true},
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
		if merged.ReplLint.Linters.Helm.Strict != true {
			t.Errorf("Helm.Strict = %v, want true", merged.ReplLint.Linters.Helm.Strict)
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
    valuesPath: ./charts/chart1
  - path: ./preflights/check2
releaseLabel: "v{{.Version}}"
manifests:
  - "replicated/**/*.yaml"
  - "manifests/**/*.yaml"
repl-lint:
  enabled: true
  linters:
    helm:
      disabled: false
      strict: true
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
  enabled: true
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

	t.Run("relative preflight paths resolved to absolute", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".replicated")

		configData := []byte(`preflights:
  - path: ./preflights/check1
    valuesPath: ./charts/chart1
  - path: preflights/check2
    valuesPath: ../parent-charts/chart2
repl-lint:
  enabled: true
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

		// Check first preflight
		expectedPath := filepath.Join(tmpDir, "preflights/check1")
		if config.Preflights[0].Path != expectedPath {
			t.Errorf("preflights[0].Path = %q, want %q", config.Preflights[0].Path, expectedPath)
		}
		expectedValuesPath := filepath.Join(tmpDir, "charts/chart1")
		if config.Preflights[0].ValuesPath != expectedValuesPath {
			t.Errorf("preflights[0].ValuesPath = %q, want %q", config.Preflights[0].ValuesPath, expectedValuesPath)
		}

		// Check second preflight
		expectedPath2 := filepath.Join(tmpDir, "preflights/check2")
		if config.Preflights[1].Path != expectedPath2 {
			t.Errorf("preflights[1].Path = %q, want %q", config.Preflights[1].Path, expectedPath2)
		}
		expectedValuesPath2 := filepath.Join(tmpDir, "../parent-charts/chart2")
		if config.Preflights[1].ValuesPath != expectedValuesPath2 {
			t.Errorf("preflights[1].ValuesPath = %q, want %q", config.Preflights[1].ValuesPath, expectedValuesPath2)
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
  enabled: true
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

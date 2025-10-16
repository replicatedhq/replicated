package lint2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ReplicatedConfig represents the .replicated configuration file
type ReplicatedConfig struct {
	AppID    string         `yaml:"appId" json:"appId"`
	AppSlug  string         `yaml:"appSlug" json:"appSlug"`
	Charts   []ChartConfig  `yaml:"charts" json:"charts"`
	ReplLint ReplLintConfig `yaml:"repl-lint" json:"repl-lint"`
}

// ChartConfig represents a chart entry in the config
type ChartConfig struct {
	Path         string `yaml:"path" json:"path"`
	ChartVersion string `yaml:"chartVersion" json:"chartVersion"`
	AppVersion   string `yaml:"appVersion" json:"appVersion"`
}

// ReplLintConfig represents the repl-lint section
type ReplLintConfig struct {
	Version int                     `yaml:"version" json:"version"`
	Enabled bool                    `yaml:"enabled" json:"enabled"`
	Linters map[string]LinterConfig `yaml:"linters" json:"linters"`
	Tools   map[string]string       `yaml:"tools" json:"tools"`
}

// LinterConfig represents configuration for a specific linter
type LinterConfig struct {
	Disabled bool `yaml:"disabled" json:"disabled"`
	Strict   bool `yaml:"strict" json:"strict"`
}

// IsEnabled returns true if the linter is not disabled
func (c LinterConfig) IsEnabled() bool {
	return !c.Disabled
}

// LoadReplicatedConfig loads and parses the .replicated config file from the current directory
func LoadReplicatedConfig() (*ReplicatedConfig, error) {
	// Try .replicated.yaml first, then .replicated.yml, then .replicated (JSON)
	candidates := []string{".replicated.yaml", ".replicated.yml", ".replicated"}

	var configPath string
	var found bool

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf(".replicated config file not found (tried: %v)", candidates)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config ReplicatedConfig

	// Try YAML first (more common), fall back to JSON
	ext := filepath.Ext(configPath)
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	} else {
		// Try JSON
		if err := json.Unmarshal(data, &config); err != nil {
			// If JSON fails, try YAML as fallback
			if err := yaml.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config as JSON or YAML: %w", err)
			}
		}
	}

	// Set defaults
	if config.ReplLint.Version == 0 {
		config.ReplLint.Version = 1
	}

	// Default enabled to true if not specified
	if config.ReplLint.Linters == nil {
		config.ReplLint.Linters = make(map[string]LinterConfig)
	}

	// If helm linter config doesn't exist, default to enabled
	if _, exists := config.ReplLint.Linters["helm"]; !exists {
		config.ReplLint.Linters["helm"] = LinterConfig{
			Disabled: false,
			Strict:   false,
		}
	}

	// Default tools map
	if config.ReplLint.Tools == nil {
		config.ReplLint.Tools = make(map[string]string)
	}

	// Apply default tool versions if not specified
	if _, exists := config.ReplLint.Tools["helm"]; !exists {
		config.ReplLint.Tools["helm"] = "3.14.4"
	}

	return &config, nil
}

// ExpandChartPaths expands glob patterns in chart paths and returns a list of concrete paths
func ExpandChartPaths(chartConfigs []ChartConfig) ([]string, error) {
	var paths []string

	for _, chartConfig := range chartConfigs {
		// Check if path contains glob pattern
		if containsGlob(chartConfig.Path) {
			matches, err := filepath.Glob(chartConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to expand glob pattern %s: %w", chartConfig.Path, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("no charts found matching pattern: %s", chartConfig.Path)
			}
			paths = append(paths, matches...)
		} else {
			paths = append(paths, chartConfig.Path)
		}
	}

	return paths, nil
}

// containsGlob checks if a path contains glob wildcards
func containsGlob(path string) bool {
	return filepath.Base(path) != path &&
		(containsAny(path, []rune{'*', '?', '['}))
}

func containsAny(s string, chars []rune) bool {
	for _, c := range s {
		for _, target := range chars {
			if c == target {
				return true
			}
		}
	}
	return false
}

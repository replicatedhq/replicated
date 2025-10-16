package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigParser handles parsing of .replicated config files
type ConfigParser struct{}

// NewConfigParser creates a new config parser
func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

// FindAndParseConfig searches for a .replicated config file starting from the given path
// and walking up the directory tree. If path is empty, starts from current directory.
// Returns the parsed config or a default config if not found.
func (p *ConfigParser) FindAndParseConfig(startPath string) (*Config, error) {
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting current directory: %w", err)
		}
	}

	// Make absolute
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	// If startPath is a file, parse it directly
	info, err := os.Stat(absPath)
	if err == nil && !info.IsDir() {
		config, err := p.ParseConfigFile(absPath)
		if err != nil {
			return nil, err
		}
		// Apply defaults for single-file case
		p.applyDefaults(config)
		return config, nil
	}

	// Collect all config files from current dir to root
	var configPaths []string
	currentDir := absPath

	for {
		// Try .replicated first, then .replicated.yaml, then .replicated.json
		candidates := []string{
			filepath.Join(currentDir, ".replicated"),
			filepath.Join(currentDir, ".replicated.yaml"),
			filepath.Join(currentDir, ".replicated.json"),
		}

		for _, configPath := range candidates {
			if stat, err := os.Stat(configPath); err == nil {
				// Found config - make sure it's a file, not a directory
				if !stat.IsDir() {
					configPaths = append(configPaths, configPath)
					break // Only take first match per directory
				}
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root
			break
		}
		currentDir = parentDir
	}

	// No config files found
	if len(configPaths) == 0 {
		return nil, fmt.Errorf("no .replicated config file found (tried .replicated, .replicated.yaml, .replicated.json)")
	}

	// If only one config, just parse and return it
	if len(configPaths) == 1 {
		return p.ParseConfigFile(configPaths[0])
	}

	// Multiple configs found - parse and merge them
	// configPaths is ordered [child...parent], reverse to [parent...child]
	var configs []*Config
	for i := len(configPaths) - 1; i >= 0; i-- {
		config, err := p.ParseConfigFile(configPaths[i])
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", configPaths[i], err)
		}
		configs = append(configs, config)
	}

	// Merge all configs (later configs override earlier)
	merged := p.mergeConfigs(configs)

	// Apply defaults to merged config
	p.applyDefaults(merged)

	return merged, nil
}

// mergeConfigs merges multiple configs with later configs taking precedence
// Configs are ordered [parent, child, grandchild] - child overrides parent
func (p *ConfigParser) mergeConfigs(configs []*Config) *Config {
	if len(configs) == 0 {
		return p.DefaultConfig()
	}

	if len(configs) == 1 {
		return configs[0]
	}

	// Start with first config (most parent)
	merged := configs[0]

	// Merge in each subsequent config (moving toward child)
	for i := 1; i < len(configs); i++ {
		child := configs[i]

		// Merge ReplLint section
		if child.ReplLint != nil {
			if merged.ReplLint == nil {
				merged.ReplLint = child.ReplLint
			} else {
				// Merge version and enabled
				if child.ReplLint.Version != 0 {
					merged.ReplLint.Version = child.ReplLint.Version
				}
				merged.ReplLint.Enabled = child.ReplLint.Enabled

				// Merge linters (child completely overrides parent for each linter)
				merged.ReplLint.Linters.Helm = child.ReplLint.Linters.Helm
				merged.ReplLint.Linters.Preflight = child.ReplLint.Linters.Preflight
				merged.ReplLint.Linters.SupportBundle = child.ReplLint.Linters.SupportBundle
				merged.ReplLint.Linters.EmbeddedCluster = child.ReplLint.Linters.EmbeddedCluster
				merged.ReplLint.Linters.Kots = child.ReplLint.Linters.Kots

				// Merge tools map (child versions override parent)
				if child.ReplLint.Tools != nil {
					if merged.ReplLint.Tools == nil {
						merged.ReplLint.Tools = make(map[string]string)
					}
					for toolName, version := range child.ReplLint.Tools {
						merged.ReplLint.Tools[toolName] = version
					}
				}
			}
		}
	}

	return merged
}

// ParseConfigFile parses a .replicated config file (supports both YAML and JSON)
func (p *ConfigParser) ParseConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	return p.ParseConfig(data)
}

// ParseConfig parses config data (auto-detects YAML or JSON)
// Does NOT apply defaults - caller should do that after merging
func (p *ConfigParser) ParseConfig(data []byte) (*Config, error) {
	var config Config

	// Try YAML first (JSON is valid YAML)
	if err := yaml.Unmarshal(data, &config); err != nil {
		// If YAML fails, try JSON explicitly
		if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
			return nil, fmt.Errorf("parsing config (tried YAML and JSON): %w", err)
		}
	}

	// Validate but don't apply defaults
	if err := p.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &config, nil
}

// DefaultConfig returns a config with default values
func (p *ConfigParser) DefaultConfig() *Config {
	config := &Config{
		ReplLint: &ReplLintConfig{
			Version: 1,
			Enabled: true,
			Linters: LintersConfig{
				Helm:            LinterConfig{Disabled: false, Strict: false}, // disabled: false = enabled
				Preflight:       LinterConfig{Disabled: false, Strict: false},
				SupportBundle:   LinterConfig{Disabled: false, Strict: false},
				EmbeddedCluster: LinterConfig{Disabled: true, Strict: false}, // disabled: true = disabled
				Kots:            LinterConfig{Disabled: true, Strict: false},
			},
			Tools: make(map[string]string),
		},
	}

	p.applyDefaults(config)
	return config
}

// applyDefaults fills in default values for missing fields
func (p *ConfigParser) applyDefaults(config *Config) {
	// Initialize lint config if nil
	if config.ReplLint == nil {
		config.ReplLint = &ReplLintConfig{
			Version: 1,
			Enabled: true,
			Linters: LintersConfig{
				Helm:            LinterConfig{Disabled: false, Strict: false},
				Preflight:       LinterConfig{Disabled: false, Strict: false},
				SupportBundle:   LinterConfig{Disabled: false, Strict: false},
				EmbeddedCluster: LinterConfig{Disabled: true, Strict: false},
				Kots:            LinterConfig{Disabled: true, Strict: false},
			},
			Tools: make(map[string]string),
		}
	}

	// Default version
	if config.ReplLint.Version == 0 {
		config.ReplLint.Version = 1
	}

	// Default tools map
	if config.ReplLint.Tools == nil {
		config.ReplLint.Tools = make(map[string]string)
	}

	// Apply default tool versions if not specified
	if _, exists := config.ReplLint.Tools[ToolHelm]; !exists {
		config.ReplLint.Tools[ToolHelm] = DefaultHelmVersion
	}
	if _, exists := config.ReplLint.Tools[ToolPreflight]; !exists {
		config.ReplLint.Tools[ToolPreflight] = DefaultPreflightVersion
	}
	if _, exists := config.ReplLint.Tools[ToolSupportBundle]; !exists {
		config.ReplLint.Tools[ToolSupportBundle] = DefaultSupportBundleVersion
	}
}

// validateConfig validates the config structure
func (p *ConfigParser) validateConfig(config *Config) error {
	// Skip validation if no lint config
	if config.ReplLint == nil {
		return nil
	}

	// Validate version (0 is allowed, will be defaulted to 1)
	if config.ReplLint.Version < 0 {
		return fmt.Errorf("invalid version %d: must be >= 0", config.ReplLint.Version)
	}

	// Validate tool versions (semantic versioning)
	for toolName, version := range config.ReplLint.Tools {
		if !isValidSemver(version) {
			return fmt.Errorf("invalid version %q for tool %q: must be semantic version (e.g., 1.2.3)", version, toolName)
		}
	}

	return nil
}

// GetToolVersions extracts the tool versions from a config
func GetToolVersions(config *Config) map[string]string {
	if config == nil || config.ReplLint == nil {
		return make(map[string]string)
	}

	// Return a copy to prevent modification
	versions := make(map[string]string, len(config.ReplLint.Tools))
	for k, v := range config.ReplLint.Tools {
		versions[k] = v
	}
	return versions
}

// isValidSemver checks if a version string is valid semantic versioning
// Accepts formats like: 1.2.3, v1.2.3, 1.2.3-beta, 1.2.3+build
func isValidSemver(version string) bool {
	// Remove leading 'v' if present
	version = strings.TrimPrefix(version, "v")

	// Basic semver regex pattern
	// Matches: major.minor.patch with optional pre-release and build metadata
	semverPattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

	matched, _ := regexp.MatchString(semverPattern, version)
	return matched
}

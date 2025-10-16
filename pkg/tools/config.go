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
		return p.ParseConfigFile(absPath)
	}

	// Walk up directory tree looking for .replicated
	currentDir := absPath
	for {
		configPath := filepath.Join(currentDir, ".replicated")
		if _, err := os.Stat(configPath); err == nil {
			return p.ParseConfigFile(configPath)
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root without finding config
			break
		}
		currentDir = parentDir
	}

	// No config found - return default config
	return p.DefaultConfig(), nil
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
func (p *ConfigParser) ParseConfig(data []byte) (*Config, error) {
	var config Config

	// Try YAML first (JSON is valid YAML)
	if err := yaml.Unmarshal(data, &config); err != nil {
		// If YAML fails, try JSON explicitly
		if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
			return nil, fmt.Errorf("parsing config (tried YAML and JSON): %w", err)
		}
	}

	// Apply defaults
	p.applyDefaults(&config)

	// Validate
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
				EmbeddedCluster: LinterConfig{Disabled: true, Strict: false},  // disabled: true = disabled
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

	// Validate version
	if config.ReplLint.Version < 1 {
		return fmt.Errorf("invalid version %d: must be >= 1", config.ReplLint.Version)
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

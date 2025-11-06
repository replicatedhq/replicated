package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
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
		p.ApplyDefaults(config)
		return config, nil
	}

	// Collect all config files from current dir to root
	var configPaths []string
	currentDir := absPath

	for {
		// Try .replicated first, then .replicated.yaml
		candidates := []string{
			filepath.Join(currentDir, ".replicated"),
			filepath.Join(currentDir, ".replicated.yaml"),
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

	// No config files found - return default config for auto-discovery mode
	if len(configPaths) == 0 {
		defaultConfig := p.DefaultConfig()
		return defaultConfig, nil
	}

	// If only one config, parse it and apply defaults
	if len(configPaths) == 1 {
		config, err := p.ParseConfigFile(configPaths[0])
		if err != nil {
			return nil, err
		}
		// Apply defaults to single config
		p.ApplyDefaults(config)
		return config, nil
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
	p.ApplyDefaults(merged)

	// Deduplicate resources (charts, preflights, manifests)
	p.deduplicateResources(merged)

	return merged, nil
}

// mergeConfigs merges multiple configs with later configs taking precedence
// Configs are ordered [parent, child, grandchild] - child overrides parent
//
// Merge strategy:
// - Scalar fields (override): appId, appSlug, releaseLabel - child wins
// - Channel arrays (override): promoteToChannelIds, promoteToChannelNames - child replaces if non-empty
// - Resource arrays (append): charts, preflights, manifests - accumulate from all configs
// - ReplLint section (override): child settings override parent
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

		// Scalar fields: child overrides parent (if non-empty)
		if child.AppId != "" {
			merged.AppId = child.AppId
		}
		if child.AppSlug != "" {
			merged.AppSlug = child.AppSlug
		}
		if child.ReleaseLabel != "" {
			merged.ReleaseLabel = child.ReleaseLabel
		}

		// Channel arrays: child completely replaces parent (if non-empty)
		// This is an override, not an append, because promotion targets are a decision
		if len(child.PromoteToChannelIds) > 0 {
			merged.PromoteToChannelIds = child.PromoteToChannelIds
		}
		if len(child.PromoteToChannelNames) > 0 {
			merged.PromoteToChannelNames = child.PromoteToChannelNames
		}

		// Resource arrays: append child to parent
		// This allows monorepo configs to accumulate resources from all levels
		merged.Charts = append(merged.Charts, child.Charts...)
		merged.Preflights = append(merged.Preflights, child.Preflights...)
		merged.Manifests = append(merged.Manifests, child.Manifests...)

		// Merge ReplLint section
		if child.ReplLint != nil {
			if merged.ReplLint == nil {
				merged.ReplLint = child.ReplLint
			} else {
				// Merge version (override if non-zero)
				if child.ReplLint.Version != 0 {
					merged.ReplLint.Version = child.ReplLint.Version
				}

				// Merge linters (only override fields explicitly set in child)
				merged.ReplLint.Linters.Helm = mergeLinterConfig(merged.ReplLint.Linters.Helm, child.ReplLint.Linters.Helm)
				merged.ReplLint.Linters.Preflight = mergeLinterConfig(merged.ReplLint.Linters.Preflight, child.ReplLint.Linters.Preflight)
				merged.ReplLint.Linters.SupportBundle = mergeLinterConfig(merged.ReplLint.Linters.SupportBundle, child.ReplLint.Linters.SupportBundle)
				merged.ReplLint.Linters.EmbeddedCluster = mergeLinterConfig(merged.ReplLint.Linters.EmbeddedCluster, child.ReplLint.Linters.EmbeddedCluster)
				merged.ReplLint.Linters.Kots = mergeLinterConfig(merged.ReplLint.Linters.Kots, child.ReplLint.Linters.Kots)

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

// ParseConfigFile parses a .replicated config file (supports YAML)
func (p *ConfigParser) ParseConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	config, err := p.ParseConfig(data)
	if err != nil {
		return nil, err
	}

	// Resolve all relative paths to absolute paths relative to the config file
	// This ensures paths work correctly regardless of where the command is invoked
	p.resolvePaths(config, path)

	return config, nil
}

// ParseConfig parses config data from YAML
// Does NOT apply defaults - caller should do that after merging
func (p *ConfigParser) ParseConfig(data []byte) (*Config, error) {
	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config as YAML: %w", err)
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
			Linters: defaultLintersConfig(),
			Tools:   make(map[string]string),
		},
	}

	p.ApplyDefaults(config)
	return config
}

// ApplyDefaults fills in default values for missing fields
func (p *ConfigParser) ApplyDefaults(config *Config) {
	// Initialize lint config if nil
	if config.ReplLint == nil {
		config.ReplLint = &ReplLintConfig{
			Version: 1,
			Linters: defaultLintersConfig(),
			Tools:   make(map[string]string),
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

	// Apply "latest" for tool versions if not specified
	// The resolver will fetch the actual latest version from GitHub
	defaultTools := []string{
		ToolHelm,
		ToolPreflight,
		ToolSupportBundle,
		ToolEmbeddedCluster,
		ToolKots,
	}
	for _, tool := range defaultTools {
		if _, exists := config.ReplLint.Tools[tool]; !exists {
			config.ReplLint.Tools[tool] = "latest"
		}
	}
}

// validateConfig validates the config structure
func (p *ConfigParser) validateConfig(config *Config) error {
	// Validate chart paths
	for i, chart := range config.Charts {
		if chart.Path == "" {
			return fmt.Errorf("chart[%d]: path is required", i)
		}
	}

	// Validate preflight paths
	for i, preflight := range config.Preflights {
		if preflight.Path == "" {
			return fmt.Errorf("preflight[%d]: path is required", i)
		}

		// chartName and chartVersion are optional, but must be provided together
		if preflight.ChartName != "" && preflight.ChartVersion == "" {
			return fmt.Errorf("preflight[%d]: chartVersion is required when chartName is specified", i)
		}
		if preflight.ChartVersion != "" && preflight.ChartName == "" {
			return fmt.Errorf("preflight[%d]: chartName is required when chartVersion is specified", i)
		}
	}

	// Validate manifest paths (array can be empty, but elements cannot be empty strings)
	for i, manifest := range config.Manifests {
		if manifest == "" {
			return fmt.Errorf("manifest[%d]: path cannot be empty string", i)
		}
	}

	// Validate glob patterns in all paths
	if err := p.validateGlobPatterns(config); err != nil {
		return err
	}

	// Skip validation if no lint config
	if config.ReplLint == nil {
		return nil
	}

	// Validate version (0 is allowed, will be defaulted to 1)
	if config.ReplLint.Version < 0 {
		return fmt.Errorf("invalid version %d: must be >= 0", config.ReplLint.Version)
	}

	// Validate tool versions (semantic versioning or "latest")
	for toolName, version := range config.ReplLint.Tools {
		// Allow "latest" as a special case
		if version != "latest" && !isValidSemver(version) {
			return fmt.Errorf("invalid version %q for tool %q: must be semantic version (e.g., 1.2.3) or 'latest'", version, toolName)
		}
	}

	return nil
}

// validateGlobPatterns validates all glob patterns in the config for correct syntax.
// This provides early validation before attempting to expand patterns during linting.
func (p *ConfigParser) validateGlobPatterns(config *Config) error {
	// Validate chart paths
	for i, chart := range config.Charts {
		if containsGlob(chart.Path) {
			if !doublestar.ValidatePattern(chart.Path) {
				return fmt.Errorf("invalid glob pattern in charts[%d].path %q: invalid glob syntax", i, chart.Path)
			}
		}
	}

	// Validate preflight paths
	for i, preflight := range config.Preflights {
		if containsGlob(preflight.Path) {
			if !doublestar.ValidatePattern(preflight.Path) {
				return fmt.Errorf("invalid glob pattern in preflights[%d].path %q: invalid glob syntax", i, preflight.Path)
			}
		}
	}

	// Validate manifest patterns
	for i, manifest := range config.Manifests {
		if containsGlob(manifest) {
			if !doublestar.ValidatePattern(manifest) {
				return fmt.Errorf("invalid glob pattern in manifests[%d] %q: invalid glob syntax", i, manifest)
			}
		}
	}

	return nil
}

// containsGlob checks if a path contains glob wildcards (* ? [ {)
func containsGlob(path string) bool {
	return strings.ContainsAny(path, "*?[{")
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

// resolvePaths resolves all relative paths in the config to absolute paths
// relative to the config file's directory. This ensures paths work correctly
// regardless of where the command is invoked.
func (p *ConfigParser) resolvePaths(config *Config, configFilePath string) {
	if config == nil {
		return
	}

	// Get the directory containing the config file
	configDir := filepath.Dir(configFilePath)

	// Resolve chart paths
	for i := range config.Charts {
		// Only resolve relative paths - leave absolute paths as-is
		if !filepath.IsAbs(config.Charts[i].Path) {
			config.Charts[i].Path = filepath.Join(configDir, config.Charts[i].Path)
		}
	}

	// Resolve preflight paths
	for i := range config.Preflights {
		// Resolve preflight path
		if config.Preflights[i].Path != "" && !filepath.IsAbs(config.Preflights[i].Path) {
			config.Preflights[i].Path = filepath.Join(configDir, config.Preflights[i].Path)
		}
		// Note: chartName and chartVersion are not paths - don't resolve them
	}

	// Resolve manifest paths (glob patterns)
	for i := range config.Manifests {
		// Manifests are glob patterns - resolve base directory but preserve pattern
		if !filepath.IsAbs(config.Manifests[i]) {
			config.Manifests[i] = filepath.Join(configDir, config.Manifests[i])
		}
	}
}

// mergeLinterConfig merges two linter configs
// Only overrides parent fields if child explicitly sets them (non-nil)
func mergeLinterConfig(parent, child LinterConfig) LinterConfig {
	result := parent

	// Override disabled if child explicitly sets it
	if child.Disabled != nil {
		result.Disabled = child.Disabled
	}

	return result
}

// defaultLintersConfig returns the default linter configuration
// with all linters enabled by default
func defaultLintersConfig() LintersConfig {
	return LintersConfig{
		Helm:            LinterConfig{Disabled: boolPtr(false)},
		Preflight:       LinterConfig{Disabled: boolPtr(false)},
		SupportBundle:   LinterConfig{Disabled: boolPtr(false)},
		EmbeddedCluster: LinterConfig{Disabled: boolPtr(false)},
		Kots:            LinterConfig{Disabled: boolPtr(false)},
	}
}

// boolPtr returns a pointer to a boolean value
// Helper for creating pointer booleans in config defaults
func boolPtr(b bool) *bool {
	return &b
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

// deduplicateResources removes duplicate entries from resource arrays
// Deduplication is based on absolute paths (which have already been resolved)
func (p *ConfigParser) deduplicateResources(config *Config) {
	if config == nil {
		return
	}

	// Deduplicate charts by path
	if len(config.Charts) > 0 {
		seen := make(map[string]bool)
		unique := make([]ChartConfig, 0, len(config.Charts))
		for _, chart := range config.Charts {
			if !seen[chart.Path] {
				seen[chart.Path] = true
				unique = append(unique, chart)
			}
		}
		config.Charts = unique
	}

	// Deduplicate preflights by path
	if len(config.Preflights) > 0 {
		seen := make(map[string]bool)
		unique := make([]PreflightConfig, 0, len(config.Preflights))
		for _, preflight := range config.Preflights {
			if !seen[preflight.Path] {
				seen[preflight.Path] = true
				unique = append(unique, preflight)
			}
		}
		config.Preflights = unique
	}

	// Deduplicate manifests (they are just strings)
	if len(config.Manifests) > 0 {
		seen := make(map[string]bool)
		unique := make([]string, 0, len(config.Manifests))
		for _, manifest := range config.Manifests {
			if !seen[manifest] {
				seen[manifest] = true
				unique = append(unique, manifest)
			}
		}
		config.Manifests = unique
	}
}

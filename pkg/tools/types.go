package tools

// Config represents the parsed .replicated configuration file
type Config struct {
	AppId                 string            `yaml:"appId,omitempty"`
	AppSlug               string            `yaml:"appSlug,omitempty"`
	PromoteToChannelIds   []string          `yaml:"promoteToChannelIds,omitempty"`
	PromoteToChannelNames []string          `yaml:"promoteToChannelNames,omitempty"`
	Charts                []ChartConfig     `yaml:"charts,omitempty"`
	Preflights            []PreflightConfig `yaml:"preflights,omitempty"`
	ReleaseLabel          string            `yaml:"releaseLabel,omitempty"`
	Manifests             []string          `yaml:"manifests,omitempty"`
	ReplLint              *ReplLintConfig   `yaml:"repl-lint,omitempty"`
}

// ChartConfig represents a chart entry in the config
type ChartConfig struct {
	Path         string `yaml:"path"`
	ChartVersion string `yaml:"chartVersion,omitempty"`
	AppVersion   string `yaml:"appVersion,omitempty"`
}

// PreflightConfig represents a preflight entry in the config
// Path is required. ChartName and ChartVersion are optional but must be provided together.
type PreflightConfig struct {
	Path         string `yaml:"path"`
	ChartName    string `yaml:"chartName,omitempty"`    // Optional: explicit chart reference (must provide chartVersion if set)
	ChartVersion string `yaml:"chartVersion,omitempty"` // Optional: explicit chart version (must provide chartName if set)
}

// ReplLintConfig is the lint configuration section
type ReplLintConfig struct {
	Version int               `yaml:"version"`
	Linters LintersConfig     `yaml:"linters"`
	Tools   map[string]string `yaml:"tools,omitempty"`
}

// LintersConfig contains configuration for each linter
type LintersConfig struct {
	Helm            LinterConfig `yaml:"helm"`
	Preflight       LinterConfig `yaml:"preflight"`
	SupportBundle   LinterConfig `yaml:"support-bundle"`
	EmbeddedCluster LinterConfig `yaml:"embedded-cluster"`
	Kots            LinterConfig `yaml:"kots"`
}

// LinterConfig represents the configuration for a single linter
type LinterConfig struct {
	Disabled *bool `yaml:"disabled,omitempty"` // pointer allows nil = not set
}

// IsEnabled returns true if the linter is not disabled
// nil Disabled means not set, defaults to enabled (false = not disabled)
func (c LinterConfig) IsEnabled() bool {
	return c.Disabled == nil || !*c.Disabled
}

// Default tool versions - kept for backward compatibility in tests
// In production, "latest" is used to fetch the most recent stable version from GitHub
const (
	DefaultHelmVersion          = "3.14.4" // Deprecated: Use "latest" instead
	DefaultPreflightVersion     = "latest"
	DefaultSupportBundleVersion = "latest"
)

// Supported tool names
const (
	ToolHelm          = "helm"
	ToolPreflight     = "preflight"
	ToolSupportBundle = "support-bundle"
)

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
type PreflightConfig struct {
	Path       string `yaml:"path"`
	ValuesPath string `yaml:"valuesPath,omitempty"`
}

// ReplLintConfig is the lint configuration section
type ReplLintConfig struct {
	Version int               `yaml:"version"`
	Enabled bool              `yaml:"enabled"`
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
	Disabled bool `yaml:"disabled"`
	Strict   bool `yaml:"strict"`
}

// IsEnabled returns true if the linter is not disabled
func (c LinterConfig) IsEnabled() bool {
	return !c.Disabled
}

// Default tool versions
const (
	DefaultHelmVersion          = "3.14.4"
	DefaultPreflightVersion     = "0.123.9"
	DefaultSupportBundleVersion = "0.123.9"
)

// Supported tool names
const (
	ToolHelm          = "helm"
	ToolPreflight     = "preflight"
	ToolSupportBundle = "support-bundle"
)

package tools

// Config represents the parsed .replicated configuration file
type Config struct {
	AppId                 string            `json:"appId,omitempty" yaml:"appId,omitempty"`
	AppSlug               string            `json:"appSlug,omitempty" yaml:"appSlug,omitempty"`
	PromoteToChannelIds   []string          `json:"promoteToChannelIds,omitempty" yaml:"promoteToChannelIds,omitempty"`
	PromoteToChannelNames []string          `json:"promoteToChannelNames,omitempty" yaml:"promoteToChannelNames,omitempty"`
	Charts                []ChartConfig     `json:"charts,omitempty" yaml:"charts,omitempty"`
	Preflights            []PreflightConfig `json:"preflights,omitempty" yaml:"preflights,omitempty"`
	ReleaseLabel          string            `json:"releaseLabel,omitempty" yaml:"releaseLabel,omitempty"`
	Manifests             []string          `json:"manifests,omitempty" yaml:"manifests,omitempty"`
	ReplLint              *ReplLintConfig   `json:"repl-lint,omitempty" yaml:"repl-lint,omitempty"`
}

// ChartConfig represents a chart entry in the config
type ChartConfig struct {
	Path         string `yaml:"path" json:"path"`
	ChartVersion string `yaml:"chartVersion,omitempty" json:"chartVersion,omitempty"`
	AppVersion   string `yaml:"appVersion,omitempty" json:"appVersion,omitempty"`
}

// PreflightConfig represents a preflight entry in the config
type PreflightConfig struct {
	Path       string `yaml:"path" json:"path"`
	ValuesPath string `yaml:"valuesPath,omitempty" json:"valuesPath,omitempty"`
}

// ReplLintConfig is the lint configuration section
type ReplLintConfig struct {
	Version int               `json:"version" yaml:"version"`
	Enabled bool              `json:"enabled" yaml:"enabled"`
	Linters LintersConfig     `json:"linters" yaml:"linters"`
	Tools   map[string]string `json:"tools,omitempty" yaml:"tools,omitempty"`
}

// LintersConfig contains configuration for each linter
type LintersConfig struct {
	Helm            LinterConfig `json:"helm" yaml:"helm"`
	Preflight       LinterConfig `json:"preflight" yaml:"preflight"`
	SupportBundle   LinterConfig `json:"support-bundle" yaml:"support-bundle"`
	EmbeddedCluster LinterConfig `json:"embedded-cluster" yaml:"embedded-cluster"`
	Kots            LinterConfig `json:"kots" yaml:"kots"`
}

// LinterConfig represents the configuration for a single linter
type LinterConfig struct {
	Disabled bool `json:"disabled" yaml:"disabled"`
	Strict   bool `json:"strict" yaml:"strict"`
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

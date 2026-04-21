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
	Helm            LinterConfig   `yaml:"helm"`
	Preflight       LinterConfig   `yaml:"preflight"`
	SupportBundle   LinterConfig   `yaml:"support-bundle"`
	EmbeddedCluster ECLinterConfig `yaml:"embedded-cluster"`
	Kots            LinterConfig   `yaml:"kots"`
}

// LinterConfig represents the configuration for a single linter.
// nil Disabled is a transient parse state — ApplyDefaults always fills it in
// before IsEnabled is called.
type LinterConfig struct {
	Disabled *bool `yaml:"disabled,omitempty"`
}

// IsEnabled returns true if the linter is not disabled.
// nil is treated as enabled; in practice ApplyDefaults always sets an explicit value.
func (c LinterConfig) IsEnabled() bool {
	return c.Disabled == nil || !*c.Disabled
}

// DefaultECDisableChecks are the checker IDs disabled by default when running EC lint.
var DefaultECDisableChecks = []string{"helmchart-archive", "ecconfig-helmchart-archive"}

// ECLinterConfig is the linter config for the Embedded Cluster linter.
// It embeds LinterConfig and adds EC-specific fields.
type ECLinterConfig struct {
	LinterConfig  `yaml:",inline"`
	DisableChecks []string `yaml:"disable-checks,omitempty"`
	BinaryPath    string   `yaml:"binary-path,omitempty"`
}

// GetDisableChecks returns the checks to disable. If DisableChecks is not set,
// it returns the default list.
func (c ECLinterConfig) GetDisableChecks() []string {
	if len(c.DisableChecks) > 0 {
		return c.DisableChecks
	}
	return DefaultECDisableChecks
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
	ToolHelm            = "helm"
	ToolPreflight       = "preflight"
	ToolSupportBundle   = "support-bundle"
	ToolEmbeddedCluster = "embedded-cluster"
)

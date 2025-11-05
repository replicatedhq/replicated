package cmd

import (
	"time"

	"github.com/replicatedhq/replicated/pkg/imageextract"
	"github.com/replicatedhq/replicated/pkg/lint2"
)

// JSONLintOutput represents the complete JSON output structure for lint results
type JSONLintOutput struct {
	Metadata               LintMetadata                `json:"metadata"`
	HelmResults            *HelmLintResults            `json:"helm_results,omitempty"`
	PreflightResults       *PreflightLintResults       `json:"preflight_results,omitempty"`
	SupportBundleResults   *SupportBundleLintResults   `json:"support_bundle_results,omitempty"`
	EmbeddedClusterResults *EmbeddedClusterLintResults `json:"embedded_cluster_results,omitempty"`
	KotsResults            *KotsLintResults            `json:"kots_results,omitempty"`
	Summary                LintSummary                 `json:"summary"`
	Images                 *ImageExtractResults        `json:"images,omitempty"` // Only if --verbose
}

// LintMetadata contains execution context and environment information
type LintMetadata struct {
	Timestamp              string `json:"timestamp"`
	ConfigFile             string `json:"config_file"`
	HelmVersion            string `json:"helm_version,omitempty"`
	PreflightVersion       string `json:"preflight_version,omitempty"`
	SupportBundleVersion   string `json:"support_bundle_version,omitempty"`
	EmbeddedClusterVersion string `json:"embedded_cluster_version,omitempty"`
	KotsVersion            string `json:"kots_version,omitempty"`
	CLIVersion             string `json:"cli_version"`
}

// HelmLintResults contains all Helm chart lint results
type HelmLintResults struct {
	Enabled bool              `json:"enabled"`
	Charts  []ChartLintResult `json:"charts"`
}

// ChartLintResult represents lint results for a single Helm chart
type ChartLintResult struct {
	Path     string          `json:"path"`
	Success  bool            `json:"success"`
	Messages []LintMessage   `json:"messages"`
	Summary  ResourceSummary `json:"summary"`
}

// PreflightLintResults contains all Preflight spec lint results
type PreflightLintResults struct {
	Enabled bool                  `json:"enabled"`
	Specs   []PreflightLintResult `json:"specs"`
}

// PreflightLintResult represents lint results for a single Preflight spec
type PreflightLintResult struct {
	Path     string          `json:"path"`
	Success  bool            `json:"success"`
	Messages []LintMessage   `json:"messages"`
	Summary  ResourceSummary `json:"summary"`
}

// SupportBundleLintResults contains all Support Bundle spec lint results
type SupportBundleLintResults struct {
	Enabled bool                      `json:"enabled"`
	Specs   []SupportBundleLintResult `json:"specs"`
}

// SupportBundleLintResult represents lint results for a single Support Bundle spec
type SupportBundleLintResult struct {
	Path     string          `json:"path"`
	Success  bool            `json:"success"`
	Messages []LintMessage   `json:"messages"`
	Summary  ResourceSummary `json:"summary"`
}

// EmbeddedClusterLintResults contains all Embedded Cluster config lint results
type EmbeddedClusterLintResults struct {
	Enabled bool                        `json:"enabled"`
	Configs []EmbeddedClusterLintResult `json:"configs"`
}

// EmbeddedClusterLintResult represents lint results for a single Embedded Cluster config
type EmbeddedClusterLintResult struct {
	Path     string          `json:"path"`
	Success  bool            `json:"success"`
	Messages []LintMessage   `json:"messages"`
	Summary  ResourceSummary `json:"summary"`
}

// KotsLintResults contains all KOTS manifest lint results
type KotsLintResults struct {
	Enabled   bool             `json:"enabled"`
	Manifests []KotsLintResult `json:"manifests"`
}

// KotsLintResult represents lint results for a single KOTS manifest
type KotsLintResult struct {
	Path     string          `json:"path"`
	Success  bool            `json:"success"`
	Messages []LintMessage   `json:"messages"`
	Summary  ResourceSummary `json:"summary"`
}

// LintMessage represents a single lint issue (wraps lint2.LintMessage with JSON tags)
type LintMessage struct {
	Severity string `json:"severity"` // ERROR, WARNING, INFO
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
}

// ResourceSummary contains counts by severity for a resource
type ResourceSummary struct {
	ErrorCount   int `json:"error_count"`
	WarningCount int `json:"warning_count"`
	InfoCount    int `json:"info_count"`
}

// LintSummary contains overall statistics across all linted resources
type LintSummary struct {
	TotalResources  int  `json:"total_resources"`
	PassedResources int  `json:"passed_resources"`
	FailedResources int  `json:"failed_resources"`
	TotalErrors     int  `json:"total_errors"`
	TotalWarnings   int  `json:"total_warnings"`
	TotalInfo       int  `json:"total_info"`
	OverallSuccess  bool `json:"overall_success"`
}

// ImageExtractResults contains extracted image information
type ImageExtractResults struct {
	Images   []imageextract.ImageRef `json:"images"`
	Warnings []imageextract.Warning  `json:"warnings"`
	Summary  ImageSummary            `json:"summary"`
}

// ImageSummary contains summary statistics for extracted images
type ImageSummary struct {
	TotalImages  int `json:"total_images"`
	UniqueImages int `json:"unique_images"`
}

// ExtractedPaths contains all paths and metadata needed for linting.
// This struct consolidates extraction logic across all linters to avoid duplication.
type ExtractedPaths struct {
	// Helm: simple paths (Chart.yaml validation delegated to helm tool)
	ChartPaths []string

	// Preflight: paths with chart metadata for template rendering
	Preflights []lint2.PreflightWithValues

	// Support bundles: simple paths
	SupportBundles []string

	// Embedded cluster: simple paths
	EmbeddedClusterPaths []string

	// KOTS: simple paths
	KotsPaths []string

	// Shared: HelmChart manifests (used by preflight + image extraction)
	HelmChartManifests map[string]*lint2.HelmChartManifest

	// Image extraction: charts with metadata (only if verbose)
	ChartsWithMetadata []lint2.ChartWithMetadata

	// Tool versions
	HelmVersion      string
	PreflightVersion string
	SBVersion        string
	ECVersion        string
	KotsVersion      string

	// Metadata
	ConfigPath string
}

// LintableResult is an interface for types that contain lint results.
// This allows generic handling of chart, preflight, and support bundle results.
type LintableResult interface {
	GetPath() string
	GetSuccess() bool
	GetMessages() []LintMessage
	GetSummary() ResourceSummary
}

// Implement LintableResult interface for ChartLintResult
func (c ChartLintResult) GetPath() string             { return c.Path }
func (c ChartLintResult) GetSuccess() bool            { return c.Success }
func (c ChartLintResult) GetMessages() []LintMessage  { return c.Messages }
func (c ChartLintResult) GetSummary() ResourceSummary { return c.Summary }

// Implement LintableResult interface for PreflightLintResult
func (p PreflightLintResult) GetPath() string             { return p.Path }
func (p PreflightLintResult) GetSuccess() bool            { return p.Success }
func (p PreflightLintResult) GetMessages() []LintMessage  { return p.Messages }
func (p PreflightLintResult) GetSummary() ResourceSummary { return p.Summary }

// Implement LintableResult interface for SupportBundleLintResult
func (s SupportBundleLintResult) GetPath() string             { return s.Path }
func (s SupportBundleLintResult) GetSuccess() bool            { return s.Success }
func (s SupportBundleLintResult) GetMessages() []LintMessage  { return s.Messages }
func (s SupportBundleLintResult) GetSummary() ResourceSummary { return s.Summary }

// Implement LintableResult interface for EmbeddedClusterLintResult
func (e EmbeddedClusterLintResult) GetPath() string             { return e.Path }
func (e EmbeddedClusterLintResult) GetSuccess() bool            { return e.Success }
func (e EmbeddedClusterLintResult) GetMessages() []LintMessage  { return e.Messages }
func (e EmbeddedClusterLintResult) GetSummary() ResourceSummary { return e.Summary }

// Implement LintableResult interface for KotsLintResult
func (k KotsLintResult) GetPath() string             { return k.Path }
func (k KotsLintResult) GetSuccess() bool            { return k.Success }
func (k KotsLintResult) GetMessages() []LintMessage  { return k.Messages }
func (k KotsLintResult) GetSummary() ResourceSummary { return k.Summary }

// Helper functions to convert between types

// convertLint2Messages converts lint2.LintMessage slice to LintMessage slice
func convertLint2Messages(messages []lint2.LintMessage) []LintMessage {
	result := make([]LintMessage, len(messages))
	for i, msg := range messages {
		result[i] = LintMessage{
			Severity: msg.Severity,
			Path:     msg.Path,
			Message:  msg.Message,
		}
	}
	return result
}

// calculateResourceSummary calculates summary from lint messages
func calculateResourceSummary(messages []lint2.LintMessage) ResourceSummary {
	summary := ResourceSummary{}
	for _, msg := range messages {
		switch msg.Severity {
		case "ERROR":
			summary.ErrorCount++
		case "WARNING":
			summary.WarningCount++
		case "INFO":
			summary.InfoCount++
		}
	}
	return summary
}

// newLintMetadata creates metadata for the lint output
func newLintMetadata(configFile, helmVersion, preflightVersion, supportBundleVersion, embeddedClusterVersion, kotsVersion, cliVersion string) LintMetadata {
	return LintMetadata{
		Timestamp:              time.Now().UTC().Format(time.RFC3339),
		ConfigFile:             configFile,
		HelmVersion:            helmVersion,
		PreflightVersion:       preflightVersion,
		SupportBundleVersion:   supportBundleVersion,
		EmbeddedClusterVersion: embeddedClusterVersion,
		KotsVersion:            kotsVersion,
		CLIVersion:             cliVersion,
	}
}

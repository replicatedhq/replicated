package cmd

import (
	"time"

	"github.com/replicatedhq/replicated/pkg/imageextract"
	"github.com/replicatedhq/replicated/pkg/lint2"
)

// JSONLintOutput represents the complete JSON output structure for lint results
type JSONLintOutput struct {
	Metadata         LintMetadata          `json:"metadata"`
	HelmResults      *HelmLintResults      `json:"helm_results,omitempty"`
	PreflightResults *PreflightLintResults `json:"preflight_results,omitempty"`
	Summary          LintSummary           `json:"summary"`
	Images           *ImageExtractResults  `json:"images,omitempty"` // Only if --verbose
}

// LintMetadata contains execution context and environment information
type LintMetadata struct {
	Timestamp   string `json:"timestamp"`
	ConfigFile  string `json:"config_file"`
	HelmVersion string `json:"helm_version,omitempty"`
	CLIVersion  string `json:"cli_version"`
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
func newLintMetadata(configFile, helmVersion, cliVersion string) LintMetadata {
	return LintMetadata{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		ConfigFile:  configFile,
		HelmVersion: helmVersion,
		CLIVersion:  cliVersion,
	}
}

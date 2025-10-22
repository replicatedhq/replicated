// Package lint2 provides linting functionality for Replicated resources.
// It supports linting Helm charts via helm lint and Preflight specs via preflight lint.
// Each linter executes the appropriate tool binary and parses the output into structured results.
package lint2

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/replicatedhq/replicated/pkg/tools"
	"gopkg.in/yaml.v3"
)

// PreflightLintResult represents the JSON output from preflight lint
type PreflightLintResult struct {
	Results []PreflightFileResult `json:"results"`
}

type PreflightFileResult struct {
	FilePath string               `json:"filePath"`
	Errors   []PreflightLintIssue `json:"errors"`
	Warnings []PreflightLintIssue `json:"warnings"`
	Infos    []PreflightLintIssue `json:"infos"`
}

type PreflightLintIssue struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// LintPreflight executes preflight lint on the given spec, with optional template rendering
// If valuesPath is provided, the spec is rendered with chart values and builder values before linting
func LintPreflight(
	ctx context.Context,
	specPath string,
	valuesPath string,
	chartName string,
	chartVersion string,
	helmChartManifests map[string]*HelmChartManifest,
	preflightVersion string,
) (*LintResult, error) {
	// If no valuesPath, lint directly (no templating needed)
	if valuesPath == "" {
		return lintPreflightDirect(ctx, specPath, preflightVersion)
	}

	// Templated preflight - render with builder values first
	return lintPreflightWithTemplating(ctx, specPath, valuesPath, chartName, chartVersion, helmChartManifests, preflightVersion)
}

// lintPreflightDirect lints a preflight spec without template rendering (current behavior)
func lintPreflightDirect(ctx context.Context, specPath string, preflightVersion string) (*LintResult, error) {
	// Use resolver to get preflight binary
	resolver := tools.NewResolver()
	preflightPath, err := resolver.Resolve(ctx, tools.ToolPreflight, preflightVersion)
	if err != nil {
		return nil, fmt.Errorf("resolving preflight: %w", err)
	}

	// Defensive check: validate spec path exists
	// Note: specs are validated during config parsing, but we check again here
	// since LintPreflight is a public function that could be called directly
	if _, err := os.Stat(specPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("preflight spec path does not exist: %s", specPath)
		}
		return nil, fmt.Errorf("failed to access preflight spec path: %w", err)
	}

	// Execute preflight lint with JSON output for easier parsing
	cmd := exec.CommandContext(ctx, preflightPath, "lint", "--format", "json", specPath)
	output, err := cmd.CombinedOutput()

	// preflight lint returns exit code 2 if there are errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the JSON output
	messages, parseErr := ParsePreflightOutput(outputStr)
	if parseErr != nil {
		// If we can't parse the output, return both the parse error and original error
		if err != nil {
			return nil, fmt.Errorf("preflight lint failed and output parsing failed: %w\nParse error: %v\nOutput: %s", err, parseErr, outputStr)
		}
		return nil, fmt.Errorf("failed to parse preflight lint output: %w\nOutput: %s", parseErr, outputStr)
	}

	// Determine success based on exit code
	// Exit code 0 = no errors, exit code 2 = validation errors
	success := err == nil

	return &LintResult{
		Success:  success,
		Messages: messages,
	}, nil
}

// lintPreflightWithTemplating renders a templated preflight spec with builder values, then lints it
func lintPreflightWithTemplating(
	ctx context.Context,
	specPath string,
	valuesPath string,
	chartName string,
	chartVersion string,
	helmChartManifests map[string]*HelmChartManifest,
	preflightVersion string,
) (*LintResult, error) {
	// Look up builder values from HelmChart manifest
	key := fmt.Sprintf("%s:%s", chartName, chartVersion)
	helmChart, found := helmChartManifests[key]
	if !found {
		return nil, fmt.Errorf("no HelmChart manifest found for chart %q (required for templated preflights)\n"+
			"Check that your manifests paths include the HelmChart definition", key)
	}

	// Use resolver to get preflight binary
	resolver := tools.NewResolver()
	preflightPath, err := resolver.Resolve(ctx, tools.ToolPreflight, preflightVersion)
	if err != nil {
		return nil, fmt.Errorf("resolving preflight: %w", err)
	}

	// Create temp file for builder values with secure permissions
	builderFile, err := os.CreateTemp("", "replicated-builder-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for builder values: %w", err)
	}
	builderValuesPath := builderFile.Name()
	defer func() {
		if err := os.Remove(builderValuesPath); err != nil && !os.IsNotExist(err) {
			// Log warning but don't fail - cleanup is best effort
			fmt.Fprintf(os.Stderr, "Warning: failed to cleanup builder values temp file %s: %v\n", builderValuesPath, err)
		}
	}()

	// Set restrictive permissions (owner read/write only) for security
	if err := os.Chmod(builderValuesPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set permissions on builder values: %w", err)
	}

	// Write builder values as YAML
	builderYAML, err := yaml.Marshal(helmChart.BuilderValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal builder values: %w", err)
	}
	if _, err := builderFile.Write(builderYAML); err != nil {
		return nil, fmt.Errorf("failed to write builder values: %w", err)
	}
	builderFile.Close()

	// Create temp file for rendered output with secure permissions
	renderedFile, err := os.CreateTemp("", "replicated-rendered-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for rendered output: %w", err)
	}
	renderedPath := renderedFile.Name()
	renderedFile.Close() // Close immediately, preflight will write to it
	defer func() {
		if err := os.Remove(renderedPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to cleanup rendered temp file %s: %v\n", renderedPath, err)
		}
	}()

	// Set restrictive permissions
	if err := os.Chmod(renderedPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set permissions on rendered output: %w", err)
	}

	// Render template with both values files (builder overrides chart)
	templateArgs := []string{
		"template",
		specPath,
		"--values", valuesPath,        // Chart values first
		"--values", builderValuesPath, // Builder overrides second
		"--output", renderedPath,
	}

	templateCmd := exec.CommandContext(ctx, preflightPath, templateArgs...)
	if output, err := templateCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to render preflight template:\nPreflight: %s\nValues: %s\nBuilder values: chart %q\n\nHint: Check for invalid template expressions ({{ ... }})\n\nTemplate error: %w\nOutput: %s",
			specPath, valuesPath, key, err, string(output))
	}

	// Lint the rendered spec
	return lintPreflightDirect(ctx, renderedPath, preflightVersion)
}

// ParsePreflightOutput parses preflight lint JSON output into structured messages.
// Uses the common troubleshoot.sh JSON parsing infrastructure.
func ParsePreflightOutput(output string) ([]LintMessage, error) {
	result, err := parseTroubleshootJSON[PreflightLintIssue](output)
	if err != nil {
		return nil, err
	}
	return convertTroubleshootResultToMessages(result), nil
}

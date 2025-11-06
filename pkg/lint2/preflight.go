// Package lint2 provides linting functionality for Replicated resources.
// It supports linting Helm charts via helm lint and Preflight specs via preflight lint.
// Each linter executes the appropriate tool binary and parses the output into structured results.
package lint2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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

// LintPreflight executes preflight lint on the given spec path and returns structured results.
// The preflight CLI tool handles template rendering and validation internally.
// For v1beta3 specs, we validate that HelmChart manifests exist and extract builder values.
func LintPreflight(
	ctx context.Context,
	specPath string,
	valuesPath string,
	chartName string,
	chartVersion string,
	helmChartManifests map[string]*HelmChartManifest,
	preflightVersion string,
) (*LintResult, error) {
	// Resolve preflight binary (supports REPLICATED_PREFLIGHT_PATH override for development)
	preflightPath, err := resolveLinterBinary(ctx, tools.ToolPreflight, preflightVersion, "REPLICATED_PREFLIGHT_PATH")
	if err != nil {
		return nil, err
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

	// Check if this is a v1beta3 preflight that requires HelmChart manifests
	isV1Beta3, err := isPreflightV1Beta3(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check preflight version: %w", err)
	}

	if isV1Beta3 && len(helmChartManifests) == 0 {
		return nil, fmt.Errorf("v1beta3 preflight spec requires HelmChart manifests\nCheck that your manifests paths include the HelmChart definition")
	}

	// Build command arguments
	args := []string{"lint", "--format", "json"}

	// Add chart values file if provided from config
	if valuesPath != "" {
		args = append(args, "--values", valuesPath)
	}

	// Extract and add builder values if HelmChart manifest exists
	var builderValuesPath string
	if chartName != "" && chartVersion != "" {
		key := fmt.Sprintf("%s:%s", chartName, chartVersion)
		if helmChart, found := helmChartManifests[key]; found && helmChart.BuilderValues != nil {
			// Create temp file for builder values
			builderFile, err := os.CreateTemp("", "replicated-builder-*.yaml")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file for builder values: %w", err)
			}
			builderValuesPath = builderFile.Name()
			defer func() {
				if err := os.Remove(builderValuesPath); err != nil && !os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "Warning: failed to cleanup builder values temp file %s: %v\n", builderValuesPath, err)
				}
			}()

			// Write builder values as YAML
			builderYAML, err := yaml.Marshal(helmChart.BuilderValues)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal builder values: %w", err)
			}
			if _, err := builderFile.Write(builderYAML); err != nil {
				return nil, fmt.Errorf("failed to write builder values: %w", err)
			}
			builderFile.Close()

			// Add builder values to command
			args = append(args, "--values", builderValuesPath)
		}
	}

	args = append(args, specPath)

	// Execute preflight lint
	cmd := exec.CommandContext(ctx, preflightPath, args...)
	output, err := cmd.CombinedOutput()

	// preflight lint returns exit code 2 if there are errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the JSON output
	messages, parseErr := parsePreflightOutput(outputStr)
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

// isPreflightV1Beta3 checks if a preflight spec is apiVersion v1beta3
// Tries YAML parsing first for accurate detection, falls back to string matching for templated specs
// Scans ALL documents in multi-document YAML files
func isPreflightV1Beta3(specPath string) (bool, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return false, fmt.Errorf("failed to read spec file: %w", err)
	}

	// Strategy 1: Try YAML parsing for accurate detection (multi-document aware)
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Iterate through all documents in the file
	for {
		var doc struct {
			APIVersion string `yaml:"apiVersion"`
			Kind       string `yaml:"kind"`
		}

		err := decoder.Decode(&doc)
		if err != nil {
			if err == io.EOF {
				// Reached end of file without finding v1beta3 Preflight
				break
			}
			// Parse error - fall through to string matching strategy
			break
		}

		// Check if this document is a v1beta3 Preflight
		if doc.Kind == "Preflight" && strings.Contains(doc.APIVersion, "v1beta3") {
			return true, nil
		}
	}

	// Strategy 2: Fall back to string matching for templated specs (with {{ }} syntax)
	// This handles specs that contain Helm template expressions and aren't valid YAML yet
	content := string(data)
	hasPreflightKind := strings.Contains(content, "kind: Preflight") || strings.Contains(content, "kind:Preflight")
	hasV1Beta3 := strings.Contains(content, "v1beta3")

	return hasPreflightKind && hasV1Beta3, nil
}

// parsePreflightOutput parses preflight lint JSON output into structured messages.
// Uses the common troubleshoot.sh JSON parsing infrastructure.
func parsePreflightOutput(output string) ([]LintMessage, error) {
	result, err := parseTroubleshootJSON[PreflightLintIssue](output)
	if err != nil {
		return nil, err
	}
	return convertTroubleshootResultToMessages(result), nil
}

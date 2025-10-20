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

// LintPreflight executes preflight lint on the given spec path and returns structured results
func LintPreflight(ctx context.Context, specPath string, preflightVersion string) (*LintResult, error) {
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

// ParsePreflightOutput parses preflight lint JSON output into structured messages.
// Uses the common troubleshoot.sh JSON parsing infrastructure.
func ParsePreflightOutput(output string) ([]LintMessage, error) {
	result, err := parseTroubleshootJSON[PreflightLintIssue](output)
	if err != nil {
		return nil, err
	}
	return convertTroubleshootResultToMessages(result), nil
}

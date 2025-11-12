package lint2

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// SupportBundleLintResult represents the JSON output from support-bundle lint
// This structure mirrors PreflightLintResult since both tools come from the same
// troubleshoot repository and share the same validation infrastructure.
type SupportBundleLintResult struct {
	Results []SupportBundleFileResult `json:"results"`
}

type SupportBundleFileResult struct {
	FilePath string                   `json:"filePath"`
	Errors   []SupportBundleLintIssue `json:"errors"`
	Warnings []SupportBundleLintIssue `json:"warnings"`
	Infos    []SupportBundleLintIssue `json:"infos"`
}

type SupportBundleLintIssue struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// LintSupportBundle executes support-bundle lint on the given spec path and returns structured results
func LintSupportBundle(ctx context.Context, specPath string, sbVersion string) (*LintResult, error) {
	// Resolve support-bundle binary (supports REPLICATED_SUPPORT_BUNDLE_PATH override for development)
	sbPath, err := resolveLinterBinary(ctx, tools.ToolSupportBundle, sbVersion, "REPLICATED_SUPPORT_BUNDLE_PATH")
	if err != nil {
		return nil, err
	}

	// Defensive check: validate spec path exists
	// Note: specs are validated during config parsing, but we check again here
	// since LintSupportBundle is a public function that could be called directly
	if _, err := os.Stat(specPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("support bundle spec path does not exist: %s", specPath)
		}
		return nil, fmt.Errorf("failed to access support bundle spec path: %w", err)
	}

	// Execute support-bundle lint with JSON output for easier parsing
	// Note: The support-bundle lint command may be in active development.
	// If it's currently broken, this will fail, but the infrastructure is ready
	// for when the command is fixed.
	cmd := exec.CommandContext(ctx, sbPath, "lint", "--format", "json", specPath)
	output, cmdErr := cmd.CombinedOutput()

	// support-bundle lint returns exit code 2 if there are errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the JSON output
	messages, parseErr := parseSupportBundleOutput(outputStr)
	if parseErr != nil {
		// If we can't parse the output, return both the parse error and original error
		if cmdErr != nil {
			return nil, fmt.Errorf("support-bundle lint failed and output parsing failed: %w\nParse error: %v\nOutput: %s", cmdErr, parseErr, outputStr)
		}
		return nil, fmt.Errorf("failed to parse support-bundle lint output: %w\nOutput: %s", parseErr, outputStr)
	}

	// Success when linter binary exits cleanly (exit code 0)
	lintSuccess := (cmdErr == nil)

	return &LintResult{
		Success:  lintSuccess,
		Messages: messages,
	}, nil
}

// parseSupportBundleOutput parses support-bundle lint JSON output into structured messages.
// Uses the common troubleshoot.sh JSON parsing infrastructure.
func parseSupportBundleOutput(output string) ([]LintMessage, error) {
	result, err := parseTroubleshootJSON[SupportBundleLintIssue](output)
	if err != nil {
		return nil, err
	}
	return convertTroubleshootResultToMessages(result), nil
}

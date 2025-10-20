package lint2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// SupportBundleLintResult represents the JSON output from support-bundle lint
// This structure mirrors PreflightLintResult since both tools come from the same
// troubleshoot repository and share the same validation infrastructure.
type SupportBundleLintResult struct {
	Results []SupportBundleFileResult `json:"results"`
}

type SupportBundleFileResult struct {
	FilePath string                     `json:"filePath"`
	Errors   []SupportBundleLintIssue   `json:"errors"`
	Warnings []SupportBundleLintIssue   `json:"warnings"`
	Infos    []SupportBundleLintIssue   `json:"infos"`
}

type SupportBundleLintIssue struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// LintSupportBundle executes support-bundle lint on the given spec path and returns structured results
func LintSupportBundle(ctx context.Context, specPath string, sbVersion string) (*LintResult, error) {
	// Use resolver to get support-bundle binary
	resolver := tools.NewResolver()
	sbPath, err := resolver.Resolve(ctx, tools.ToolSupportBundle, sbVersion)
	if err != nil {
		return nil, fmt.Errorf("resolving support-bundle: %w", err)
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
	output, err := cmd.CombinedOutput()

	// support-bundle lint returns exit code 2 if there are errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the JSON output
	messages, parseErr := ParseSupportBundleOutput(outputStr)
	if parseErr != nil {
		// If we can't parse the output, return both the parse error and original error
		if err != nil {
			return nil, fmt.Errorf("support-bundle lint failed and output parsing failed: %w\nParse error: %v\nOutput: %s", err, parseErr, outputStr)
		}
		return nil, fmt.Errorf("failed to parse support-bundle lint output: %w\nOutput: %s", parseErr, outputStr)
	}

	// Determine success based on exit code
	// Exit code 0 = no errors, exit code 2 = validation errors
	success := err == nil

	return &LintResult{
		Success:  success,
		Messages: messages,
	}, nil
}

// ParseSupportBundleOutput parses support-bundle lint JSON output into structured messages
func ParseSupportBundleOutput(output string) ([]LintMessage, error) {
	// The support-bundle binary may output "Error:" on stderr after the JSON when there are issues.
	// This gets combined with stdout by CombinedOutput(), so we extract just the JSON portion.
	// We use Index (first '{') and LastIndex (last '}') to handle nested objects in the JSON structure.
	// This approach works reliably because the top-level JSON object encompasses all nested content.
	startIdx := strings.Index(output, "{")
	if startIdx == -1 {
		return nil, fmt.Errorf("no JSON found in output")
	}

	endIdx := strings.LastIndex(output, "}")
	if endIdx == -1 || endIdx < startIdx {
		return nil, fmt.Errorf("malformed JSON in output")
	}

	jsonOutput := output[startIdx : endIdx+1]

	var result SupportBundleLintResult
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON: %w", err)
	}

	var messages []LintMessage

	// Process all file results
	for _, fileResult := range result.Results {
		// Process errors
		for _, issue := range fileResult.Errors {
			messages = append(messages, LintMessage{
				Severity: "ERROR",
				Path:     fileResult.FilePath,
				Message:  formatSupportBundleMessage(issue),
			})
		}

		// Process warnings
		for _, issue := range fileResult.Warnings {
			messages = append(messages, LintMessage{
				Severity: "WARNING",
				Path:     fileResult.FilePath,
				Message:  formatSupportBundleMessage(issue),
			})
		}

		// Process infos
		for _, issue := range fileResult.Infos {
			messages = append(messages, LintMessage{
				Severity: "INFO",
				Path:     fileResult.FilePath,
				Message:  formatSupportBundleMessage(issue),
			})
		}
	}

	return messages, nil
}

// formatSupportBundleMessage formats a support bundle issue into a readable message
func formatSupportBundleMessage(issue SupportBundleLintIssue) string {
	msg := issue.Message

	// Add line number if available
	if issue.Line > 0 {
		msg = fmt.Sprintf("line %d: %s", issue.Line, msg)
	}

	// Add field information if available
	if issue.Field != "" {
		msg = fmt.Sprintf("%s (field: %s)", msg, issue.Field)
	}

	return msg
}

package lint2

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Implement LintIssue for PreflightLintIssue
func (i PreflightLintIssue) GetLine() int       { return i.Line }
func (i PreflightLintIssue) GetColumn() int     { return i.Column }
func (i PreflightLintIssue) GetMessage() string { return i.Message }
func (i PreflightLintIssue) GetField() string   { return i.Field }

// Implement LintIssue for SupportBundleLintIssue
func (i SupportBundleLintIssue) GetLine() int       { return i.Line }
func (i SupportBundleLintIssue) GetColumn() int     { return i.Column }
func (i SupportBundleLintIssue) GetMessage() string { return i.Message }
func (i SupportBundleLintIssue) GetField() string   { return i.Field }

// parseLintJSON extracts and decodes JSON from linting tool output.
// The tool binaries may output "Error:" on stderr before/after the JSON when there
// are issues. This gets combined with stdout by CombinedOutput(). We search for each
// potential JSON object and try to decode it. The decoder automatically handles
// trailing garbage after valid JSON.
func parseLintJSON[T LintIssue](output string) (*LintOutput[T], error) {
	var result LintOutput[T]
	var lastErr error

	// Try to find and decode JSON starting from each { in the output
	// This handles cases where error messages contain braces before the actual JSON
	searchOffset := 0
	for {
		idx := strings.Index(output[searchOffset:], "{")
		if idx == -1 {
			break
		}

		startIdx := searchOffset + idx
		decoder := json.NewDecoder(strings.NewReader(output[startIdx:]))
		err := decoder.Decode(&result)
		if err == nil {
			break
		}

		lastErr = err
		searchOffset = startIdx + 1
	}

	if result.Results == nil {
		if lastErr != nil {
			return nil, fmt.Errorf("no valid JSON found in output: %w", lastErr)
		}
		return nil, fmt.Errorf("no JSON found in output")
	}

	return &result, nil
}

// formatLintMessage formats a lint issue into a readable message string.
func formatLintMessage(issue LintIssue) string {
	msg := issue.GetMessage()

	if issue.GetLine() > 0 {
		msg = fmt.Sprintf("line %d: %s", issue.GetLine(), msg)
	}

	if issue.GetField() != "" {
		msg = fmt.Sprintf("%s (field: %s)", msg, issue.GetField())
	}

	return msg
}

// convertLintOutputToMessages converts a LintOutput into LintMessages.
func convertLintOutputToMessages[T LintIssue](result *LintOutput[T]) []LintMessage {
	var messages []LintMessage

	for _, fileResult := range result.Results {
		for _, issue := range fileResult.Errors {
			messages = append(messages, LintMessage{
				Severity: "ERROR",
				Path:     fileResult.FilePath,
				Message:  formatLintMessage(issue),
			})
		}
		for _, issue := range fileResult.Warnings {
			messages = append(messages, LintMessage{
				Severity: "WARNING",
				Path:     fileResult.FilePath,
				Message:  formatLintMessage(issue),
			})
		}
		for _, issue := range fileResult.Info {
			messages = append(messages, LintMessage{
				Severity: "INFO",
				Path:     fileResult.FilePath,
				Message:  formatLintMessage(issue),
			})
		}
	}

	return messages
}

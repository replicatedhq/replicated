package lint2

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TroubleshootIssue is an interface that both PreflightLintIssue and
// SupportBundleLintIssue satisfy, allowing common formatting logic.
// Both tools come from the troubleshoot.sh repository and share the same
// validation infrastructure and output format.
type TroubleshootIssue interface {
	GetLine() int
	GetColumn() int
	GetMessage() string
	GetField() string
}

// Implement TroubleshootIssue interface for PreflightLintIssue
func (i PreflightLintIssue) GetLine() int       { return i.Line }
func (i PreflightLintIssue) GetColumn() int     { return i.Column }
func (i PreflightLintIssue) GetMessage() string { return i.Message }
func (i PreflightLintIssue) GetField() string   { return i.Field }

// Implement TroubleshootIssue interface for SupportBundleLintIssue
func (i SupportBundleLintIssue) GetLine() int       { return i.Line }
func (i SupportBundleLintIssue) GetColumn() int     { return i.Column }
func (i SupportBundleLintIssue) GetMessage() string { return i.Message }
func (i SupportBundleLintIssue) GetField() string   { return i.Field }

// TroubleshootFileResult represents the common structure for file-level results
// from troubleshoot.sh linting tools (preflight, support-bundle, etc.)
type TroubleshootFileResult[T TroubleshootIssue] struct {
	FilePath string `json:"filePath"`
	Errors   []T    `json:"errors"`
	Warnings []T    `json:"warnings"`
	Infos    []T    `json:"infos"`
}

// TroubleshootLintResult represents the common JSON structure for
// troubleshoot.sh linting tool output
type TroubleshootLintResult[T TroubleshootIssue] struct {
	Results []TroubleshootFileResult[T] `json:"results"`
}

// parseTroubleshootJSON extracts and decodes JSON from troubleshoot tool output.
// The tool binaries may output "Error:" on stderr before/after the JSON when there
// are issues. This gets combined with stdout by CombinedOutput(). We search for each
// potential JSON object and try to decode it. The decoder automatically handles
// trailing garbage after valid JSON.
func parseTroubleshootJSON[T TroubleshootIssue](output string) (*TroubleshootLintResult[T], error) {
	var result TroubleshootLintResult[T]
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
			// Successfully decoded JSON
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

// formatTroubleshootMessage formats a troubleshoot issue into a readable message
func formatTroubleshootMessage(issue TroubleshootIssue) string {
	msg := issue.GetMessage()

	// Add line number if available
	if issue.GetLine() > 0 {
		msg = fmt.Sprintf("line %d: %s", issue.GetLine(), msg)
	}

	// Add field information if available
	if issue.GetField() != "" {
		msg = fmt.Sprintf("%s (field: %s)", msg, issue.GetField())
	}

	return msg
}

// convertTroubleshootResultToMessages processes troubleshoot issues into LintMessages.
// This handles the common pattern of processing errors, warnings, and infos from
// troubleshoot.sh tool output.
func convertTroubleshootResultToMessages[T TroubleshootIssue](
	result *TroubleshootLintResult[T],
) []LintMessage {
	var messages []LintMessage

	for _, fileResult := range result.Results {
		// Process errors
		for _, issue := range fileResult.Errors {
			messages = append(messages, LintMessage{
				Severity: "ERROR",
				Path:     fileResult.FilePath,
				Message:  formatTroubleshootMessage(issue),
			})
		}

		// Process warnings
		for _, issue := range fileResult.Warnings {
			messages = append(messages, LintMessage{
				Severity: "WARNING",
				Path:     fileResult.FilePath,
				Message:  formatTroubleshootMessage(issue),
			})
		}

		// Process infos
		for _, issue := range fileResult.Infos {
			messages = append(messages, LintMessage{
				Severity: "INFO",
				Path:     fileResult.FilePath,
				Message:  formatTroubleshootMessage(issue),
			})
		}
	}

	return messages
}

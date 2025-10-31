package lint2

import (
	"encoding/json"
	"fmt"
	"strings"
)

// extractJSONFromOutput extracts a JSON object from command output that may
// contain error messages before or after the JSON. Uses brace counting to
// find the complete JSON object boundaries while properly handling strings
// and escape sequences.
//
// This is used by all linters (preflight, support-bundle, embedded-cluster) to
// handle output where the tool prints "ERROR: ..." or other messages alongside
// the JSON results.
//
// The function searches for each '{' in the output, extracts a complete JSON object
// using brace counting, and attempts to decode it. This handles cases where error
// messages contain braces (e.g., "Error: failed {something}") before the actual JSON.
func extractJSONFromOutput(output string) (string, error) {
	if output == "" {
		return "", fmt.Errorf("empty output")
	}

	// Try to find and extract valid JSON starting from each { in the output
	searchOffset := 0
	var lastErr error

	for {
		// Find next opening brace
		idx := strings.Index(output[searchOffset:], "{")
		if idx == -1 {
			break
		}

		startIdx := searchOffset + idx

		// Use brace counting to find matching closing brace
		jsonEnd := -1
		braceCount := 0
		inString := false
		escaped := false

		for i := startIdx; i < len(output); i++ {
			ch := rune(output[i])

			if escaped {
				escaped = false
				continue
			}

			if ch == '\\' {
				escaped = true
				continue
			}

			if ch == '"' {
				inString = !inString
				continue
			}

			if inString {
				continue
			}

			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					jsonEnd = i + 1
					break
				}
			}
		}

		if jsonEnd == -1 {
			// Unclosed braces from this position, try next {
			searchOffset = startIdx + 1
			lastErr = fmt.Errorf("unclosed JSON object")
			continue
		}

		// We found a complete brace-balanced substring, try to validate it as JSON
		candidate := output[startIdx:jsonEnd]

		// Quick validation: try to unmarshal to check if it's valid JSON
		var testObj interface{}
		if err := json.Unmarshal([]byte(candidate), &testObj); err == nil {
			// Valid JSON found
			return candidate, nil
		} else {
			lastErr = err
		}

		// Not valid JSON, try next {
		searchOffset = startIdx + 1
	}

	if lastErr != nil {
		return "", fmt.Errorf("no valid JSON found in output: %w", lastErr)
	}
	return "", fmt.Errorf("no JSON object found in output")
}

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
// are issues. This gets combined with stdout by CombinedOutput().
func parseTroubleshootJSON[T TroubleshootIssue](output string) (*TroubleshootLintResult[T], error) {
	// Extract clean JSON from output that may contain error messages
	jsonStr, err := extractJSONFromOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from output: %w", err)
	}

	// Decode the JSON into the result structure
	var result TroubleshootLintResult[T]
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
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

package lint2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// ECLintOutput represents the JSON output from embedded-cluster lint
type ECLintOutput struct {
	Files []ECLintFileResult `json:"files"`
}

type ECLintFileResult struct {
	Path     string         `json:"path"`
	Valid    bool           `json:"valid"`
	Errors   []ECLintIssue  `json:"errors,omitempty"`
	Warnings []ECLintIssue  `json:"warnings,omitempty"`
	Infos    []ECLintIssue  `json:"infos,omitempty"`
}

type ECLintIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// LintEmbeddedCluster executes embedded-cluster lint on the given config path and returns structured results
//
// The caller (cli/cmd/lint.go) sets the following environment variables for ALL linter binaries:
//   - REPLICATED_APP: Canonical app ID for vendor portal context (if available)
//   - REPLICATED_API_TOKEN: Authentication token for vendor portal API
//   - REPLICATED_API_ORIGIN: API endpoint (e.g., https://api.replicated.com)
//
// These environment variables enable all linters (helm, preflight, support-bundle, embedded-cluster)
// to make vendor portal API calls for enhanced validation capabilities.
func LintEmbeddedCluster(ctx context.Context, configPath string, ecVersion string) (*LintResult, error) {
	// Check for local binary override (for development)
	ecPath := os.Getenv("REPLICATED_EMBEDDED_CLUSTER_PATH")
	if ecPath == "" {
		// Use resolver to get embedded-cluster binary
		resolver := tools.NewResolver()
		var err error
		ecPath, err = resolver.Resolve(ctx, tools.ToolEmbeddedCluster, ecVersion)
		if err != nil {
			return nil, fmt.Errorf("resolving embedded-cluster: %w", err)
		}
	}

	// Defensive check: validate config path exists
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("embedded cluster config path does not exist: %s", configPath)
		}
		return nil, fmt.Errorf("failed to access embedded cluster config path: %w", err)
	}

	// Build command arguments
	args := []string{"lint", "--output", "json", configPath}

	// Execute embedded-cluster lint
	cmd := exec.CommandContext(ctx, ecPath, args...)
	output, err := cmd.CombinedOutput()

	// embedded-cluster lint returns non-zero exit code if there are validation errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the JSON output
	messages, parseErr := parseEmbeddedClusterOutput(outputStr)
	if parseErr != nil {
		// If we can't parse the output, return both the parse error and original error
		if err != nil {
			return nil, fmt.Errorf("embedded-cluster lint failed and output parsing failed: %w\nParse error: %v\nOutput: %s", err, parseErr, outputStr)
		}
		return nil, fmt.Errorf("failed to parse embedded-cluster lint output: %w\nOutput: %s", parseErr, outputStr)
	}

	// Determine success based on exit code
	// Exit code 0 = no errors, non-zero = validation errors
	success := err == nil

	return &LintResult{
		Success:  success,
		Messages: messages,
	}, nil
}

// parseEmbeddedClusterOutput parses embedded-cluster lint JSON output into structured messages
func parseEmbeddedClusterOutput(output string) ([]LintMessage, error) {
	if output == "" {
		return []LintMessage{}, nil
	}

	// embedded-cluster lint outputs JSON followed by "ERROR: validation failed with errors" on failures
	// We need to extract just the JSON part. Look for the closing brace of the root object.
	// The output format is: {"files":[...]}
	// Followed by optional "ERROR: ..." message
	jsonEnd := -1
	braceCount := 0
	inString := false
	escaped := false

	for i, ch := range output {
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

	if jsonEnd > 0 {
		output = output[:jsonEnd]
	}

	var ecOutput ECLintOutput
	if err := json.Unmarshal([]byte(output), &ecOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON output: %w", err)
	}

	var messages []LintMessage

	// Process each file in the output
	for _, fileResult := range ecOutput.Files {
		// Add error messages
		for _, issue := range fileResult.Errors {
			msg := issue.Message
			if issue.Field != "" {
				msg = fmt.Sprintf("%s: %s", issue.Field, issue.Message)
			}
			messages = append(messages, LintMessage{
				Severity: "error",
				Message:  msg,
				Path:     fileResult.Path,
			})
		}

		// Add warning messages
		for _, issue := range fileResult.Warnings {
			msg := issue.Message
			if issue.Field != "" {
				msg = fmt.Sprintf("%s: %s", issue.Field, issue.Message)
			}
			messages = append(messages, LintMessage{
				Severity: "warning",
				Message:  msg,
				Path:     fileResult.Path,
			})
		}

		// Add info messages
		for _, issue := range fileResult.Infos {
			msg := issue.Message
			if issue.Field != "" {
				msg = fmt.Sprintf("%s: %s", issue.Field, issue.Message)
			}
			messages = append(messages, LintMessage{
				Severity: "info",
				Message:  msg,
				Path:     fileResult.Path,
			})
		}
	}

	return messages, nil
}

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
	Path     string        `json:"path"`
	Valid    bool          `json:"valid"`
	Errors   []ECLintIssue `json:"errors,omitempty"`
	Warnings []ECLintIssue `json:"warnings,omitempty"`
	Infos    []ECLintIssue `json:"infos,omitempty"`
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
	// Defensive check: validate config path exists (before resolving binary)
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("embedded cluster config path does not exist: %s", configPath)
		}
		return nil, fmt.Errorf("failed to access embedded cluster config path: %w", err)
	}

	// Resolve embedded-cluster binary (supports REPLICATED_EMBEDDED_CLUSTER_PATH override for development)
	ecPath, err := resolveLinterBinary(ctx, tools.ToolEmbeddedCluster, ecVersion, "REPLICATED_EMBEDDED_CLUSTER_PATH")
	if err != nil {
		return nil, err
	}

	// Build command arguments
	args := []string{"lint", "--output", "json", configPath}

	// Execute embedded-cluster lint
	cmd := exec.CommandContext(ctx, ecPath, args...)
	output, cmdErr := cmd.CombinedOutput()

	// embedded-cluster lint returns non-zero exit code if there are validation errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the JSON output
	messages, parseErr := parseEmbeddedClusterOutput(outputStr)
	if parseErr != nil {
		// If we can't parse the output, return both the parse error and original error
		if cmdErr != nil {
			return nil, fmt.Errorf("embedded-cluster lint failed and output parsing failed: %w\nParse error: %v\nOutput: %s", cmdErr, parseErr, outputStr)
		}
		return nil, fmt.Errorf("failed to parse embedded-cluster lint output: %w\nOutput: %s", parseErr, outputStr)
	}

	// Success when linter binary exits cleanly (exit code 0)
	lintSuccess := (cmdErr == nil)

	return &LintResult{
		Success:  lintSuccess,
		Messages: messages,
	}, nil
}

// parseEmbeddedClusterOutput parses embedded-cluster lint JSON output into structured messages
func parseEmbeddedClusterOutput(output string) ([]LintMessage, error) {
	if output == "" {
		return []LintMessage{}, nil
	}

	// Extract clean JSON from output that may contain error messages
	// (e.g., "ERROR: validation failed with errors" after the JSON)
	jsonStr, err := extractJSONFromOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from output: %w", err)
	}

	// Decode the JSON into the embedded-cluster result structure
	var ecOutput ECLintOutput
	if err := json.Unmarshal([]byte(jsonStr), &ecOutput); err != nil {
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
				Severity: "ERROR",
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
				Severity: "WARNING",
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
				Severity: "INFO",
				Message:  msg,
				Path:     fileResult.Path,
			})
		}
	}

	return messages, nil
}

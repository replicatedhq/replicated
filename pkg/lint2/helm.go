package lint2

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/replicatedhq/replicated/pkg/tools"
)

// LintChart executes helm lint on the given chart path and returns structured results
func LintChart(ctx context.Context, chartPath string, helmVersion string) (*LintResult, error) {
	// Use resolver to get helm binary
	resolver := tools.NewResolver()
	helmPath, err := resolver.Resolve(ctx, tools.ToolHelm, helmVersion)
	if err != nil {
		return nil, fmt.Errorf("resolving helm: %w", err)
	}

	// Validate chart path exists
	if _, err := os.Stat(chartPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("chart path does not exist: %s", chartPath)
		}
		return nil, fmt.Errorf("failed to access chart path: %w", err)
	}

	// Execute helm lint
	cmd := exec.CommandContext(ctx, helmPath, "lint", chartPath)
	output, err := cmd.CombinedOutput()

	// helm lint returns non-zero exit code if there are errors,
	// but we still want to parse and display the output
	outputStr := string(output)

	// Parse the output
	messages := ParseHelmOutput(outputStr)

	// Determine success based on exit code
	// We trust helm's exit code: 0 = success, non-zero = failure
	success := err == nil

	// However, if helm failed but we got parseable output, we should
	// still return the parsed messages
	if err != nil && len(messages) == 0 {
		// If helm failed and we have no parsed messages, return the error
		return nil, fmt.Errorf("helm lint failed: %w\n%s", err, outputStr)
	}

	return &LintResult{
		Success:  success,
		Messages: messages,
	}, nil
}

// ParseHelmOutput parses helm lint output into structured messages
func ParseHelmOutput(output string) []LintMessage {
	var messages []LintMessage

	// Pattern to match: [SEVERITY] path: message
	// Example: [INFO] Chart.yaml: icon is recommended
	pattern := regexp.MustCompile(`^\[(INFO|WARNING|ERROR)\]\s+([^:]+):\s*(.+)$`)

	// Pattern for messages without path: [SEVERITY] message
	patternNoPath := regexp.MustCompile(`^\[(INFO|WARNING|ERROR)\]\s+(.+)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try pattern with path first
		if matches := pattern.FindStringSubmatch(line); matches != nil {
			messages = append(messages, LintMessage{
				Severity: matches[1],
				Path:     strings.TrimSpace(matches[2]),
				Message:  strings.TrimSpace(matches[3]),
			})
			continue
		}

		// Try pattern without path
		if matches := patternNoPath.FindStringSubmatch(line); matches != nil {
			messages = append(messages, LintMessage{
				Severity: matches[1],
				Path:     "",
				Message:  strings.TrimSpace(matches[2]),
			})
		}

		// Ignore lines that don't match (headers, summaries, etc.)
	}

	return messages
}

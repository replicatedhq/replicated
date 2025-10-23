package lint2

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/replicatedhq/replicated/pkg/tools"
	"gopkg.in/yaml.v3"
)

// LintChart executes helm lint on the given chart path and returns structured results
func LintChart(ctx context.Context, chartPath string, helmVersion string) (*LintResult, error) {
	// Use resolver to get helm binary
	resolver := tools.NewResolver()
	helmPath, err := resolver.Resolve(ctx, tools.ToolHelm, helmVersion)
	if err != nil {
		return nil, fmt.Errorf("resolving helm: %w", err)
	}

	// Defensive check: validate chart path exists
	// Note: charts are validated during config parsing, but we check again here
	// since LintChart is a public function that could be called directly
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
	messages := parseHelmOutput(outputStr)

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

// parseHelmOutput parses helm lint output into structured messages
func parseHelmOutput(output string) []LintMessage {
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

// ChartMetadata represents basic metadata from a Helm chart's Chart.yaml
type ChartMetadata struct {
	Name    string
	Version string
}

// GetChartMetadata reads Chart.yaml and returns the chart name and version
func GetChartMetadata(chartPath string) (*ChartMetadata, error) {
	chartYamlPath := filepath.Join(chartPath, "Chart.yaml")

	data, err := os.ReadFile(chartYamlPath)
	if err != nil {
		// Try Chart.yml as fallback (some charts use lowercase extension)
		chartYmlPath := filepath.Join(chartPath, "Chart.yml")
		data, err = os.ReadFile(chartYmlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read Chart.yaml or Chart.yml: %w", err)
		}
	}

	var chart struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	if chart.Name == "" {
		return nil, fmt.Errorf("chart name is empty in Chart.yaml")
	}
	if chart.Version == "" {
		return nil, fmt.Errorf("chart version is empty in Chart.yaml")
	}

	return &ChartMetadata{
		Name:    chart.Name,
		Version: chart.Version,
	}, nil
}

package cmd

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/lint2"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/spf13/cobra"
)

func (r *runners) InitLint(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint Helm charts",
		Long:  `Lint Helm charts defined in .replicated config file. This command reads chart paths from the .replicated config and executes helm lint locally on each chart.`,
	}

	cmd.RunE = r.runLint

	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) runLint(cmd *cobra.Command, args []string) error {
	// Load .replicated config using tools parser (supports monorepos)
	parser := tools.NewConfigParser()
	config, err := parser.FindAndParseConfig(".")
	if err != nil {
		return errors.Wrap(err, "failed to load .replicated config")
	}

	// Check if helm linting is enabled
	if !config.ReplLint.Linters.Helm.IsEnabled() {
		fmt.Fprintf(r.w, "Helm linting is disabled in .replicated config\n")
		return nil
	}

	// Get helm version from config
	helmVersion := tools.DefaultHelmVersion
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolHelm]; ok {
			helmVersion = v
		}
	}

	// Check if there are any charts configured
	chartPaths, err := lint2.GetChartPathsFromConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to expand chart paths")
	}

	// Lint all charts and collect results
	var allResults []*lint2.LintResult
	var allPaths []string
	hasFailure := false

	for _, chartPath := range chartPaths {
		result, err := lint2.LintChart(cmd.Context(), chartPath, helmVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to lint chart: %s", chartPath)
		}

		allResults = append(allResults, result)
		allPaths = append(allPaths, chartPath)

		if !result.Success {
			hasFailure = true
		}
	}

	// Display results for all charts
	if err := displayAllLintResults(r.w, allPaths, allResults); err != nil {
		return errors.Wrap(err, "failed to display lint results")
	}

	// Flush the tab writer
	if err := r.w.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush output")
	}

	// Return error if any chart failed linting
	if hasFailure {
		return errors.New("linting failed for one or more charts")
	}

	return nil
}

type chartSummary struct {
	errorCount   int
	warningCount int
	infoCount    int
}

func displayAllLintResults(w io.Writer, chartPaths []string, results []*lint2.LintResult) error {
	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	totalChartsFailed := 0

	// Display results for each chart
	for i, result := range results {
		chartPath := chartPaths[i]
		summary := displaySingleChartResult(w, chartPath, result)

		totalErrors += summary.errorCount
		totalWarnings += summary.warningCount
		totalInfo += summary.infoCount

		if !result.Success {
			totalChartsFailed++
		}
	}

	// Print overall summary if multiple charts
	if len(results) > 1 {
		displayOverallSummary(w, len(results), totalChartsFailed, totalErrors, totalWarnings, totalInfo)
	}

	return nil
}

func displaySingleChartResult(w io.Writer, chartPath string, result *lint2.LintResult) chartSummary {
	// Print header for this chart
	fmt.Fprintf(w, "==> Linting %s\n\n", chartPath)

	// Print messages
	if len(result.Messages) == 0 {
		fmt.Fprintf(w, "No issues found\n")
	} else {
		for _, msg := range result.Messages {
			displayLintMessage(w, msg)
		}
	}

	// Count messages by severity
	summary := countMessagesBySeverity(result.Messages)

	// Print per-chart summary
	fmt.Fprintf(w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
		chartPath, summary.errorCount, summary.warningCount, summary.infoCount)

	// Print per-chart status
	if result.Success {
		fmt.Fprintf(w, "Status: Passed\n\n")
	} else {
		fmt.Fprintf(w, "Status: Failed\n\n")
	}

	return summary
}

func displayLintMessage(w io.Writer, msg lint2.LintMessage) {
	if msg.Path != "" {
		fmt.Fprintf(w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
	} else {
		fmt.Fprintf(w, "[%s] %s\n", msg.Severity, msg.Message)
	}
}

func countMessagesBySeverity(messages []lint2.LintMessage) chartSummary {
	summary := chartSummary{}
	for _, msg := range messages {
		switch msg.Severity {
		case "ERROR":
			summary.errorCount++
		case "WARNING":
			summary.warningCount++
		case "INFO":
			summary.infoCount++
		}
	}
	return summary
}

func displayOverallSummary(w io.Writer, totalCharts, failedCharts, totalErrors, totalWarnings, totalInfo int) {
	fmt.Fprintf(w, "==> Overall Summary\n")
	fmt.Fprintf(w, "Charts linted: %d\n", totalCharts)
	fmt.Fprintf(w, "Charts passed: %d\n", totalCharts-failedCharts)
	fmt.Fprintf(w, "Charts failed: %d\n", failedCharts)
	fmt.Fprintf(w, "Total errors: %d\n", totalErrors)
	fmt.Fprintf(w, "Total warnings: %d\n", totalWarnings)
	fmt.Fprintf(w, "Total info: %d\n", totalInfo)

	if failedCharts > 0 {
		fmt.Fprintf(w, "\nOverall Status: Failed\n")
	} else {
		fmt.Fprintf(w, "\nOverall Status: Passed\n")
	}
}

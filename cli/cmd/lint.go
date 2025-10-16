package cmd

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/lint2"
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
	// Load .replicated config
	config, err := lint2.LoadReplicatedConfig()
	if err != nil {
		return errors.Wrap(err, "failed to load .replicated config")
	}

	// Check if helm linting is enabled
	helmConfig, exists := config.ReplLint.Linters["helm"]
	if exists && !helmConfig.Enabled {
		fmt.Fprintf(r.w, "Helm linting is disabled in .replicated config\n")
		return nil
	}

	// Check if there are any charts configured
	if len(config.Charts) == 0 {
		return errors.New("no charts found in .replicated config")
	}

	// Expand chart paths (handle globs)
	chartPaths, err := lint2.ExpandChartPaths(config.Charts)
	if err != nil {
		return errors.Wrap(err, "failed to expand chart paths")
	}

	// Lint all charts and collect results
	var allResults []*lint2.LintResult
	var allPaths []string
	hasFailure := false

	for _, chartPath := range chartPaths {
		result, err := lint2.LintChart(cmd.Context(), chartPath)
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

func displayAllLintResults(w io.Writer, chartPaths []string, results []*lint2.LintResult) error {
	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	totalChartsFailed := 0

	// Display results for each chart
	for i, result := range results {
		chartPath := chartPaths[i]

		// Print header for this chart
		fmt.Fprintf(w, "==> Linting %s\n\n", chartPath)

		// Print messages
		if len(result.Messages) == 0 {
			fmt.Fprintf(w, "No issues found\n")
		} else {
			for _, msg := range result.Messages {
				if msg.Path != "" {
					fmt.Fprintf(w, "[%s] %s: %s\n", msg.Severity, msg.Path, msg.Message)
				} else {
					fmt.Fprintf(w, "[%s] %s\n", msg.Severity, msg.Message)
				}
			}
		}

		// Count messages by severity for this chart
		errorCount := 0
		warningCount := 0
		infoCount := 0
		for _, msg := range result.Messages {
			switch msg.Severity {
			case "ERROR":
				errorCount++
				totalErrors++
			case "WARNING":
				warningCount++
				totalWarnings++
			case "INFO":
				infoCount++
				totalInfo++
			}
		}

		// Print per-chart summary
		fmt.Fprintf(w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n", chartPath, errorCount, warningCount, infoCount)

		// Print per-chart status
		if result.Success {
			fmt.Fprintf(w, "Status: Passed\n\n")
		} else {
			fmt.Fprintf(w, "Status: Failed\n\n")
			totalChartsFailed++
		}
	}

	// Print overall summary if multiple charts
	if len(results) > 1 {
		fmt.Fprintf(w, "==> Overall Summary\n")
		fmt.Fprintf(w, "Charts linted: %d\n", len(results))
		fmt.Fprintf(w, "Charts passed: %d\n", len(results)-totalChartsFailed)
		fmt.Fprintf(w, "Charts failed: %d\n", totalChartsFailed)
		fmt.Fprintf(w, "Total errors: %d\n", totalErrors)
		fmt.Fprintf(w, "Total warnings: %d\n", totalWarnings)
		fmt.Fprintf(w, "Total info: %d\n", totalInfo)

		if totalChartsFailed > 0 {
			fmt.Fprintf(w, "\nOverall Status: Failed\n")
		} else {
			fmt.Fprintf(w, "\nOverall Status: Passed\n")
		}
	}

	return nil
}

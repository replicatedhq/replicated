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
		Use:          "lint",
		Short:        "Lint Helm charts and Preflight specs",
		Long:         `Lint Helm charts and Preflight specs defined in .replicated config file. This command reads paths from the .replicated config and executes linting locally on each resource.`,
		SilenceUsage: true,
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

	hasFailure := false

	// Lint Helm charts if enabled
	if config.ReplLint.Linters.Helm.IsEnabled() {
		helmFailed, err := r.lintHelmCharts(cmd, config)
		if err != nil {
			return err
		}
		if helmFailed {
			hasFailure = true
		}
	} else {
		fmt.Fprintf(r.w, "Helm linting is disabled in .replicated config\n\n")
	}

	// Lint Preflight specs if enabled
	if config.ReplLint.Linters.Preflight.IsEnabled() {
		preflightFailed, err := r.lintPreflightSpecs(cmd, config)
		if err != nil {
			return err
		}
		if preflightFailed {
			hasFailure = true
		}
	} else {
		fmt.Fprintf(r.w, "Preflight linting is disabled in .replicated config\n\n")
	}

	// Flush the tab writer
	if err := r.w.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush output")
	}

	// Return error if any linting failed
	if hasFailure {
		return errors.New("linting failed")
	}

	return nil
}

func (r *runners) lintHelmCharts(cmd *cobra.Command, config *tools.Config) (bool, error) {
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
		return false, errors.Wrap(err, "failed to expand chart paths")
	}

	// Lint all charts and collect results
	var allResults []*lint2.LintResult
	var allPaths []string
	hasFailure := false

	for _, chartPath := range chartPaths {
		result, err := lint2.LintChart(cmd.Context(), chartPath, helmVersion)
		if err != nil {
			return false, errors.Wrapf(err, "failed to lint chart: %s", chartPath)
		}

		allResults = append(allResults, result)
		allPaths = append(allPaths, chartPath)

		if !result.Success {
			hasFailure = true
		}
	}

	// Display results for all charts
	if err := displayAllLintResults(r.w, "chart", allPaths, allResults); err != nil {
		return false, errors.Wrap(err, "failed to display lint results")
	}

	return hasFailure, nil
}

func (r *runners) lintPreflightSpecs(cmd *cobra.Command, config *tools.Config) (bool, error) {
	// Get preflight version from config
	preflightVersion := tools.DefaultPreflightVersion
	if config.ReplLint.Tools != nil {
		if v, ok := config.ReplLint.Tools[tools.ToolPreflight]; ok {
			preflightVersion = v
		}
	}

	// Check if there are any preflight specs configured
	preflightPaths, err := lint2.GetPreflightPathsFromConfig(config)
	if err != nil {
		return false, errors.Wrap(err, "failed to expand preflight paths")
	}

	// Lint all preflight specs and collect results
	var allResults []*lint2.LintResult
	var allPaths []string
	hasFailure := false

	for _, specPath := range preflightPaths {
		result, err := lint2.LintPreflight(cmd.Context(), specPath, preflightVersion)
		if err != nil {
			return false, errors.Wrapf(err, "failed to lint preflight spec: %s", specPath)
		}

		allResults = append(allResults, result)
		allPaths = append(allPaths, specPath)

		if !result.Success {
			hasFailure = true
		}
	}

	// Display results for all preflight specs
	if err := displayAllLintResults(r.w, "preflight spec", allPaths, allResults); err != nil {
		return false, errors.Wrap(err, "failed to display lint results")
	}

	return hasFailure, nil
}

type resourceSummary struct {
	errorCount   int
	warningCount int
	infoCount    int
}

func displayAllLintResults(w io.Writer, resourceType string, resourcePaths []string, results []*lint2.LintResult) error {
	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	totalResourcesFailed := 0

	// Display results for each resource
	for i, result := range results {
		resourcePath := resourcePaths[i]
		summary := displaySingleResourceResult(w, resourceType, resourcePath, result)

		totalErrors += summary.errorCount
		totalWarnings += summary.warningCount
		totalInfo += summary.infoCount

		if !result.Success {
			totalResourcesFailed++
		}
	}

	// Print overall summary if multiple resources
	if len(results) > 1 {
		displayOverallSummary(w, resourceType, len(results), totalResourcesFailed, totalErrors, totalWarnings, totalInfo)
	}

	return nil
}

func displaySingleResourceResult(w io.Writer, resourceType string, resourcePath string, result *lint2.LintResult) resourceSummary {
	// Print header for this resource
	fmt.Fprintf(w, "==> Linting %s: %s\n\n", resourceType, resourcePath)

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

	// Print per-resource summary
	fmt.Fprintf(w, "\nSummary for %s: %d error(s), %d warning(s), %d info\n",
		resourcePath, summary.errorCount, summary.warningCount, summary.infoCount)

	// Print per-resource status
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

func countMessagesBySeverity(messages []lint2.LintMessage) resourceSummary {
	summary := resourceSummary{}
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

func displayOverallSummary(w io.Writer, resourceType string, totalResources, failedResources, totalErrors, totalWarnings, totalInfo int) {
	// Pluralize resource type (simple 's' suffix)
	// Note: Works for current resource types (chart→charts, preflight spec→preflight specs)
	// Would need enhancement for irregular plurals if new resource types are added
	resourceTypePlural := resourceType + "s"

	fmt.Fprintf(w, "==> Overall Summary\n")
	fmt.Fprintf(w, "%s linted: %d\n", resourceTypePlural, totalResources)
	fmt.Fprintf(w, "%s passed: %d\n", resourceTypePlural, totalResources-failedResources)
	fmt.Fprintf(w, "%s failed: %d\n", resourceTypePlural, failedResources)
	fmt.Fprintf(w, "Total errors: %d\n", totalErrors)
	fmt.Fprintf(w, "Total warnings: %d\n", totalWarnings)
	fmt.Fprintf(w, "Total info: %d\n", totalInfo)

	if failedResources > 0 {
		fmt.Fprintf(w, "\nOverall Status: Failed\n")
	} else {
		fmt.Fprintf(w, "\nOverall Status: Passed\n")
	}
}

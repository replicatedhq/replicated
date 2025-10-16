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
		Short: "Lint a Helm chart",
		Long:  `Lint a Helm chart using helm lint. This command executes helm lint locally and displays the results.`,
	}

	cmd.Flags().StringVar(&r.args.lintChart, "chart", "", "Path to Helm chart (directory or .tgz)")
	cmd.MarkFlagRequired("chart")

	cmd.RunE = r.runLint

	parent.AddCommand(cmd)
	return cmd
}

func (r *runners) runLint(cmd *cobra.Command, args []string) error {
	chartPath := r.args.lintChart

	// Execute lint
	result, err := lint2.LintChart(cmd.Context(), chartPath)
	if err != nil {
		return errors.Wrap(err, "failed to lint chart")
	}

	// Display results
	if err := displayLintResults(r.w, chartPath, result); err != nil {
		return errors.Wrap(err, "failed to display lint results")
	}

	// Flush the tab writer
	if err := r.w.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush output")
	}

	// Return error if linting failed (for correct exit code)
	if !result.Success {
		return errors.New("linting failed")
	}

	return nil
}

func displayLintResults(w io.Writer, chartPath string, result *lint2.LintResult) error {
	// Print header
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

	// Print summary
	fmt.Fprintf(w, "\n")
	errorCount := 0
	warningCount := 0
	infoCount := 0
	for _, msg := range result.Messages {
		switch msg.Severity {
		case "ERROR":
			errorCount++
		case "WARNING":
			warningCount++
		case "INFO":
			infoCount++
		}
	}

	fmt.Fprintf(w, "Summary: %d error(s), %d warning(s), %d info\n", errorCount, warningCount, infoCount)

	// Print overall status
	if result.Success {
		fmt.Fprintf(w, "\nLinting passed\n")
	} else {
		fmt.Fprintf(w, "\nLinting failed\n")
	}

	return nil
}

package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/pkg/errors"
)

// LintOutput represents the complete lint output structure
// This is imported from cli/cmd but redefined here to avoid circular imports
type LintOutput interface{}

// LintResults formats and prints lint results in the specified format
func LintResults(format string, w *tabwriter.Writer, output interface{}) error {
	switch format {
	case "table":
		// Table format is handled by the display functions in lint.go
		// This function is only called for non-table formats
		return errors.New("table format should be handled by display functions")
	case "json":
		return printLintResultsJSON(w, output)
	default:
		return errors.Errorf("invalid output format: %s. Supported formats: json, table", format)
	}
}

// printLintResultsJSON outputs lint results as formatted JSON
func printLintResultsJSON(w *tabwriter.Writer, output interface{}) error {
	// Marshal to JSON with pretty printing
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal lint results to JSON")
	}

	// Write JSON to output
	if _, err := fmt.Fprintln(w, string(jsonBytes)); err != nil {
		return errors.Wrap(err, "failed to write JSON output")
	}

	// Flush the writer
	if err := w.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush output")
	}

	return nil
}

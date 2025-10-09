package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/types"
)

// NetworkReport prints network report in JSON format
func NetworkReport(w *tabwriter.Writer, report *types.NetworkReport) error {
	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to json: %v", err)
	}
	if _, err := fmt.Fprintln(w, string(reportBytes)); err != nil {
		return err
	}
	return w.Flush()
}

// NetworkEvents prints network events in JSON format (for watch mode)
func NetworkEvents(w *tabwriter.Writer, events []*types.NetworkEventData) error {
	for _, event := range events {
		eventBytes, err := json.Marshal(event)
		if err != nil {
			continue // Skip events that can't be marshaled
		}
		if _, err := fmt.Fprintln(w, string(eventBytes)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func NetworkReportSummary(w *tabwriter.Writer, summary *types.NetworkReportSummary) error {
	summaryBytes, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report summary to json: %v", err)
	}
	if _, err := fmt.Fprintln(w, string(summaryBytes)); err != nil {
		return err
	}
	return w.Flush()
}

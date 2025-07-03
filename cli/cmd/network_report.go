package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkReport(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Get network report",
		Long:  "Get a network report showing traffic analysis for a specified network",
		Example: `# Get report for a network by ID
replicated network report --id abc123

# Watch for new network events
replicated network report --id abc123 --watch`,
		RunE:              r.getNetworkReport,
		ValidArgsFunction: r.completeNetworkIDs,
		Hidden:            true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.networkReportID, "id", "", "Network ID to get report for (required)")
	cmd.MarkFlagRequired("id")
	cmd.RegisterFlagCompletionFunc("id", r.completeNetworkIDs)

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "json", "The output format to use. One of: json|table")
	cmd.Flags().BoolVarP(&r.args.networkReportWatch, "watch", "w", false, "watch for new network events")

	return cmd
}

func (r *runners) getNetworkReport(_ *cobra.Command, args []string) error {
	if r.args.networkReportID == "" {
		return errors.New("network ID is required")
	}

	// Initialize the client
	err := r.initNetworkClient()
	if err != nil {
		return errors.Wrap(err, "initialize client")
	}

	// Don't call getNetworkIDFromArg here. Reporting API supports short IDs and will also work for networks that have been deleted.

	// Get the initial network report
	report, err := r.kotsAPI.GetNetworkReport(r.args.networkReportID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "get network report")
	}

	// Handle watch mode
	if r.args.networkReportWatch {
		// Check output format for watch
		if r.outputFormat != "table" {
			return errors.New("watch is only supported for table output")
		}

		// Print initial events
		if len(report.Events) > 0 {
			printEventsTable(report.Events, true)
		}

		// Track the last seen event time
		var lastEventTime *time.Time
		if len(report.Events) > 0 {
			lastEventTime = &report.Events[len(report.Events)-1].CreatedAt
		}

		// Poll for new events
		for range time.Tick(2 * time.Second) {
			var newReport *types.NetworkReport
			if lastEventTime != nil {
				newReport, err = r.kotsAPI.GetNetworkReportAfter(r.args.networkReportID, lastEventTime)
			} else {
				newReport, err = r.kotsAPI.GetNetworkReport(r.args.networkReportID)
			}

			if err != nil {
				return errors.Wrap(err, "get network report")
			}

			// Print new events
			if len(newReport.Events) > 0 {
				printEventsTable(newReport.Events, false)
				// Update last seen time
				lastEventTime = &newReport.Events[len(newReport.Events)-1].CreatedAt
			}
		}
		return nil
	}

	// Output the report (non-watch mode)
	return outputReport(report, r.outputFormat)
}

func outputReport(report *types.NetworkReport, outputFormat string) error {
	switch outputFormat {
	case "json":
		output, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return errors.Wrap(err, "marshal report to json")
		}
		fmt.Println(string(output))
	case "table":
		if len(report.Events) == 0 {
			fmt.Println("No network events found.")
			return nil
		}
		printEventsTable(report.Events, true)
	default:
		return errors.Errorf("unsupported output format: %s", outputFormat)
	}
	return nil
}

func printEventsTable(events []*types.NetworkEvent, includeHeader bool) {
	if includeHeader {
		fmt.Printf("%-20s %-15s %-15s %-8s %-8s %-8s %-12s %-8s %s\n",
			"CREATED AT", "SRC IP", "DST IP", "SRC PORT", "DST PORT", "PROTOCOL", "COMMAND", "PID", "SERVICE")
		fmt.Println("---")
	}

	for _, event := range events {
		// Parse the event data if it's JSON
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(event.EventData), &eventData); err == nil {
			fmt.Printf("%-20s %-15s %-15s %-8.0f %-8.0f %-8s %-12s %-8.0f %s\n",
				event.CreatedAt.Format("2006-01-02 15:04:05"),
				getStringValue(eventData, "srcIp"),
				getStringValue(eventData, "dstIp"),
				getFloatValue(eventData, "srcPort"),
				getFloatValue(eventData, "dstPort"),
				getStringValue(eventData, "proto"),
				getStringValue(eventData, "comm"),
				getFloatValue(eventData, "pid"),
				getStringValue(eventData, "likelyService"))
		} else {
			// Fallback if event data is not valid JSON
			fmt.Printf("%-20s %s\n",
				event.CreatedAt.Format("2006-01-02 15:04:05"),
				event.EventData)
		}
	}
}

func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatValue(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

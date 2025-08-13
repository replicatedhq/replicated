package cmd

import (
	"os"
	"text/tabwriter"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
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

# Watch for new network events (table format)
replicated network report --id abc123 --watch --output table

# Watch for new network events (JSON Lines format)
replicated network report --id abc123 --watch --output json`,
		RunE:              r.getNetworkReport,
		ValidArgsFunction: r.completeNetworkIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.networkReportID, "id", "", "Network ID to get report for (required)")
	cmd.MarkFlagRequired("id")
	cmd.RegisterFlagCompletionFunc("id", r.completeNetworkIDs)

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "json", "The output format to use. One of: json|table")
	cmd.Flags().BoolVarP(&r.args.networkReportWatch, "watch", "w", false, "Watch for new network events")

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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Print initial events
		if len(report.Events) > 0 {
			if err := print.NetworkEvents(r.outputFormat, w, report.Events, true); err != nil {
				return errors.Wrap(err, "print initial network events")
			}
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
				if err := print.NetworkEvents(r.outputFormat, w, newReport.Events, false); err != nil {
					return errors.Wrap(err, "print new network events")
				}
				// Update last seen time
				lastEventTime = &newReport.Events[len(newReport.Events)-1].CreatedAt
			}
		}
		return nil
	}

	// Output the report (non-watch mode)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	return print.NetworkReport(r.outputFormat, w, report)
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/util"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkReport(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report [network-id]",
		Short: "Get network report",
		Long: `Get a network report showing detailed network activity for a specified network.

The report shows individual network events including source/destination IPs, ports, protocols, 
pods, processes, and DNS queries. Reports must be enabled with 'replicated network update <network-id> --collect-report'.

Output formats:
  - Default: Full event details in JSON format
  - --summary: Aggregated statistics with top domains and destinations
  - --watch: Continuous stream of new events in JSON Lines format`,
		Example: `# Get full network traffic report
replicated network report <network-id>

# Get aggregated summary with statistics. Only available for networks that have been terminated.
replicated network report <network-id> --summary

# Watch for new network events in real-time
replicated network report <network-id> --watch`,
		RunE:              r.getNetworkReport,
		ValidArgsFunction: r.completeNetworkIDs,
		Hidden:            true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.networkReportID, "id", "", "Network ID to get report for")
	cmd.RegisterFlagCompletionFunc("id", r.completeNetworkIDs)

	cmd.Flags().BoolVarP(&r.args.networkReportWatch, "watch", "w", false, "Watch for new network events in real-time (polls every 2 seconds)")
	cmd.Flags().BoolVar(&r.args.networkReportSummary, "summary", false, "Get aggregated report summary with statistics instead of individual events")
	cmd.Flags().BoolVar(&r.args.networkReportIgnoreDefault, "ignore-default", true, "Ignore special-use domains network traffic in the report")

	return cmd
}

func (r *runners) getNetworkReport(cmd *cobra.Command, args []string) error {
	// Use positional argument if --id flag wasn't provided
	if r.args.networkReportID == "" {
		if len(args) == 0 {
			return errors.New("network ID is required (provide as first argument or use --id flag)")
		}
		r.args.networkReportID = args[0]
	}

	// Initialize the client
	err := r.initNetworkClient()
	if err != nil {
		return errors.Wrap(err, "initialize client")
	}

	// Don't call getNetworkIDFromArg here. Reporting API supports short IDs and will also work for networks that have been deleted.

	// Validate flags
	if r.args.networkReportSummary && cmd.Flags().Lookup("ignore-default").Changed {
		return fmt.Errorf("cannot use --ignore-default and --summary flags together")
	}

	// Get the initial network report or summary depending on args provided
	if r.args.networkReportSummary {
		return r.getNetworkReportSummary(cmd.Context())
	} else {
		return r.getNetworkReportEvents()
	}
}

func (r *runners) getNetworkReportEvents() error {
	report, err := r.kotsAPI.GetNetworkReport(r.args.networkReportID, r.args.networkReportIgnoreDefault)
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
			if err := print.NetworkEvents(w, report.Events); err != nil {
				return errors.Wrap(err, "print initial network events")
			}
		} else {
			// Print empty report
			print.NetworkReport(w, report)
		}

		// Track the last seen event time
		var lastEventTime *time.Time
		if len(report.Events) > 0 {
			// Extract timestamp from the last event
			if parsedTime, err := util.ParseTime(report.Events[len(report.Events)-1].Timestamp); err == nil {
				lastEventTime = &parsedTime
			}
		}

		// Poll for new events
		for range time.Tick(2 * time.Second) {
			var newReport *types.NetworkReport
			if lastEventTime != nil {
				newReport, err = r.kotsAPI.GetNetworkReportAfter(r.args.networkReportID, lastEventTime, r.args.networkReportIgnoreDefault)
			} else {
				newReport, err = r.kotsAPI.GetNetworkReport(r.args.networkReportID, r.args.networkReportIgnoreDefault)
			}

			if err != nil {
				return errors.Wrap(err, "get network report")
			}

			// Print new events
			if len(newReport.Events) > 0 {
				if err := print.NetworkEvents(w, newReport.Events); err != nil {
					return errors.Wrap(err, "print new network events")
				}
				// Update last seen time
				if parsedTime, err := util.ParseTime(newReport.Events[len(newReport.Events)-1].Timestamp); err == nil {
					lastEventTime = &parsedTime
				}
			}
		}
		return nil
	}

	// Output the report (non-watch mode)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	return print.NetworkReport(w, report)
}

func (r *runners) getNetworkReportSummary(ctx context.Context) error {
	if r.args.networkReportWatch {
		return fmt.Errorf("cannot use watch and summary flags together")
	}

	summary, err := r.kotsAPI.GetNetworkReportSummary(ctx, r.args.networkReportID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if errors.Cause(err) == platformclient.ErrNotFound {
		return fmt.Errorf("network report summary not found for network %s, network must be terminated and events must have been processed", r.args.networkReportID)
	} else if err != nil {
		return errors.Wrap(err, "get network report summary")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	return print.NetworkReportSummary(w, summary)
}

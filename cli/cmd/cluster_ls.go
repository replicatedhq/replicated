package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls",
		Short:        "list test clusters",
		Long:         `list test clusters`,
		RunE:         r.listClusters,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.lsClusterHideTerminated, "hide-terminated", false, "when set, do not show terminated clusters")
	cmd.Flags().StringVar(&r.args.lsClusterStartTime, "start-time", "", "limit output to clusters that exist during or after this date time (optional)")
	cmd.Flags().StringVar(&r.args.lsClusterEndTime, "end-time", "", "limit output to clusters that exist during or before this date time (optional)")
	cmd.Flags().StringSliceVar(&r.args.lsClusterIDs, "cluster-ids", nil, "cluster ids to request data for (optional, overrides start-time and end-time)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listClusters(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	opts := kotsclient.ListClustersOpts{
		IncludeTerminated: !r.args.lsClusterHideTerminated,
	}

	if len(r.args.lsClusterIDs) > 0 {
		opts.ClusterIDs = r.args.lsClusterIDs
	} else {
		if r.args.lsClusterStartTime != "" {
			startTime, err := time.Parse(time.RFC3339, r.args.lsClusterStartTime)
			if err != nil {
				return errors.Wrap(err, "invalid startTime")
			}
			opts.StartTime = startTime
		}

		if r.args.lsClusterEndTime != "" {
			endTime, err := time.Parse(time.RFC3339, r.args.lsClusterEndTime)
			if err != nil {
				return errors.Wrap(err, "invalid endTime format")
			}
			opts.EndTime = endTime
		}
	}

	clusters, err := kotsRestClient.ListClusters(opts)
	if err == platformclient.ErrForbidden {
		return errors.New("This command is not available for your account or team. Please contact your customer success representative for more information.")
	}
	if err != nil {
		return errors.Wrap(err, "list clusters")
	}

	if len(clusters) == 0 {
		return print.NoClusters(r.outputFormat, r.w)
	}

	return print.Clusters(r.outputFormat, r.w, clusters)
}

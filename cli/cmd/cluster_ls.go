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

	cmd.Flags().BoolVar(&r.args.lsClusterShowTerminated, "show-terminated", false, "when set, only show terminated clusters")
	cmd.Flags().StringVar(&r.args.lsClusterStartTime, "start-time", "", "start time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.args.lsClusterEndTime, "end-time", "", "end time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listClusters(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	const longForm = "2006-01-02T15:04:05Z"
	var startTime, endTime *time.Time
	if r.args.lsClusterStartTime != "" {
		st, err := time.Parse(longForm, r.args.lsClusterStartTime)
		if err != nil {
			return errors.Wrap(err, "parse start time")
		}
		startTime = &st
	}
	if r.args.lsClusterEndTime != "" {
		et, err := time.Parse(longForm, r.args.lsClusterEndTime)
		if err != nil {
			return errors.Wrap(err, "parse end time")
		}
		endTime = &et
	}

	clusters, err := kotsRestClient.ListClusters(r.args.lsClusterShowTerminated, startTime, endTime)
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

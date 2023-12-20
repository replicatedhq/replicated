package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterUpdateTTL(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ttl [ID]",
		Short:        "Update TTL for a test cluster",
		Long:         `Update TTL for a test cluster`,
		Args:         cobra.ExactArgs(1),
		RunE:         r.updateClusterTTL,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateClusterTTL, "ttl", "", "Update TTL which starts from the moment the cluster is running (duration, max 48h).")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	_ = cmd.MarkFlagRequired("ttl")

	return cmd
}

func (r *runners) updateClusterTTL(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("cluster id is required")
	}
	clusterID := args[0]

	opts := kotsclient.UpdateClusterTTLOpts{
		TTL: r.args.updateClusterTTL,
	}
	cl, err := r.kotsAPI.UpdateClusterTTL(clusterID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "update cluster ttl")
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}

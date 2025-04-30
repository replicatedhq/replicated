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
		Use:   "ttl [ID]",
		Short: "Update TTL for a test cluster.",
		Long:  `The 'ttl' command allows you to update the Time-To-Live (TTL) of a test cluster. The TTL represents the duration for which the cluster will remain active before it is automatically terminated. The duration starts from the moment the cluster becomes active. You must provide a valid duration, with a maximum limit of 48 hours. If no TTL is specified, the default TTL is 1 hour.`,
		Example: `# Update the TTL for a specific cluster
replicated cluster update ttl CLUSTER_ID --ttl 24h`,
		RunE:              r.updateClusterTTL,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateClusterTTL, "ttl", "", "Update TTL which starts from the moment the cluster is running (duration, max 48h).")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	cmd.MarkFlagRequired("ttl")

	return cmd
}

func (r *runners) updateClusterTTL(cmd *cobra.Command, args []string) error {
	if err := r.ensureUpdateClusterIDArg(args); err != nil {
		return errors.Wrap(err, "ensure cluster id arg")
	}

	opts := kotsclient.UpdateClusterTTLOpts{
		TTL: r.args.updateClusterTTL,
	}
	cl, err := r.kotsAPI.UpdateClusterTTL(r.args.updateClusterID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "update cluster ttl")
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}

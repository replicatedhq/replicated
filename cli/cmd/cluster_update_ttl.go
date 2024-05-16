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
		RunE:         r.updateClusterTTL,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateClusterTTL, "ttl", "", "Update TTL which starts from the moment the cluster is running (duration, max 48h).")
	cmd.Flags().StringVar(&r.args.updateClusterName, "name", "", "Name of the cluster to update TTL.")
	cmd.Flags().StringVar(&r.args.updateClusterID, "id", "", "id of the cluster to update TTL (when name is not provided)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	_ = cmd.MarkFlagRequired("ttl")

	return cmd
}

func (r *runners) updateClusterTTL(cmd *cobra.Command, args []string) error {
	// by default, we look at args[0] as the id
	// but if it's not provided, we look for a viper flag named "name" and use it
	// as the name of the cluster, not the id
	clusterID := ""
	if len(args) > 0 {
		clusterID = args[0]
	} else if r.args.updateClusterName != "" {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		} else if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			if cluster.Name == r.args.updateClusterName {
				clusterID = cluster.ID
				break
			}
		}
	} else if r.args.updateClusterID != "" {
		clusterID = r.args.updateClusterID
	} else {
		return errors.New("must provide cluster id or name")
	}

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

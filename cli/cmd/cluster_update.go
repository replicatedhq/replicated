package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update cluster settings.",
		Long: `The 'update' command allows you to update various settings of a test cluster, such as its name or ID.

You can either specify the cluster ID directly or provide the cluster name, and the command will resolve the corresponding cluster ID. This allows you to modify the cluster's configuration based on the unique identifier or the name of the cluster.`,
		Example: `# Update a cluster using its ID
replicated cluster update --id <cluster-id> [subcommand]

# Update a cluster using its name
replicated cluster update --name <cluster-name> [subcommand]`,
	}
	parent.AddCommand(cmd)

	cmd.PersistentFlags().StringVar(&r.args.updateClusterName, "name", "", "Name of the cluster to update.")
	cmd.RegisterFlagCompletionFunc("name", r.completeClusterNames)

	cmd.PersistentFlags().StringVar(&r.args.updateClusterID, "id", "", "id of the cluster to update (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeClusterIDs)

	return cmd
}

func (r *runners) ensureUpdateClusterIDArg(args []string) error {
	// by default, we look at args[0] as the id or name
	// but if it's not provided, we look for a flag named "name" or "id"
	if len(args) > 0 {
		clusterID, err := r.getClusterIDFromArg(args[0])
		if err != nil {
			return errors.Wrap(err, "get cluster id from arg")
		}
		r.args.updateClusterID = clusterID
	} else if r.args.updateClusterName != "" {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		} else if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			if cluster.Name == r.args.updateClusterName {
				r.args.updateClusterID = cluster.ID
				break
			}
		}
	} else if r.args.updateClusterID != "" {
		// do nothing
		// but this is here for readability
	} else {
		return errors.New("must provide cluster id or name")
	}

	return nil
}

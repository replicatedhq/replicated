package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update cluster settings",
		Long:  `cluster update can be used to update cluster settings`,
	}
	parent.AddCommand(cmd)

	cmd.PersistentFlags().StringVar(&r.args.updateClusterName, "name", "", "Name of the cluster to update TTL.")
	cmd.RegisterFlagCompletionFunc("name", r.completeClusterNames)

	cmd.PersistentFlags().StringVar(&r.args.updateClusterID, "id", "", "id of the cluster to update TTL (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeClusterIDs)

	return cmd
}

func (r *runners) ensureUpdateClusterIDArg(args []string) error {
	// by default, we look at args[0] as the id
	// but if it's not provided, we look for a viper flag named "name" and use it
	// as the name of the cluster, not the id
	if len(args) > 0 {
		r.args.updateClusterID = args[0]
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

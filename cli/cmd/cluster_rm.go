package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rm",
		Short:        "remove test clusters",
		Long:         `remove test clusters`,
		RunE:         r.removeCluster,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.removeClusterID, "id", "", "cluster id")
	cmd.Flags().BoolVar(&r.args.removeClusterForce, "force", false, "force remove cluster")

	return cmd
}

func (r *runners) removeCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	err := kotsRestClient.RemoveCluster(r.args.removeClusterID, r.args.removeClusterForce)
	if err != nil {
		return errors.Wrap(err, "remove cluster")
	}

	return nil
}

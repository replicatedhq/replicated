package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rm ID [ID …]",
		Short:        "remove test clusters",
		Long:         `remove test clusters`,
		RunE:         r.removeCluster,
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.removeClusterForce, "force", false, "force remove cluster")

	return cmd
}

func (r *runners) removeCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	for _, arg := range args {
		err := kotsRestClient.RemoveCluster(arg, r.args.removeClusterForce)
		if err != nil {
			return errors.Wrap(err, "remove cluster")
		}
	}

	return nil
}

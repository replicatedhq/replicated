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

	return cmd
}

func (r *runners) removeCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	err := kotsRestClient.RemoveCluster("e6fe42e9-bebf-4949-4abc-d99bf34a3d33")
	if err != nil {
		return errors.Wrap(err, "remove cluster")
	}

	return nil
}

package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create test clusters",
		Long:         `create test clusters`,
		RunE:         r.createCluster,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) createCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	_, err := kotsRestClient.CreateCluster("")
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

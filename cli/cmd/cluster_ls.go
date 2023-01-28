package cmd

import (
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

	return cmd
}

func (r *runners) listClusters(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	clusters, err := kotsRestClient.ListClusters()
	if err == platformclient.ErrForbidden {
		return errors.New("This command is not available for your account or team. Please contact your customer success representative for more information.")
	}
	if err != nil {
		return errors.Wrap(err, "list clusters")
	}

	if len(clusters) == 0 {
		return print.NoClusters(r.w)
	}

	return print.Clusters(r.w, clusters)
}
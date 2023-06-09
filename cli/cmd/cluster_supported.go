package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterSupported(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "supported",
		Short:        "list supported clusters",
		Long:         `list supported clusters`,
		RunE:         r.listSupportedClusters,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lsSupportedClusterKubernetesDistribution, "distribution", "", "Kubernetes distribution to filter by.")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listSupportedClusters(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	sc, err := kotsRestClient.ListSupportedClusters()
	if err == platformclient.ErrForbidden {
		return errors.New("This command is not available for your account or team. Please contact your customer success representative for more information.")
	}
	if err != nil {
		return errors.Wrap(err, "list supported clusters")
	}

	if r.args.lsSupportedClusterKubernetesDistribution != "" {
		var filteredSC []*types.SupportedCluster
		for _, cluster := range sc {
			if cluster.Name == r.args.lsSupportedClusterKubernetesDistribution {
				filteredSC = append(filteredSC, cluster)
				break
			}
		}
		sc = filteredSC
	}

	if len(sc) == 0 {
		return print.NoSupportedClusters(r.outputFormat, r.w)
	}

	return print.SupportedClusters(r.outputFormat, r.w, sc)
}

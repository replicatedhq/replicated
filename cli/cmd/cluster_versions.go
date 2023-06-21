package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterVersions(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "versions",
		Short:        "list cluster versions",
		Long:         `list cluster versions`,
		RunE:         r.listClusterVersions,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lsVersionsClusterKubernetesDistribution, "distribution", "", "Kubernetes distribution to filter by.")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listClusterVersions(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	cv, err := kotsRestClient.ListClusterVersions()
	if err == platformclient.ErrForbidden {
		return errors.New("This command is not available for your account or team. Please contact your customer success representative for more information.")
	}
	if err != nil {
		return errors.Wrap(err, "list cluster versions")
	}

	if r.args.lsVersionsClusterKubernetesDistribution != "" {
		var filteredCV []*types.ClusterVersion
		for _, cluster := range cv {
			if cluster.Name == r.args.lsVersionsClusterKubernetesDistribution {
				filteredCV = append(filteredCV, cluster)
				break
			}
		}
		cv = filteredCV
	}

	if len(cv) == 0 {
		return print.NoClusterVersions(r.outputFormat, r.w)
	}

	return print.ClusterVersions(r.outputFormat, r.w, cv)
}

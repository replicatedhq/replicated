package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterVersions(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List cluster versions.",
		Long:  `The 'versions' command lists available Kubernetes versions for supported distributions. You can filter the versions by specifying a distribution and choose between different output formats.`,
		Example: `  # List all available Kubernetes cluster versions
  replicated cluster versions

  # List available versions for a specific distribution (e.g., eks)
  replicated cluster versions --distribution eks

  # Output the versions in JSON format
  replicated cluster versions --output json`,
		RunE: r.listClusterVersions,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lsVersionsDistribution, "distribution", "", "Kubernetes distribution to filter by.")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listClusterVersions(_ *cobra.Command, args []string) error {
	cv, err := r.kotsAPI.ListClusterVersions()
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list cluster versions")
	}

	if r.args.lsVersionsDistribution != "" {
		var filteredCV []*types.ClusterVersion
		for _, cluster := range cv {
			if cluster.Name == r.args.lsVersionsDistribution {
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

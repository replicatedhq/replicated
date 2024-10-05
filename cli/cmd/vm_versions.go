package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMVersions(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "List available VM versions.",
		Long: `The 'vm versions' command lists all the available versions of virtual machines that can be provisioned. This includes the available distributions and their respective versions.

- You can filter the list by a specific distribution using the '--distribution' flag.
- The output can be formatted as a table or in JSON format using the '--output' flag.`,
		Example: `  # List all available VM versions
  replicated vm versions

  # List VM versions for a specific distribution (e.g., Ubuntu)
  replicated vm versions --distribution ubuntu

  # Display the output in JSON format
  replicated vm versions --output json`,
		RunE: r.listVMVersions,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lsVersionsDistribution, "distribution", "", "Kubernetes distribution to filter by.")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listVMVersions(_ *cobra.Command, args []string) error {
	vmVersions, err := r.kotsAPI.ListVMVersions()
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list vm versions")
	}

	if r.args.lsVersionsDistribution != "" {
		var filteredCV []*types.VMVersion
		for _, vmVersion := range vmVersions {
			if vmVersion.Name == r.args.lsVersionsDistribution {
				filteredCV = append(filteredCV, vmVersion)
				break
			}
		}
		vmVersions = filteredCV
	}

	if len(vmVersions) == 0 {
		return print.NoVMVersions(r.outputFormat, r.w)
	}

	return print.VMVersions(r.outputFormat, r.w, vmVersions)
}

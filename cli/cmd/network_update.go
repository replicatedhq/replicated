package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [ID_OR_NAME]",
		Short: "Update network settings",
		Long: `The 'update' command allows you to update various settings of a test network.

You can either specify the network ID directly or provide the network name, and the command will resolve the corresponding network ID.

Network Policies are currently a beta feature.`,
		Example: `# Update a network using its ID
replicated network update <network-id> --policy airgap

# Update a network using its name
replicated network update <network-name> --policy airgap

# Update using --id or --name flags
replicated network update --id <network-id> --policy airgap
replicated network update --name <network-name> --policy airgap
`,
		RunE:              r.updateNetwork,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeNetworkIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.PersistentFlags().StringVar(&r.args.updateNetworkName, "name", "", "Name of the network to update")
	cmd.RegisterFlagCompletionFunc("name", r.completeNetworkNames)

	cmd.PersistentFlags().StringVar(&r.args.updateNetworkID, "id", "", "id of the network to update (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeNetworkIDs)

	cmd.Flags().StringVarP(&r.args.updateNetworkPolicy, "policy", "p", "", "Update network policy setting")
	cmd.Flags().BoolVarP(&r.args.updateNetworkCollectReport, "collect-report", "r", false, "Enable report collection on this network (use --collect-report=false to disable)")
	// TODO: Remove this once report collection is Beta, ensure we add
	// examples in the above help text as well.
	cmd.Flags().MarkHidden("collect-report")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) updateNetwork(cmd *cobra.Command, args []string) error {
	if err := r.ensureUpdateNetworkIDArg(args); err != nil {
		return errors.Wrap(err, "must provide network id or name")
	}

	// Check if any update flags are provided
	if r.args.updateNetworkPolicy == "" && !cmd.Flags().Changed("collect-report") {
		// If no specific update flags are provided, show help
		return cmd.Help()
	}

	// Prepare update options
	opts := kotsclient.UpdateNetworkOpts{}

	if r.args.updateNetworkPolicy != "" {
		opts.Policy = r.args.updateNetworkPolicy
	}

	if cmd.Flags().Changed("collect-report") {
		opts.CollectReport = &r.args.updateNetworkCollectReport
	}

	// Update the network
	network, err := r.kotsAPI.UpdateNetwork(r.args.updateNetworkID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	}
	if err != nil {
		return errors.Wrap(err, "update network")
	}

	return print.Network(r.outputFormat, r.w, network)
}

func (r *runners) ensureUpdateNetworkIDArg(args []string) error {
	if len(args) > 0 {
		networkID, err := r.getNetworkIDFromArg(args[0])
		if err != nil {
			return err
		}
		r.args.updateNetworkID = networkID
	} else if r.args.updateNetworkName != "" {
		networks, err := r.kotsAPI.ListNetworks(false, nil, nil)
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		} else if err != nil {
			return errors.Wrap(err, "list networks")
		}
		for _, network := range networks {
			if network.Name == r.args.updateNetworkName {
				r.args.updateNetworkID = network.ID
				break
			}
		}
	} else if r.args.updateNetworkID != "" {
		// do nothing
	} else {
		return errors.New("must provide network id or name")
	}

	return nil
}

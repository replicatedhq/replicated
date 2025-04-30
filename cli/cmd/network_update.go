package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update network settings.",
		Long: `The 'update' command allows you to update various settings of a test network.

You can either specify the network ID directly or provide the network name, and the command will resolve the corresponding network ID.`,
		Example: `# Update a network using its ID
replicated network update --id <network-id> [subcommand]

# Update a network using its name
replicated network update --name <network-name> [subcommand]`,
	}
	parent.AddCommand(cmd)

	cmd.PersistentFlags().StringVar(&r.args.updateNetworkName, "name", "", "Name of the network to update.")
	cmd.RegisterFlagCompletionFunc("name", r.completeNetworkNames)

	cmd.PersistentFlags().StringVar(&r.args.updateNetworkID, "id", "", "id of the network to update (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeNetworkIDs)

	return cmd
}

func (r *runners) ensureUpdateNetworkIDArg(args []string) error {
	if len(args) > 0 {
		networkID, err := r.getNetworkIDFromArg(args[0])
		if err != nil {
			return err
		}
		r.args.updateNetworkID = networkID
	} else if r.args.updateNetworkName != "" {
		networks, err := r.kotsAPI.ListNetworks(nil, nil)
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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm ID [ID …]",
		Aliases: []string{"delete"},
		Short:   "Remove test network(s) immediately, with options to filter by name or remove all networks.",
		Long: `The 'rm' command allows you to remove test networks from your account immediately. You can specify one or more network IDs directly, or use flags to filter which networks to remove based on their name or simply remove all networks at once.

This command supports multiple filtering options, including removing networks by their name or by specifying the '--all' flag to remove all networks in your account.

You can also use the '--dry-run' flag to simulate the removal without actually deleting the networks.`,
		Example: `# Remove a network by ID
replicated network rm aaaaa11

# Remove multiple networks by ID
replicated network rm aaaaa11 bbbbb22 ccccc33

# Remove all networks with a specific name
replicated network rm --name test-network

# Remove all networks
replicated network rm --all

# Perform a dry run of removing all networks
replicated network rm --all --dry-run`,
		RunE:              r.removeNetworks,
		ValidArgsFunction: r.completeNetworkIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringArrayVar(&r.args.removeNetworkNames, "name", []string{}, "Name of the network to remove (can be specified multiple times)")
	cmd.RegisterFlagCompletionFunc("name", r.completeNetworkNames)

	cmd.Flags().BoolVar(&r.args.removeNetworkAll, "all", false, "remove all networks")
	cmd.Flags().BoolVar(&r.args.removeNetworkDryRun, "dry-run", false, "Dry run")

	return cmd
}

func (r *runners) removeNetworks(_ *cobra.Command, args []string) error {
	if len(args) == 0 && !r.args.removeNetworkAll && len(r.args.removeNetworkNames) == 0 {
		return errors.New("One of ID, --all, or --name flag required")
	} else if len(args) > 0 && (r.args.removeNetworkAll || len(r.args.removeNetworkNames) > 0) {
		return errors.New("cannot specify ID and --all or --name flag")
	} else if len(args) == 0 && r.args.removeNetworkAll && len(r.args.removeNetworkNames) > 0 {
		return errors.New("cannot specify --all and --name flag")
	}

	if len(r.args.removeNetworkNames) > 0 {
		networks, err := r.kotsAPI.ListNetworks(nil, nil)
		if err != nil {
			return errors.Wrap(err, "list networks")
		}
		for _, network := range networks {
			for _, name := range r.args.removeNetworkNames {
				if network.Name == name {
					err := removeNetwork(r, network.ID)
					if err != nil {
						return errors.Wrap(err, "remove network")
					}
				}
			}
		}
	}

	if r.args.removeNetworkAll {
		networks, err := r.kotsAPI.ListNetworks(nil, nil)
		if err != nil {
			return errors.Wrap(err, "list networks")
		}
		for _, network := range networks {
			err := removeNetwork(r, network.ID)
			if err != nil {
				return errors.Wrap(err, "remove network")
			}
		}
	}

	for _, arg := range args {
		err := removeNetwork(r, arg)
		if err != nil {
			return errors.Wrap(err, "remove network")
		}
	}

	return nil
}

func removeNetwork(r *runners, networkID string) error {
	if r.args.removeNetworkDryRun {
		fmt.Printf("would remove network %s\n", networkID)
		return nil
	}
	err := r.kotsAPI.RemoveNetwork(networkID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "remove network")
	} else {
		fmt.Printf("removed network %s\n", networkID)
	}
	return nil
}

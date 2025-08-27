package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage test networks for VMs and Clusters",
		Long: `The 'network' command allows you to manage and interact with networks used for testing purposes.
With this command you can list the networks in use by VMs and clusters.`,
		Example: `# List all networks
replicated network ls

# Update a network with an airgap policy
replicated network update <network-id> --policy airgap
`,
		Hidden: false,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) initNetworkClient() error {
	if apiToken == "" {
		creds, err := credentials.GetCurrentCredentials()
		if err != nil {
			return err
		}

		apiToken = creds.APIToken
	}

	httpClient := platformclient.NewHTTPClient(platformOrigin, apiToken)
	kotsAPI := &kotsclient.VendorV3Client{HTTPClient: *httpClient}
	r.kotsAPI = kotsAPI
	return nil
}

func (r *runners) completeNetworkNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := r.initNetworkClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	networks, err := r.kotsAPI.ListNetworks(nil, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, network := range networks {
		if network.Name != "" {
			names = append(names, network.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (r *runners) completeNetworkIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := r.initNetworkClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	networks, err := r.kotsAPI.ListNetworks(nil, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var ids []string
	for _, network := range networks {
		ids = append(ids, network.ID)
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

func (r *runners) completeNetworkIDsAndNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := r.initNetworkClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	networks, err := r.kotsAPI.ListNetworks(nil, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, network := range networks {
		completions = append(completions, network.ID)
		if network.Name != "" {
			completions = append(completions, network.Name)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func (r *runners) getNetworkIDFromArg(arg string) (string, error) {
	_, err := r.kotsAPI.GetNetwork(arg)
	if err == nil {
		return arg, nil
	}

	cause := errors.Cause(err)
	if cause != platformclient.ErrNotFound && cause != platformclient.ErrForbidden {
		return "", errors.Wrap(err, "get network")
	}

	networks, err := r.kotsAPI.ListNetworks(nil, nil)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return "", ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return "", errors.Wrap(err, "list networks")
	}

	var matchingNetworks []string
	for _, network := range networks {
		if network.Name == arg {
			matchingNetworks = append(matchingNetworks, network.ID)
		}
	}

	switch len(matchingNetworks) {
	case 0:
		return "", errors.Errorf("Network with name or ID '%s' not found", arg)
	case 1:
		return matchingNetworks[0], nil
	default:
		return "", errors.Errorf("Multiple networks found with name '%s'. Please use the network ID instead. Matching networks: %s. To view all network IDs run `replicated network ls`",
			arg,
			fmt.Sprintf("%s (and %d more)", matchingNetworks[0], len(matchingNetworks)-1))
	}
}

package cmd

import (
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "network",
		Short:  "Manage test networks for VMs and Clusters",
		Long:   ``,
		Hidden: true,
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

package cmd

import (
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Manage test vms",
		Long:  ``,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) initVMClient() error {
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

func (r *runners) completeVMIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := r.initVMClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var vmIDs []string
	vms, err := r.kotsAPI.ListVMs(false, nil, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	for _, vm := range vms {
		vmIDs = append(vmIDs, vm.ID)
	}
	return vmIDs, cobra.ShellCompDirectiveNoFileComp
}

func (r *runners) completeVMNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := r.initVMClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var vmNames []string
	vms, err := r.kotsAPI.ListVMs(false, nil, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	for _, vm := range vms {
		vmNames = append(vmNames, vm.Name)
	}
	return vmNames, cobra.ShellCompDirectiveNoFileComp
}

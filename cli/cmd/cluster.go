package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

var (
	ErrCompatibilityMatrixTermsNotAccepted = errors.New("You must read and accept the Compatibility Matrix Terms of Service before using this command. To view, please visit https://vendor.replicated.com/compatibility-matrix")
)

func (r *runners) InitClusterCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage test clusters",
		Long:  ``,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) initClient() error {
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

func (r *runners) completeClusterIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := r.initClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var clusterIDs []string
	clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	for _, cluster := range clusters {
		clusterIDs = append(clusterIDs, cluster.ID)
	}
	return clusterIDs, cobra.ShellCompDirectiveNoFileComp
}

func (r *runners) completeClusterNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	err := r.initClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var clusterNames []string
	clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	for _, cluster := range clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}
	return clusterNames, cobra.ShellCompDirectiveNoFileComp
}

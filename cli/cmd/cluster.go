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
		Short: "Manage test Kubernetes clusters.",
		Long:  `The 'cluster' command allows you to manage and interact with Kubernetes clusters used for testing purposes. With this command, you can create, list, remove, and manage node groups within clusters, as well as retrieve information about available clusters.`,
		Example: `  # Create a single-node EKS cluster
  replicated cluster create --distribution eks --version 1.31

  # List all clusters
  replicated cluster ls

  # Remove a specific cluster by ID
  replicated cluster rm <cluster-id>

  # Create a node group within a specific cluster
  replicated cluster nodegroup create --cluster-id <cluster-id> --instance-type m6.large --nodes 3`,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) initClusterClient() error {
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
	err := r.initClusterClient()
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
	err := r.initClusterClient()
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

package cmd

import (
	"github.com/pkg/errors"
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

func (r *runners) completeClusterIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	clusters, _ := r.kotsAPI.ListClusters(false, nil, nil)
	var clusterIDs []string
	for _, cluster := range clusters {
		clusterIDs = append(clusterIDs, cluster.ID)
	}
	return []string{"steve", "john"}, cobra.ShellCompDirectiveNoFileComp
}

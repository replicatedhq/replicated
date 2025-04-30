package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroup(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup",
		Short: "Manage node groups for clusters.",
		Long: `The 'cluster nodegroup' command provides functionality to manage node groups within a cluster. This command allows you to list node groups in a Kubernetes or VM-based cluster.

Node groups define a set of nodes with specific configurations, such as instance types, node counts, or scaling rules. You can use subcommands to perform various actions on node groups.`,
		Example: `# List all node groups for a cluster
replicated cluster nodegroup ls CLUSTER_ID_OR_NAME`,
	}
	parent.AddCommand(cmd)

	return cmd
}

package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortLs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls CLUSTER_ID_OR_NAME",
		Aliases: []string{"list"},
		Short:   "List cluster ports for a cluster.",
		Long: `The 'cluster port ls' command lists all the ports configured for a specific cluster. You must provide the cluster ID or name to retrieve and display the ports.

This command is useful for viewing the current port configurations, protocols, and other related settings of your test cluster. The output format can be customized to suit your needs, and the available formats include table, JSON, and wide views.`,
		Example: `# List ports for a cluster in the default table format
replicated cluster port ls CLUSTER_ID_OR_NAME

# List ports for a cluster in JSON format
replicated cluster port ls CLUSTER_ID_OR_NAME --output json

# List ports for a cluster in wide format
replicated cluster port ls CLUSTER_ID_OR_NAME --output wide`,
		RunE:              r.clusterPortList,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeClusterIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) clusterPortList(_ *cobra.Command, args []string) error {
	clusterID, err := r.getClusterIDFromArg(args[0])
	if err != nil {
		return err
	}

	ports, err := r.kotsAPI.ListClusterPorts(clusterID)
	if err != nil {
		return err
	}

	return print.ClusterPorts(r.outputFormat, ports, true)
}

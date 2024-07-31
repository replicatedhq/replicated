package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortLs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ls CLUSTER_ID",
		Short:             "List cluster ports for a cluster",
		RunE:              r.clusterPortList,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	return cmd
}

func (r *runners) clusterPortList(_ *cobra.Command, args []string) error {
	clusterID := args[0]

	ports, err := r.kotsAPI.ListClusterPorts(clusterID)
	if err != nil {
		return err
	}

	return print.ClusterPorts(r.outputFormat, r.w, ports, true)
}

package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortLs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "ls",
		RunE: r.clusterPortList,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

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

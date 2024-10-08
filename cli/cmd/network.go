package cmd

import (
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

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create networks for VMs and Clusters",
		Long:         ``,
		SilenceUsage: true,
		RunE:         r.createNetwork,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) createNetwork(_ *cobra.Command, args []string) error {
	return nil
}

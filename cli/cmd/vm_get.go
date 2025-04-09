package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitVMGetCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get information about a VM",
		Long:  `Get information about a VM, such as its SSH endpoint.`,
	}
	parent.AddCommand(cmd)

	return cmd
}

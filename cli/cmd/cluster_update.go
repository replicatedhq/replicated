package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update cluster settings",
		Long:  `cluster update can be used to update cluster settings`,
	}
	parent.AddCommand(cmd)

	return cmd
}

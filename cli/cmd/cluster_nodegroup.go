package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroupCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup",
		Short: "Manage nodegroups for compatibility matrix clusters",
		Long:  `Add, list, or delete node groups from compatibility matrix clusters`,
	}
	parent.AddCommand(cmd)

	return cmd
}

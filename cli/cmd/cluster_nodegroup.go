package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroup(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "nodegroup",
	}
	parent.AddCommand(cmd)

	return cmd
}

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitVMGroup(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "group",
	}
	parent.AddCommand(cmd)

	return cmd
}

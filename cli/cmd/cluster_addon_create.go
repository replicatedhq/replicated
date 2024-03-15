package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddonCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create cluster addons",
	}
	parent.AddCommand(cmd)

	return cmd
}

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddonCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create cluster add-ons",
	}
	parent.AddCommand(cmd)

	return cmd
}

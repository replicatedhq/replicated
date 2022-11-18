package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage registries",
		Long:  `registry can be used to manage existing registries and add new registries to a team`,
	}
	parent.AddCommand(cmd)

	return cmd
}

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitConfigCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage .replicated configuration",
		Long:  `Manage .replicated configuration files for your project.`,
	}

	parent.AddCommand(cmd)
	return cmd
}

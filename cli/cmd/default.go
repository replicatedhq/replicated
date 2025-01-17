package cmd

import "github.com/spf13/cobra"

func (r *runners) InitDefaultCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "default",
		Short:   "Manage default values used by other commands",
		Long:    ``,
		Example: `  `,
	}

	parent.AddCommand(cmd)

	return cmd
}

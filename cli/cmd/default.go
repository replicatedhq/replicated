package cmd

import "github.com/spf13/cobra"

func (r *runners) InitDefaultCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "default",
	}

	parent.AddCommand(cmd)

	return cmd
}

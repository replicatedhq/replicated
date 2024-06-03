package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitModelCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "model",
		Short:  "Manage models and collections",
		Long:   `model can be used to manage existing models and collections`,
		Hidden: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

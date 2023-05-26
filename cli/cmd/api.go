package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAPICommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "api",
		Short:  "Make ad-hoc API calls to the Replicated API",
		Long:   ``,
		Hidden: false,
	}
	parent.AddCommand(cmd)

	return cmd
}

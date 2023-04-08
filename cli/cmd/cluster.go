package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "cluster",
		Short:  "Manage test clusters",
		Long:   ``,
		Hidden: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPort(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "port",
		SilenceUsage: true,
		Hidden:       false,
	}
	parent.AddCommand(cmd)

	return cmd
}

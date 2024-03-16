package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPort(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "port",
		SilenceUsage: true,
		Hidden:       true, // this feature is not fully implemented and controlled behind a feature toggle in the api until ready
	}
	parent.AddCommand(cmd)

	return cmd
}

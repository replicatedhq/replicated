package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAppCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage apps",
		Long:  `app can be used to list apps and create new apps`,
	}
	parent.AddCommand(cmd)

	return cmd
}

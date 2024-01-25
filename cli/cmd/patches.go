package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitPatchesCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patches",
		Short: "View and manage application patches and recommendations",
		Long:  ``,
	}
	parent.AddCommand(cmd)

	return cmd
}

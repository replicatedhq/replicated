package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorsCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collector",
		Short: "Manage customer collectors",
		Long:  `The collector command allows vendors to create, display, modify entitlement values for end customer licensing.`,
	}
	cmd.Hidden=true; // Not supported in KOTS (ch #22646)
	parent.AddCommand(cmd)

	var tmp bool
	cmd.Flags().BoolVar(&tmp, "active", false, "Only show active collectors")

	return cmd
}

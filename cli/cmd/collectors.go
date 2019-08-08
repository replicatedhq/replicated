package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorsCommand(parent *cobra.Command) *cobra.Command {
	collectorsCmd := &cobra.Command{
		Use:   "collector",
		Short: "Manage customer collectors",
		Long:  `The collector command allows vendors to create, display, modify entitlement values for end customer licensing.`,
	}
	parent.AddCommand(collectorsCmd)

	var tmp bool
	collectorsCmd.Flags().BoolVar(&tmp, "active", false, "Only show active collectors")

	return collectorsCmd
}

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectionCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Manage model collections",
		Long:  `collection can be used to manage model collections`,
	}
	parent.AddCommand(cmd)

	return cmd
}

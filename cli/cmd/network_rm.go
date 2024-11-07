package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm ID [ID …]",
		Aliases: []string{"delete"},
		Short:   "Remove test network",
		Long:    ``,
		RunE:    r.removeNetworks,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) removeNetworks(_ *cobra.Command, args []string) error {
	return nil
}

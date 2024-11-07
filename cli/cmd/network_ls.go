package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List test networks",
		Long:    ``,
		RunE:    r.listNetworks,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) listNetworks(_ *cobra.Command, args []string) error {
	return nil
}

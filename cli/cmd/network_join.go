package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkJoin(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join a test network",
		Long:  ``,
		RunE:  r.joinNetwork,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) joinNetwork(_ *cobra.Command, args []string) error {
	return nil
}

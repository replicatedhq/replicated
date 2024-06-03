package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitModelPush(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "push",
		Short:        "",
		Long:         ``,
		RunE:         r.pushModel,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) pushModel(_ *cobra.Command, args []string) error {
	return nil
}

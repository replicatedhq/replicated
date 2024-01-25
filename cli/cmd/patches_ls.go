package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitPatchesList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List patches",
		Long:  `List patches`,
		RunE:  r.listPatches,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) listPatches(_ *cobra.Command, args []string) error {
	return nil
}

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddOnRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "rm",
		RunE: r.addOnClusterRm,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) addOnClusterRm(_ *cobra.Command, args []string) error {
	err := r.kotsAPI.DeleteClusterAddOn(args[0])
	if err != nil {
		return err
	}

	return nil
}

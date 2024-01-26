package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddOnLs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "ls",
		RunE: r.addOnClusterLs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) addOnClusterLs(_ *cobra.Command, args []string) error {
	addons, err := r.kotsAPI.ListClusterAddOns()
	if err != nil {
		return err
	}

	return print.AddOns(r.outputFormat, r.w, addons, true)
}

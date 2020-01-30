package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all of an app's collectors",
		Long:  "List all of an app's collectors",
	}
	cmd.Hidden=true; // Not supported in KOTS 
	parent.AddCommand(cmd)
	cmd.RunE = r.collectorList
}

func (r *runners) collectorList(cmd *cobra.Command, args []string) error {
	collectors, err := r.api.ListCollectors(r.appID, r.appType)
	if err != nil {
		return err
	}

	return print.Collectors(r.stdoutWriter, collectors)
}

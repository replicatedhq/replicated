package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all of an app's collectors",
		Long:    "List all of an app's collectors",
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)
	cmd.RunE = r.collectorList
}

func (r *runners) collectorList(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	collectors, err := r.api.ListCollectors(r.appID, r.appType)
	if err != nil {
		return err
	}

	return print.Collectors(r.w, collectors)
}

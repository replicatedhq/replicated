package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitModelList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls",
		Aliases:      []string{"list"},
		Short:        "list models",
		Long:         `list models`,
		RunE:         r.listModels,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) listModels(_ *cobra.Command, args []string) error {
	models, err := r.kotsAPI.ListModels()
	if err != nil {
		return err
	}

	return print.Models(r.outputFormat, r.w, models)
}

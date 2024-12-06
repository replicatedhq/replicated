package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitModelRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rm NAME",
		Aliases:      []string{"delete"},
		Short:        "remove a model",
		Long:         `remove a model`,
		RunE:         r.removeModel,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) removeModel(_ *cobra.Command, args []string) error {
	models, err := r.kotsAPI.RemoveModel(args[0])
	if err != nil {
		return err
	}

	// we delete from the oci repository here

	return print.Models(r.outputFormat, r.w, models)
}

package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitModelList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls",
		Short:        "list models",
		Long:         `list models`,
		RunE:         r.listModels,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listModels(_ *cobra.Command, args []string) error {
	return nil
}

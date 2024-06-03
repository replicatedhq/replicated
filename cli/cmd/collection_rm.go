package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectionRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls [NAME]",
		Short:        "list registries",
		Long:         `list registries, or a single registry by name`,
		RunE:         r.listRegistries,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) removeCollection(_ *cobra.Command, args []string) error {
	return nil
}

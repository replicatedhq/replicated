package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectionList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls",
		Aliases:      []string{"list"},
		Short:        "list model collections",
		Long:         `list model collections`,
		RunE:         r.listCollections,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listCollections(_ *cobra.Command, args []string) error {
	collections, err := r.kotsAPI.ListCollections()
	if err != nil {
		return errors.Wrap(err, "list model collections")
	}

	return print.ModelCollections(r.outputFormat, r.w, collections)
}

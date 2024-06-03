package cmd

import (
	"fmt"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectionCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create a collection",
		Long:         `create a new collection to group ai models together`,
		RunE:         r.createCollection,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.modelCollectionCreateName, "name", "", "The name of the collection")
	cmd.MarkFlagRequired("name")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) createCollection(_ *cobra.Command, args []string) error {
	fmt.Printf("Creating collection %s\n", r.args.modelCollectionCreateName)
	collection, err := r.kotsAPI.CreateCollection(r.args.modelCollectionCreateName)
	if err != nil {
		return err
	}

	return print.ModelCollections(r.outputFormat, r.w, []types.ModelCollection{*collection})
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitCollectionRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rm COLLECTION_ID",
		Aliases:      []string{"delete"},
		Short:        "remove a collection",
		Long:         `remove a collection, unlinking any model from the collection (but not deleting the model)`,
		RunE:         r.removeCollection,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) removeCollection(_ *cobra.Command, args []string) error {
	collectionID := args[0]
	err := r.kotsAPI.DeleteCollection(collectionID)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.w, "Collection %s has been deleted\n", collectionID)
	return err
}

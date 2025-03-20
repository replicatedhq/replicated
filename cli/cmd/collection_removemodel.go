package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectionRemoveModel(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "remove-model",
		Short:        "remove a model from a collection",
		Long:         `remove a model from a collection`,
		RunE:         r.removeModelFromCollection,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.modelCollectionRmModelName, "model-name", "", "The name of the model")
	cmd.MarkFlagRequired("model-id")
	cmd.Flags().StringVar(&r.args.modelCollectionRmModelCollectionID, "collection-id", "", "The ID of the collection")
	cmd.MarkFlagRequired("collection-id")

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) removeModelFromCollection(_ *cobra.Command, args []string) error {
	err := r.kotsAPI.UpdateModelsInCollection(r.args.modelCollectionRmModelCollectionID, nil, []string{r.args.modelCollectionRmModelName})
	if err != nil {
		return errors.Wrap(err, "list model collections")
	}

	fmt.Fprintf(r.w, "Model %s has been removed from collection %s\n", r.args.modelCollectionRmModelName, r.args.modelCollectionRmModelCollectionID)
	return nil
}

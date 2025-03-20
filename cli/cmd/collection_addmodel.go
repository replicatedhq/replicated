package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCollectionAddModel(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "add-model",
		Short:        "add a model to a collection",
		Long:         `add a model to a collection`,
		RunE:         r.addModelToCollection,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.modelCollectionAddModelName, "model-name", "", "The name of the model")
	cmd.MarkFlagRequired("model-id")
	cmd.Flags().StringVar(&r.args.modelCollectionAddModelCollectionID, "collection-id", "", "The ID of the collection")
	cmd.MarkFlagRequired("collection-id")

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) addModelToCollection(_ *cobra.Command, args []string) error {
	err := r.kotsAPI.UpdateModelsInCollection(r.args.modelCollectionAddModelCollectionID, []string{r.args.modelCollectionAddModelName}, nil)
	if err != nil {
		return errors.Wrap(err, "list model collections")
	}

	fmt.Fprintf(r.w, "Model %s has been added to collection %s\n", r.args.modelCollectionAddModelName, r.args.modelCollectionAddModelCollectionID)
	return nil
}

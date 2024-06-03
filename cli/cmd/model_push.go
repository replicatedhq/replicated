package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitModelPush(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "push",
		Short:        "push a model to the model repository",
		Long:         `push a model to the mdoel repository`,
		RunE:         r.pushModel,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.modelPushName, "name", "", "The name of the model to push")
	cmd.MarkFlagRequired("name")

	return cmd
}

func (r *runners) pushModel(_ *cobra.Command, args []string) error {
	path := args[0]

	// THIS IS JUST A PLACEHOLDER
	// IT WILL ACTUALLY PUSH TO OCI
	// BUT NOW IT'S CALLING AN API TO SIMULATE

	err := r.kotsAPI.PushModel(r.args.modelPushName, path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.w, "Model %s has been pushed\n", r.args.modelPushName)
	return err
}

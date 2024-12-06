package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rm [ENDPOINT]",
		Aliases:      []string{"delete"},
		Short:        "remove registry",
		Long:         `remove registry by endpoint`,
		RunE:         r.removeRegistry,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) removeRegistry(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("missing registry endpoint")
	}

	if err := r.kotsAPI.RemoveRegistry(args[0]); err != nil {
		return err
	}

	fmt.Println("Registry removed")

	return nil
}

package cmd

import "github.com/spf13/cobra"

func (r *runners) InitDefaultClearCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "clear",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.clearDefault(cmd, args[0])
		},
	}

	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) clearDefault(cmd *cobra.Command, defaultType string) error {
	if err := cache.ClearDefault(defaultType); err != nil {
		return err
	}

	return nil
}

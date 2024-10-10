package cmd

import "github.com/spf13/cobra"

func (r *runners) InitDefaultClearAllCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "clear-all",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.clearAllDefaults(cmd)
		},
	}

	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) clearAllDefaults(cmd *cobra.Command) error {
	if err := cache.ClearDefault("app"); err != nil {
		return err
	}

	return nil
}

package cmd

import "github.com/spf13/cobra"

func (r *runners) InitDefaultClearAllCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-all",
		Short: "Clear all default values",
		Long: `Clears all default values that are used by other commands.

This command removes all default values that are used by other commands run by the current user.`,
		Example: `  # Clear all default values
  replicated default clear-all`,
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

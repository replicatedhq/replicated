package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitDefaultClearCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear KEY",
		Short: "Clear default value for a key",
		Long: `Clears default value for the specified key.

This command removes default values that are used by other commands run by the current user.

Supported keys:
- app: the default application to use`,
		Example: `# Clear default application
replicated default clear app`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.clearDefault(cmd, args[0])
		},
	}

	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) clearDefault(cmd *cobra.Command, defaultType string) error {
	if err := cache.ClearDefault(defaultType); err != nil {
		return errors.Wrap(err, "clear default")
	}

	return nil
}

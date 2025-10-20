package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/spf13/cobra"
)

func (r *runners) InitProfileSetDefaultCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default [profile-name]",
		Short: "Set the default authentication profile",
		Long: `Set the default authentication profile that will be used when no profile is specified
in .replicated.yaml and no environment variables are set.`,
		Example: `# Set production as the default profile
replicated profile set-default prod`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         r.profileSetDefault,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) profileSetDefault(_ *cobra.Command, args []string) error {
	profileName := args[0]

	if profileName == "" {
		return errors.New("profile name cannot be empty")
	}

	// Check if profile exists
	_, err := credentials.GetProfile(profileName)
	if err == credentials.ErrProfileNotFound {
		return errors.Errorf("profile '%s' not found", profileName)
	}
	if err != nil {
		return errors.Wrap(err, "failed to get profile")
	}

	// Set as default
	if err := credentials.SetDefaultProfile(profileName); err != nil {
		return errors.Wrap(err, "failed to set default profile")
	}

	fmt.Printf("Default profile set to '%s'\n", profileName)
	return nil
}

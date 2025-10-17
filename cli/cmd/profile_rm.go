package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/spf13/cobra"
)

func (r *runners) InitProfileRmCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [profile-name]",
		Short: "Remove an authentication profile",
		Long: `Remove an authentication profile by name.

If the removed profile was the default profile, the default will be automatically
set to another available profile (if any exist).`,
		Example: `# Remove a profile
replicated profile rm dev`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         r.profileRm,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) profileRm(_ *cobra.Command, args []string) error {
	profileName := args[0]

	if profileName == "" {
		return errors.New("profile name cannot be empty")
	}

	// Check if profile exists before removing
	_, err := credentials.GetProfile(profileName)
	if err == credentials.ErrProfileNotFound {
		return errors.Errorf("profile '%s' not found", profileName)
	}
	if err != nil {
		return errors.Wrap(err, "failed to get profile")
	}

	// Remove the profile
	if err := credentials.RemoveProfile(profileName); err != nil {
		return errors.Wrap(err, "failed to remove profile")
	}

	fmt.Printf("Profile '%s' removed successfully\n", profileName)

	// Check if there's a new default
	_, newDefault, err := credentials.ListProfiles()
	if err != nil {
		return errors.Wrap(err, "failed to check new default profile")
	}

	if newDefault != "" {
		fmt.Printf("Default profile is now '%s'\n", newDefault)
	} else {
		fmt.Println("No profiles remaining")
	}

	return nil
}

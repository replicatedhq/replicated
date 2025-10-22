package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/spf13/cobra"
)

func (r *runners) InitProfileEditCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [profile-name]",
		Short: "Edit an existing authentication profile",
		Long: `Edit an existing authentication profile.

You can update the API token, API origin, and/or registry origin for an existing profile.
Only the flags you provide will be updated; other fields will remain unchanged.

The profile will be stored in ~/.replicated/config.yaml with file permissions 600 (owner read/write only).`,
		Example: `# Update the token for a profile
replicated profile edit dev --token=new-dev-token

# Update the API origin for a profile
replicated profile edit dev --api-origin=https://vendor-api-noahecampbell.okteto.repldev.com

# Update multiple fields at once
replicated profile edit dev \
  --token=new-token \
  --api-origin=https://vendor-api-noahecampbell.okteto.repldev.com \
  --registry-origin=vendor-registry-v2-noahecampbell.okteto.repldev.com`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         r.profileEdit,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.profileEditToken, "token", "", "New API token for this profile (optional)")
	cmd.Flags().StringVar(&r.args.profileEditAPIOrigin, "api-origin", "", "New API origin (optional, e.g., https://api.replicated.com/vendor). Mutually exclusive with --namespace")
	cmd.Flags().StringVar(&r.args.profileEditRegistryOrigin, "registry-origin", "", "New registry origin (optional, e.g., registry.replicated.com). Mutually exclusive with --namespace")
	cmd.Flags().StringVar(&r.args.profileEditNamespace, "namespace", "", "Okteto namespace for dev environments (e.g., 'noahecampbell'). Auto-generates service URLs. Mutually exclusive with --api-origin and --registry-origin")

	return cmd
}

func (r *runners) profileEdit(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	if profileName == "" {
		return errors.New("profile name cannot be empty")
	}

	// Check for mutually exclusive flags
	hasNamespace := cmd.Flags().Changed("namespace")
	hasAPIOrigin := cmd.Flags().Changed("api-origin")
	hasRegistryOrigin := cmd.Flags().Changed("registry-origin")

	if hasNamespace && (hasAPIOrigin || hasRegistryOrigin) {
		return errors.New("--namespace cannot be used with --api-origin or --registry-origin. Use --namespace for dev environments, or use explicit origins for custom endpoints")
	}

	// Load existing profile
	profile, err := credentials.GetProfile(profileName)
	if err != nil {
		return errors.Wrapf(err, "failed to load profile '%s'. Use 'replicated profile ls' to see available profiles", profileName)
	}

	// Track if any changes were made
	changed := false

	// Update token if provided
	if cmd.Flags().Changed("token") {
		profile.APIToken = r.args.profileEditToken
		changed = true
	}

	// Update namespace if provided (clears explicit origins)
	if cmd.Flags().Changed("namespace") {
		profile.Namespace = r.args.profileEditNamespace
		// Clear explicit origins when using namespace
		profile.APIOrigin = ""
		profile.RegistryOrigin = ""
		changed = true
	}

	// Update API origin if provided (clears namespace)
	if cmd.Flags().Changed("api-origin") {
		profile.APIOrigin = r.args.profileEditAPIOrigin
		profile.Namespace = "" // Clear namespace when using explicit origin
		changed = true
	}

	// Update registry origin if provided (clears namespace)
	if cmd.Flags().Changed("registry-origin") {
		profile.RegistryOrigin = r.args.profileEditRegistryOrigin
		profile.Namespace = "" // Clear namespace when using explicit origin
		changed = true
	}

	if !changed {
		return errors.New("no changes specified. Use --token, --namespace, --api-origin, or --registry-origin to update the profile")
	}

	// Save the updated profile (dereference the pointer)
	if err := credentials.AddProfile(profileName, *profile); err != nil {
		return errors.Wrap(err, "failed to update profile")
	}

	fmt.Printf("Profile '%s' updated successfully\n", profileName)
	if cmd.Flags().Changed("api-origin") {
		if profile.APIOrigin != "" {
			fmt.Printf("  API Origin: %s\n", profile.APIOrigin)
		} else {
			fmt.Printf("  API Origin: (removed, using default)\n")
		}
	}
	if cmd.Flags().Changed("registry-origin") {
		if profile.RegistryOrigin != "" {
			fmt.Printf("  Registry Origin: %s\n", profile.RegistryOrigin)
		} else {
			fmt.Printf("  Registry Origin: (removed, using default)\n")
		}
	}

	return nil
}

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/credentials"
	"github.com/replicatedhq/replicated/pkg/credentials/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitProfileAddCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [profile-name]",
		Short: "Add a new authentication profile",
		Long: `Add a new authentication profile with the specified name.

You must provide an API token. Optionally, you can specify custom API and registry origins.
If a profile with the same name already exists, it will be updated.

The profile will be stored in ~/.replicated/config.yaml with file permissions 600 (owner read/write only).`,
		Example: `# Add a production profile
replicated profile add prod --token=your-prod-token

# Add a development profile with custom origins
replicated profile add dev \
  --token=your-dev-token \
  --api-origin=https://vendor-api-noahecampbell.okteto.repldev.com \
  --registry-origin=vendor-registry-v2-noahecampbell.okteto.repldev.com`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         r.profileAdd,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.profileAddToken, "token", "", "API token for this profile (required)")
	cmd.Flags().StringVar(&r.args.profileAddAPIOrigin, "api-origin", "", "API origin (optional, e.g., https://api.replicated.com/vendor)")
	cmd.Flags().StringVar(&r.args.profileAddRegistryOrigin, "registry-origin", "", "Registry origin (optional, e.g., registry.replicated.com)")

	cmd.MarkFlagRequired("token")

	return cmd
}

func (r *runners) profileAdd(_ *cobra.Command, args []string) error {
	profileName := args[0]

	if profileName == "" {
		return errors.New("profile name cannot be empty")
	}

	if r.args.profileAddToken == "" {
		return errors.New("token is required")
	}

	profile := types.Profile{
		APIToken:       r.args.profileAddToken,
		APIOrigin:      r.args.profileAddAPIOrigin,
		RegistryOrigin: r.args.profileAddRegistryOrigin,
	}

	if err := credentials.AddProfile(profileName, profile); err != nil {
		return errors.Wrap(err, "failed to add profile")
	}

	fmt.Printf("Profile '%s' added successfully\n", profileName)

	// Check if this is the only profile - if so, it's now the default
	_, defaultProfile, err := credentials.ListProfiles()
	if err != nil {
		return errors.Wrap(err, "failed to check default profile")
	}

	if defaultProfile == profileName {
		fmt.Printf("Profile '%s' set as default\n", profileName)
	}

	return nil
}

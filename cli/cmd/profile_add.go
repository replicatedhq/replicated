package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
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

You can provide an API token via the --token flag, or you will be prompted to enter it securely.
Optionally, you can specify custom API and registry origins.
If a profile with the same name already exists, it will be updated.

The profile will be stored in ~/.replicated/config.yaml with file permissions 600 (owner read/write only).`,
		Example: `# Add a production profile (will prompt for token)
replicated profile add prod

# Add a production profile with token flag
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

	cmd.Flags().StringVar(&r.args.profileAddToken, "token", "", "API token for this profile (optional, will prompt if not provided)")
	cmd.Flags().StringVar(&r.args.profileAddAPIOrigin, "api-origin", "", "API origin (optional, e.g., https://api.replicated.com/vendor). Mutually exclusive with --namespace")
	cmd.Flags().StringVar(&r.args.profileAddRegistryOrigin, "registry-origin", "", "Registry origin (optional, e.g., registry.replicated.com). Mutually exclusive with --namespace")
	cmd.Flags().StringVar(&r.args.profileAddNamespace, "namespace", "", "Okteto namespace for dev environments (e.g., 'noahecampbell'). Auto-generates service URLs. Mutually exclusive with --api-origin and --registry-origin")

	return cmd
}

func (r *runners) profileAdd(cmd *cobra.Command, args []string) error {
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

	// If token is not provided via flag, prompt for it securely
	token := r.args.profileAddToken
	if token == "" {
		var err error
		token, err = r.readAPITokenFromPrompt("API Token")
		if err != nil {
			return errors.Wrap(err, "failed to read API token")
		}
	}

	profile := types.Profile{
		APIToken:       token,
		APIOrigin:      r.args.profileAddAPIOrigin,
		RegistryOrigin: r.args.profileAddRegistryOrigin,
		Namespace:      r.args.profileAddNamespace,
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

func (r *runners) readAPITokenFromPrompt(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Mask:  '*',
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("API token cannot be empty")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", errors.New("interrupted")
		}
		return "", err
	}

	return result, nil
}

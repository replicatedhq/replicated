package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitProfileCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage authentication profiles",
		Long: `The profile command allows you to manage authentication profiles for the Replicated CLI.

Profiles let you store multiple sets of credentials and easily switch between them.
This is useful when working with different Replicated accounts (production, development, etc.)
or different API endpoints.

Credentials are stored in ~/.replicated/config.yaml with file permissions set to 600 (owner read/write only).
You can reference profiles in your .replicated.yaml files using the 'profile' field.

Authentication priority:
1. REPLICATED_API_TOKEN environment variable (highest priority)
2. Profile specified in .replicated.yaml
3. Default profile from ~/.replicated/config.yaml
4. Legacy single token (backward compatibility)

Use the various subcommands to:
- Add new profiles
- List all profiles
- Remove profiles
- Set the default profile`,
		Example: `# Add a production profile
replicated profile add prod --token=your-prod-token

# Add a development profile with custom API origin
replicated profile add dev --token=your-dev-token --api-origin=https://vendor-api-dev.com

# List all profiles
replicated profile ls

# Set default profile
replicated profile set-default prod

# Remove a profile
replicated profile rm dev`,
	}
	parent.AddCommand(cmd)

	return cmd
}

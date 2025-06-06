package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAppCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage applications",
		Long: `The app command allows you to manage applications in your Replicated account.

This command provides a suite of subcommands for creating, listing, and
deleting applications. You can perform operations such as creating new apps,
viewing app details, and removing apps from your account.

Use the various subcommands to:
- Create new applications
- List all existing applications
- Delete applications from your account`,
		Example: `# List all applications
replicated app ls

# Create a new application
replicated app create "My New App"

# Delete an application
replicated app rm "app-slug"

# List applications with custom output format
replicated app ls --output json`,
	}
	parent.AddCommand(cmd)

	return cmd
}

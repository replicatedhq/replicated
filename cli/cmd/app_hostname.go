package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAppHostnameCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hostname",
		Short: "Manage custom hostnames for applications",
		Long: `The hostname command allows you to manage custom hostnames for your application.

This command provides subcommands for listing and viewing custom hostname configurations
including registry, proxy, download portal, and replicated app hostnames.`,
		Example: `# List all custom hostnames for an app
replicated app hostname ls --app myapp

# List hostnames and output as JSON
replicated app hostname ls --app myapp --output json`,
	}
	parent.AddCommand(cmd)

	return cmd
}

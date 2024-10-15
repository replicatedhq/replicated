package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalUserCmd(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Enterprise Portal users",
		Long: `The user command allows you to manage users for the Enterprise Portal.

This command provides subcommands for listing, adding, and removing users
from the Enterprise Portal. You can use these subcommands to control access
to the Enterprise Portal for your application.`,
		Example: `  # List all users with access to the Enterprise Portal
  replicated enterprise-portal user ls

  # List users for a specific application
  replicated enterprise-portal user ls --app myapp`,
	}
	parent.AddCommand(cmd)

	return cmd
}

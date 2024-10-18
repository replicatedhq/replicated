package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalStatusCmd(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Manage the status of the Enterprise Portal",
		Long: `The status command allows you to view and update the status of the Enterprise Portal.

This command provides subcommands for getting the current status of the Enterprise Portal
and updating it. You can use these subcommands to monitor and control the state of the
Enterprise Portal for your application.

Use 'status get' to retrieve the current status and 'status update' to change the status.`,
		Example: `  # Get the current status of the Enterprise Portal
  replicated enterprise-portal status get

  # Update the status of the Enterprise Portal
  replicated enterprise-portal status update --status active

  # Get the status for a specific application
  replicated enterprise-portal status get --app myapp

  # Update the status for a specific application
  replicated enterprise-portal status update --app myapp --status inactive`,
	}
	parent.AddCommand(cmd)

	return cmd
}

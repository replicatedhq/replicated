package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type updateEnterprisePortalStatusOpts struct {
	status string
}

func (r *runners) InitEnterprisePortalStatusUpdateCmd(parent *cobra.Command) *cobra.Command {
	opts := updateEnterprisePortalStatusOpts{}
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the status of the Enterprise Portal",
		Long: `Update the current status of the Enterprise Portal for the specified application.

This command allows you to change the status of the Enterprise Portal associated
with the current application. You can use this to activate, deactivate, or set
any other valid status for the Enterprise Portal.

If no application is specified, the command will use the default application
set in your configuration.

The new status must be provided using the --status flag.`,
		Example: `# Update the Enterprise Portal status for the default application
replicated enterprise-portal status update --status active

# Update the Enterprise Portal status for a specific application
replicated enterprise-portal status update --app myapp --status inactive

# Update the Enterprise Portal status and output in JSON format
replicated enterprise-portal status update --status pending --output json

# Update the Enterprise Portal status and output in table format (default)
replicated enterprise-portal status update --status active --output table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.enterprisePortalStatusUpdate(cmd, r.appID, opts, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&opts.status, "status", "", "The status to set for the enterprise portal")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	cmd.MarkFlagRequired("status")

	return cmd
}

func (r *runners) enterprisePortalStatusUpdate(cmd *cobra.Command, appID string, opts updateEnterprisePortalStatusOpts, outputFormat string) error {
	status, err := r.kotsAPI.UpdateEnterprisePortalStatus(appID, opts.status)
	if err != nil {
		return err
	}

	fmt.Printf("Enterprise Portal Status: %s\n", status)
	return nil
}

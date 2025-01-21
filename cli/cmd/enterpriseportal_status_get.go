package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalStatusGetCmd(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get the status of the Enterprise Portal",
		Long: `Retrieve the current status of the Enterprise Portal for the specified application.

This command fetches and displays the status of the Enterprise Portal associated
with the current application. The status information can help you understand
whether the Enterprise Portal is active, pending, or in any other state.

If no application is specified, the command will use the default application
set in your configuration.`,
		Example: `# Get the Enterprise Portal status for the default application
replicated enterprise-portal status get

# Get the Enterprise Portal status for a specific application
replicated enterprise-portal status get --app myapp

# Get the Enterprise Portal status and output in JSON format
replicated enterprise-portal status get --output json

# Get the Enterprise Portal status and output in table format (default)
replicated enterprise-portal status get --output table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.enterprisePortalStatusGet(cmd, r.appID, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) enterprisePortalStatusGet(cmd *cobra.Command, appID string, outputFormat string) error {
	status, err := r.kotsAPI.GetEnterprisePortalStatus(appID)
	if err != nil {
		return errors.Wrap(err, "get enterprise portal status")
	}

	fmt.Printf("Enterprise Portal Status: %s\n", status)
	return nil
}

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
		Use: "update",

		RunE: func(cmd *cobra.Command, args []string) error {
			return r.enterprisePortalStatusUpdate(cmd, r.appID, opts, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&opts.status, "status", "", "The status to set for the enterprise portal")
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

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

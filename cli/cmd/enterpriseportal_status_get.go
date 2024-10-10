package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalStatusGetCmd(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use: "get",

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

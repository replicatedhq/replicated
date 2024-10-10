package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enterprise-portal",
		Short:   "Manage enterprise portal",
		Long:    ``,
		Example: `  `,
	}
	parent.AddCommand(cmd)

	return cmd
}

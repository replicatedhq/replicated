package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalStatusCmd(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "status",
	}
	parent.AddCommand(cmd)

	return cmd
}

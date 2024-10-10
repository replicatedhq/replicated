package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalUserCmd(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "user",
	}
	parent.AddCommand(cmd)

	return cmd
}

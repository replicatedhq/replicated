package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseAuth(parent *cobra.Command) *cobra.Command {
	enterpriseAuthCommand := &cobra.Command{
		Use:   "auth",
		Short: "Manage enterprise authentication",
		Long:  `The auth command manages authentication`,
	}
	parent.AddCommand(enterpriseAuthCommand)

	return enterpriseAuthCommand
}

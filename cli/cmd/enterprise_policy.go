package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePolicy(parent *cobra.Command) *cobra.Command {
	enterprisePolicyCommand := &cobra.Command{
		Use:   "policy",
		Short: "Manage enterprise policies",
		Long:  `The policy command allows approved enterprise to create and manage policies`,
	}
	parent.AddCommand(enterprisePolicyCommand)

	return enterprisePolicyCommand
}

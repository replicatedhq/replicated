package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitPolicyCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage RBAC policies",
		Long:  "The policy command allows vendors to list, create, update, and remove RBAC policies.",
	}
	parent.AddCommand(cmd)
	return cmd
}

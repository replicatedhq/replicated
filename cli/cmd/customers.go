package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersCommand(parent *cobra.Command) *cobra.Command {
	customersCmd := &cobra.Command{
		Use:   "customer",
		Short: "Manage customers",
		Long:  `The customers command allows vendors to create, display, modify end customer records.`,
	}
	parent.AddCommand(customersCmd)

	return customersCmd
}

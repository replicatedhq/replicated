package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomerInstancesCommand(parent *cobra.Command) *cobra.Command {
	customerInstancesCmd := &cobra.Command{
		Use:   "instances",
		Short: "Inspect customer instances",
		Long:  `The customer instances command allows vendors to list and inspect running customer instances.`,
	}
	parent.AddCommand(customerInstancesCmd)

	return customerInstancesCmd
}

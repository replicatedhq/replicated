package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomerLicenseCommand(parent *cobra.Command) *cobra.Command {
	customersCmd := &cobra.Command{
		Use:   "license",
		Short: "Manage customer licenses",
		Long:  `The license command allows vendors to inspect and fetch customer licenses.`,
	}
	parent.AddCommand(customersCmd)

	return customersCmd
}

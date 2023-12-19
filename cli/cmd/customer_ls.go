package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersLSCommand(parent *cobra.Command) *cobra.Command {
	customersCmd := &cobra.Command{
		Use:   "ls",
		Short: "list customers",
		Long:  `list customers`,
		RunE:  r.listCustomers,
	}
	// Example to list customers by app version and app
	// replicated customer ls --app <appID> --app-version <appVersion>
	parent.AddCommand(customersCmd)
	customersCmd.Flags().StringVar(&r.args.lsAppVersion, "app-version", "", "List customers and their instances by app version")
	customersCmd.Flags().BoolVar(&r.args.customerLsIncludeTest, "include-test", false, "Include test customers in the list")
	customersCmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return customersCmd
}

func (r *runners) listCustomers(_ *cobra.Command, _ []string) error {

	// get appVersion from flags
	lsappVersion := r.args.lsAppVersion
	// if appVersion is blank, call ListCustomers
	if lsappVersion == "" {
		customers, err := r.api.ListCustomers(r.appID, r.appType, r.args.customerLsIncludeTest)
		if err != nil {
			return errors.Wrap(err, "list customers")
		}
		return print.Customers(r.outputFormat, r.w, customers)
	} else {
		// call ListCustomersByAppAndVersion
		customers, err := r.api.ListCustomersByAppAndVersion(r.appID, lsappVersion, r.appType)
		// if err and outputFormat is json, customers should be a blank struct and print will return []
		if err != nil && r.outputFormat == "json" {
			return print.CustomersWithInstances(r.outputFormat, r.w, customers)
			// error and outputFormat is table, print error
		} else if err != nil {
			return errors.Wrap(err, "list customers by app and app version")
		}
		return print.CustomersWithInstances(r.outputFormat, r.w, customers)
	}
}

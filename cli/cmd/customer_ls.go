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
	parent.AddCommand(customersCmd)
	customersCmd.Flags().BoolVar(&r.args.customerLsIncludeTest, "include-test", false, "Include test customers in the list")
	customersCmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return customersCmd
}

func (r *runners) listCustomers(_ *cobra.Command, _ []string) error {
	customers, err := r.api.ListCustomers(r.appID, r.appType, r.args.customerLsIncludeTest)
	if err != nil {
		return errors.Wrap(err, "list customers")
	}

	return print.Customers(r.outputFormat, r.w, customers)
}

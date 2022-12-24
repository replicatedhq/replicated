package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomerInstancesLSCommand(parent *cobra.Command) *cobra.Command {
	customerInstanceLSCommand := &cobra.Command{
		Use:   "ls CUSTOMER",
		Short: "list customer instances",
		Long:  `list instances for a single customer`,
		RunE:  r.listCustomerInstances,
	}
	parent.AddCommand(customerInstanceLSCommand)

	// need to add this to the base runner
	customerInstanceLSCommand.Flags().StringVarP(&r.args.output, "output", "o", "text", "options are text, json, wide")

	return customerInstanceLSCommand
}

func (r *runners) listCustomerInstances(_ *cobra.Command, args []string) error {

	if r.appType != "kots" {
		return errors.Errorf("unsupported app type: %q, only kots applications are supported", r.appType)
	}

	if len(args) != 1 {
		return errors.Errorf("requires exactly one customer name or ID")
	}
	customerNameOrID := args[0]

	customer, err := r.api.GetCustomerByNameOrID(r.appType, r.appID, customerNameOrID)
	if err != nil {
		return errors.Wrapf(err, "get customer %q", customerNameOrID)
	}

	instances, err := r.api.ListCustomerInstances(r.appType, *customer)
	if err != nil {
		return errors.Wrap(err, "list customer instances")
	}

	template := print.CustomerInstancesTmplLite
	if r.args.output == "wide" {
		template = print.CustomerInstancesTmplWide
	}

	return r.Print(template, instances)
}

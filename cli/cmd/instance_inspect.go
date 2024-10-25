package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/cli/print"
)

func (r *runners) InitInstanceInspectCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "inspect",
		Short:         "Show full details for a customer instance",
		Long:          `Show full details for a customer instance`,
		RunE:          r.inspectInstance,
		SilenceUsage:  false,
		SilenceErrors: true, // this command uses custom error printing
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.instanceInspectCustomer, "customer", "", "Customer Name or ID")
	cmd.Flags().StringVar(&r.args.instanceInspectInstance, "instance", "", "Instance Name or ID")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) inspectInstance(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()

	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.args.instanceInspectCustomer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	if r.args.instanceInspectInstance == "" {
		return errors.Errorf("missing or invalid parameters: instance")
	}

	customer, err := r.api.GetCustomerByNameOrId(r.appType, r.appID, r.args.instanceInspectCustomer)
	if err != nil {
		return errors.Wrapf(err, "get customer %q", r.args.instanceInspectCustomer)
	}

	instance, err := findInstanceByNameOrID(r.args.instanceInspectInstance, customer.Instances)
	if err != nil {
		return errors.Wrap(err, "find instance")
	}

	if err = print.Instance(r.outputFormat, r.w, instance); err != nil {
		return errors.Wrap(err, "print instance")
	}

	return nil
}

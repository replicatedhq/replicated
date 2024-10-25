package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitInstanceLSCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ls",
		Short:         "list customer instances",
		Long:          `list customer instances`,
		RunE:          r.listInstances,
		SilenceUsage:  false,
		SilenceErrors: true, // this command uses custom error printing
	}

	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.instanceListCustomer, "customer", "", "Customer Name or ID")
	cmd.Flags().StringArrayVar(&r.args.instanceListTags, "tag", []string{}, "Tags to use to filter instances (key=value format, can be specified multiple times). Only one tag needs to match (an OR operation)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listInstances(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()

	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.args.instanceListCustomer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	customer, err := r.api.GetCustomerByNameOrId(r.appType, r.appID, r.args.instanceListCustomer)
	if err != nil {
		return errors.Wrapf(err, "get customer %q", r.args.instanceListCustomer)
	}

	instances := customer.Instances
	if len(r.args.instanceListTags) > 0 {
		tags, err := parseTags(r.args.instanceListTags)
		if err != nil {
			return errors.Wrap(err, "parse tags")
		}
		instances = findInstancesByTags(tags, customer.Instances)
	}

	if err := print.Instances(r.outputFormat, r.w, instances); err != nil {
		return errors.Wrap(err, "print instances")
	}

	return nil
}

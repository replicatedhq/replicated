package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersArchiveCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "archive",
		Short:        "archive a customer",
		Long:         `archive a customer`,
		RunE:         r.archiveCustomer,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerLicenseInspectCustomer, "customer", "", "The Customer Name or ID")

	return cmd
}

func (r *runners) archiveCustomer(_ *cobra.Command, _ []string) error {
	if r.args.customerLicenseInspectCustomer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	customer, err := r.api.GetCustomerByName(r.appType, r.appID, r.args.customerLicenseInspectCustomer)
	if err != nil {
		return errors.Wrapf(err, "find customer %q", r.args.customerLicenseInspectCustomer)
	}

	err = r.api.ArchiveCustomer(r.appType, customer.ID)
	if err != nil {
		return errors.Wrapf(err, "archive customer %q", customer.Name)
	}

	return nil
}

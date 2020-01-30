package cmd

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
)

func (r *runners) InitCustomersLicenseInspectCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "fetch a customer license",
		Long:  `fetch a customer license`,
		RunE:  r.inspectCustomerLicense,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerLicenseInspectCustomer, "customer", "", "The Customer Name or ID")
	cmd.Flags().StringVarP(&r.args.customerLicenseInspectOutput, "output", "o", "-", "Path to output license to. Defaults to stdout")

	return cmd
}

func (r *runners) inspectCustomerLicense(_ *cobra.Command, _ []string) error {
	if r.args.customerLicenseInspectCustomer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	if r.args.customerLicenseInspectOutput == "" {
		return errors.Errorf("missing or invalid parameters: output")
	}

	customer, err := r.api.FindCustomerByNameOrID(r.appID, r.appType, r.args.customerLicenseInspectCustomer)
	if err != nil {
		return errors.Wrap(err, "find customer")
	}

	err = print.Customers(r.stderrWriter, []types.Customer{*customer})
	if err != nil {
		return errors.Wrap(err, "print customer detail")
	}

	license, err := r.api.FetchLicense(r.appType, r.appSlug, customer.ID)
	if err != nil {
		return errors.Wrap(err, "fetch license")
	}

	if r.args.customerLicenseInspectOutput == "-" {
		if _, err := io.Copy(r.stdout, bytes.NewReader(license)); err != nil {
			return errors.Wrap(err, "write license to stdout")
		}
		return nil
	}

	if err := ioutil.WriteFile(r.args.customerLicenseInspectOutput, license, 0644); err != nil {
		return errors.Wrapf(err, "write license to file %q", r.args.customerLicenseInspectOutput)
	}

	return nil
}

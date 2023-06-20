package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersDownloadLicenseCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "download-license",
		Short:        "download a customer license",
		Long:         `download a customer license`,
		RunE:         r.downloadCustomerLicense,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerLicenseInspectCustomer, "customer", "", "The Customer Name or ID")
	cmd.Flags().StringVarP(&r.args.customerLicenseInspectOutput, "output", "o", "-", "Path to output license to. Defaults to stdout")

	return cmd
}

func (r *runners) downloadCustomerLicense(_ *cobra.Command, _ []string) error {
	if r.args.customerLicenseInspectCustomer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}
	if r.args.customerLicenseInspectOutput == "" {
		return errors.Errorf("missing or invalid parameters: output")
	}

	customer, err := r.api.GetCustomerByName(r.appType, r.appID, r.args.customerLicenseInspectCustomer)
	if err != nil {
		return errors.Wrapf(err, "find customer %q", r.args.customerLicenseInspectCustomer)
	}

	license, err := r.api.DownloadLicense(r.appType, r.appID, customer.ID)
	if err != nil {
		return errors.Wrapf(err, "download license for customer %q", customer.Name)
	}

	defer r.w.Flush()
	if r.args.customerLicenseInspectOutput == "-" {
		_, err = fmt.Fprintln(r.w, string(license))
		return err
	}

	return ioutil.WriteFile(r.args.customerLicenseInspectOutput, license, 0644)
}

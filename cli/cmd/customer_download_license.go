package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersDownloadLicenseCommand(parent *cobra.Command) *cobra.Command {
	var (
		customer string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "download-license [flags]",
		Short: "Download a customer's license",
		Long: `The download-license command allows you to retrieve and save a customer's license.

This command fetches the license for a specified customer and either outputs it
to stdout or saves it to a file. The license contains crucial information about
the customer's subscription and usage rights.

You must specify the customer using either their name or ID with the --customer flag.`,
		Example: `  # Download license for a customer by ID and output to stdout
  replicated customer download-license --customer cus_abcdef123456

  # Download license for a customer by name and save to a file
  replicated customer download-license --customer "Acme Inc" --output license.yaml

  # Download license for a customer in a specific app (if you have multiple apps)
  replicated customer download-license --app myapp --customer "Acme Inc" --output license.yaml`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return r.downloadCustomerLicense(cmd, customer, output)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&customer, "customer", "", "The Customer Name or ID")
	cmd.Flags().StringVarP(&output, "output", "o", "-", "Path to output license to. Defaults to stdout")
	cmd.MarkFlagRequired("customer")

	return cmd
}

func (r *runners) downloadCustomerLicense(cmd *cobra.Command, customer string, output string) error {
	customerNameOrId, err := cmd.Flags().GetString("customer")
	if err != nil {
		return errors.Wrap(err, "get customer flag")
	}
	if customerNameOrId == "" {
		return errors.New("missing or invalid parameters: customer")
	}

	c, err := r.api.GetCustomerByNameOrId(r.appType, r.appID, customer)
	if err != nil {
		return errors.Wrapf(err, "find customer %q", customerNameOrId)
	}

	license, err := r.api.DownloadLicense(r.appType, r.appID, c.ID)
	if err != nil {
		return errors.Wrapf(err, "download license for customer %q", c.Name)
	}

	defer r.w.Flush()
	if output == "-" {
		_, err = fmt.Fprintln(r.w, string(license))
		return err
	}

	return ioutil.WriteFile(output, license, 0644)
}

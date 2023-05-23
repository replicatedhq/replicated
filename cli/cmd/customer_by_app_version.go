package cmd

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersByAppVersionCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "search-by-app-version VERSION --app APPID",
		Short:        "search-by-app-version VERSION --app APPID",
		Long:         `Search for customers by app version`,
		RunE:         r.ListCustomersByAppVersion,
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.appVersion, "AppVersion", "", "Version of the application")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) ListCustomersByAppVersion(_ *cobra.Command, args []string) error {

	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	if len(args) != 1 {
		return errors.New("This command requires a VERSION and APPID")
	}
	appVersion := args[0]

	customers, err := kotsRestClient.ListCustomersByAppVersion(
		r.appID,
		appVersion,
		r.appType,
	)
	if err != nil {
		return errors.Wrap(err, "search-by-app-version")
	}
	// sort by customer name to ensure they are grouped together
	sort.SliceStable(customers, func(i, j int) bool { return customers[i].Name < customers[j].Name })

	return print.CustomersWithInstances(r.outputFormat, r.w, customers)
}

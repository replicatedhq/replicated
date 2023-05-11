package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

// replicated customer search-by-app-version APPID VERSION
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
	//cmd.Flags().StringVar(&r.args.appID, "AppId", "", "ID of the application")
	cmd.Flags().StringVar(&r.args.appVersion, "AppVersion", "", "Version of the application")

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

	return print.Customers(r.outputFormat, r.w, customers)
}

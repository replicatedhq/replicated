package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersLSCommand(parent *cobra.Command) *cobra.Command {
	var (
		appVersion   string
		includeTest  bool
		outputFormat string
	)

	customersLsCmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List customers for the current application",
		Long: `List customers associated with the current application.

This command displays information about customers linked to your application.
By default, it shows all non-test customers. You can use flags to:
- Filter customers by a specific app version
- Include test customers in the results
- Change the output format (table or JSON)

The command requires an app to be set using the --app flag.`,
		Example: `# List all customers for the current application
replicated customer ls --app myapp
# Output results in JSON format
replicated customer ls --app myapp --output json

# Combine multiple flags
replicated customer ls --app myapp --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.listCustomers(appVersion, includeTest, outputFormat)
		},
	}

	parent.AddCommand(customersLsCmd)
	customersLsCmd.Flags().StringVar(&appVersion, "app-version", "", "Filter customers by a specific app version")
	customersLsCmd.Flags().BoolVar(&includeTest, "include-test", false, "Include test customers in the results")
	customersLsCmd.Flags().StringVar(&outputFormat, "output", "table", "Output format: json|table (default: table)")

	return customersLsCmd
}

func (r *runners) listCustomers(appVersion string, includeTest bool, outputFormat string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if appVersion == "" {
		customers, err := r.api.ListCustomers(r.appID, r.appType, includeTest)
		if err != nil {
			return errors.Wrap(err, "list customers")
		}
		return print.Customers(outputFormat, r.w, customers)
	} else {
		customers, err := r.api.ListCustomersByAppAndVersion(r.appID, appVersion, r.appType)
		if err != nil && outputFormat == "json" {
			return print.CustomersWithInstances(outputFormat, r.w, customers)
		} else if err != nil {
			return errors.Wrap(err, "list customers by app and app version")
		}
		return print.CustomersWithInstances(outputFormat, r.w, customers)
	}
}

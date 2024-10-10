package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersCommand(parent *cobra.Command) *cobra.Command {
	customersCmd := &cobra.Command{
		Use:   "customer",
		Short: "Manage customer accounts and licenses",
		Long: `The customer command allows vendors to manage end customer records and licenses.

This command provides a suite of subcommands for creating, updating, listing, and
managing customer accounts. You can perform operations such as creating new customers,
updating existing customer information, managing licenses, and viewing customer details.

Use the various subcommands to:
- Create new customer accounts
- Update existing customer information
- List all customers or view details of a specific customer
- Manage customer licenses
- Archive or unarchive customer accounts`,
		Example: `  # List all customers
  replicated customer ls

  # Create a new customer
  replicated customer create --name "Acme Inc" --channel stable

  # Update an existing customer
  replicated customer update --customer cus_abcdef123456 --name "New Company Name"

  # View details of a specific customer
  replicated customer inspect --customer cus_abcdef123456

  # Archive a customer
  replicated customer archive --customer cus_abcdef123456

  # Download a customer's license
  replicated customer download-license --customer cus_abcdef123456`,
	}
	parent.AddCommand(customersCmd)

	return customersCmd
}

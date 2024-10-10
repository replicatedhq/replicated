package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersArchiveCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive [flags]",
		Short: "Archive a customer",
		Long: `The archive command allows you to archive a specific customer.

This command will mark a customer as archived in the system. Archived customers
are typically those who are no longer active or have terminated their service.
Archiving a customer does not delete their data but changes their status in the system.

You must specify the customer using either their name or ID with the --customer flag.`,
		Example: `  # Archive a customer by ID
  replicated customer archive --customer cus_abcdef123456

  # Archive a customer by name
  replicated customer archive --customer "Acme Inc"

  # Archive a customer for a specific app (if you have multiple apps)
  replicated customer archive --app myapp --customer "Acme Inc"`,
		RunE:         r.archiveCustomer,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().String("customer", "", "The Customer Name or ID")
	cmd.MarkFlagRequired("customer")

	return cmd
}

func (r *runners) archiveCustomer(cmd *cobra.Command, _ []string) error {
	customerNameOrId, err := cmd.Flags().GetString("customer")
	if err != nil {
		return errors.Wrap(err, "get customer flag")
	}
	if customerNameOrId == "" {
		return errors.New("missing or invalid parameters: customer")
	}

	customer, err := r.api.GetCustomerByNameOrId(r.appType, r.appID, customerNameOrId)
	if err != nil {
		return errors.Wrapf(err, "find customer %q", customerNameOrId)
	}

	if err := r.api.ArchiveCustomer(r.appType, customer.ID); err != nil {
		return errors.Wrapf(err, "archive customer %q", customerNameOrId)
	}

	return nil
}

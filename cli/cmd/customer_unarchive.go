package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersUnarchiveCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive <customer_name_or_id>",
		Short: "Unarchive a customer",
		Long: `Unarchive a customer for the current application.

This command restores a previously archived customer, making them
visible in active customer lists again.

The customer can be specified by either their name or ID. When a name
matches more than one archived customer, all of them are unarchived.
Active customers with the same name are not affected.`,
		Example: `# Unarchive a customer by name
replicated customer unarchive "Acme Inc"

# Unarchive a customer by ID
replicated customer unarchive cus_abcdef123456

# Unarchive multiple customers by ID
replicated customer unarchive cus_abcdef123456 cus_xyz9876543210

# Unarchive a customer in a specific app (if you have multiple apps)
replicated customer unarchive --app myapp "Acme Inc"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.unarchiveCustomer(cmd, args)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) unarchiveCustomer(cmd *cobra.Command, customers []string) error {
	if len(customers) == 0 {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	if r.hasApp() && r.appType == "platform" {
		return errors.New("unarchiving customers is not supported for platform applications")
	}

	for _, customer := range customers {
		// an archived customer can share its name with other archived customers; unarchive all of them
		matches, err := r.resolveCustomers(customer, r.api.GetArchivedCustomersByName)
		if err != nil {
			return err
		}

		for _, c := range matches {
			if err := r.api.UnarchiveCustomer(c.ID); err != nil {
				return errors.Wrapf(err, "unarchive customer %q (%s)", c.Name, c.ID)
			}
			fmt.Fprintf(r.w, "Unarchived customer %s (%s)\n", c.Name, c.ID)
		}
	}
	r.w.Flush()

	return nil
}

package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersUnarchiveCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive <customer_name_or_id>",
		Short: "Unarchive a customer",
		Long: `Unarchive a customer for the current application.

This command restores a previously archived customer, making them
visible in active customer lists again.

The customer can be specified by either their name or ID.`,
		Example: `# Unarchive a customer by name
replicated customer unarchive "Acme Inc"

# Unarchive a customer by ID
replicated customer unarchive cus_abcdef123456

# Unarchive multiple customers by ID
replicated customer unarchive cus_abcdef123456 cus_xyz9876543210

# Unarchive a customer in a specific app (if you have multiple apps)
replicated customer unarchive --app myapp "Acme Inc"`,
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
		var c *types.Customer

		// try to get the customer as if we have an id first
		cc, err := r.api.GetCustomerByID(customer)
		if err != nil && errors.Cause(err) != platformclient.ErrNotFound {
			return errors.Wrapf(err, "find customer %q", customer)
		}
		if cc != nil {
			c = cc
		}

		if c == nil {
			if !r.hasApp() {
				return errors.New("no app specified: app is required when looking up customers by name")
			}
			// try to get the customer as if we have a name, including archived customers
			cc, err := r.api.GetCustomerByNameIncludeArchived(r.appID, customer)
			if err != nil {
				return errors.Wrapf(err, "find customer %q", customer)
			}
			if cc != nil {
				c = cc
			}
		}

		if c == nil {
			return errors.Errorf("customer %q not found", customer)
		}

		err = r.api.UnarchiveCustomer(c.ID)
		if err != nil {
			return errors.Wrapf(err, "unarchive customer %q", c.Name)
		}
	}

	return nil
}

package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersArchiveCommand(parent *cobra.Command) *cobra.Command {
	var (
		customer string
		app      string
	)

	cmd := &cobra.Command{
		Use:   "archive <customer_name_or_id>",
		Short: "Archive a customer",
		Long: `Archive a customer for the current application.

This command allows you to archive a customer record. Archiving a customer
will make their license inactive and remove them from active customer lists.
This action is reversible - you can unarchive a customer later if needed.

The customer can be specified by either their name or ID.`,
		Example: `# Archive a customer by name
replicated customer archive "Acme Inc"

# Archive a customer by ID
replicated customer archive cus_abcdef123456

# Archive multiple customers by ID
replicated customer archive cus_abcdef123456 cus_xyz9876543210

# Archive a customer in a specific app (if you have multiple apps)
replicated customer archive --app myapp "Acme Inc"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// for compatibility reasons, we want to continue to support --customer but also read from args[0] if that's set
			customers := []string{}
			if customer != "" {
				customers = append(customers, customer)
			} else {
				customers = args
			}

			return r.archiveCustomer(cmd, customers, app)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&customer, "customer", "", "The Customer Name or ID to archive")
	cmd.Flags().MarkHidden("customer")
	cmd.Flags().StringVar(&app, "app", "", "The app to archive the customer in (not required when using a customer id)")

	return cmd
}
func (r *runners) archiveCustomer(cmd *cobra.Command, customers []string, app string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if len(customers) == 0 {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	if r.appType == "platform" {
		return errors.New("archiving customers is not supported for platform applications")
	}

	for _, customer := range customers {
		var c *types.Customer

		// try to get the customer as if we have an id first
		cc, err := r.api.GetCustomerByID(customer)
		if err != nil && err != kotsclient.ErrNotFound {
			return errors.Wrapf(err, "find customer %q", customer)
		}
		if cc != nil {
			c = cc
		}

		if c == nil {
			// try to get the customer as if we have a name
			cc, err := r.api.GetCustomerByName(app, customer)
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

		err = r.api.ArchiveCustomer(c.ID)
		if err != nil {
			return errors.Wrapf(err, "archive customer %q", c.Name)
		}
	}

	return nil
}

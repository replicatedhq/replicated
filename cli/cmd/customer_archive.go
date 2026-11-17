package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersArchiveCommand(parent *cobra.Command) *cobra.Command {
	var customer string

	cmd := &cobra.Command{
		Use:   "archive <customer_name_or_id>",
		Short: "Archive a customer",
		Long: `Archive a customer for the current application.

This command allows you to archive a customer record. Archiving a customer
removes them from active customer lists. It is not a way to cut off the
customer's access. Some requests that use an archived customer's license are
rejected, such as logging in to the Replicated registry, but not all of them.
To stop a customer from installing, updating, or pulling images, set an
expiration date first with 'replicated customer update --expires-at'.

This action is reversible. You can unarchive a customer later if needed.

The customer can be specified by either their name or ID.`,
		Example: `# Archive a customer by name
replicated customer archive "Acme Inc"

# Archive a customer by ID
replicated customer archive cus_abcdef123456

# Archive multiple customers by ID
replicated customer archive cus_abcdef123456 cus_xyz9876543210

# Archive a customer in a specific app (if you have multiple apps)
replicated customer archive --app myapp "Acme Inc"`,
		Args: func(cmd *cobra.Command, args []string) error {
			// the hidden --customer flag can be used in place of args
			if customer != "" {
				return nil
			}
			return cobra.MinimumNArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// for compatibility reasons, we want to continue to support --customer but also read from args[0] if that's set
			customers := []string{}
			if customer != "" {
				customers = append(customers, customer)
			} else {
				customers = args
			}

			return r.archiveCustomer(cmd, customers)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&customer, "customer", "", "The Customer Name or ID to archive")
	cmd.Flags().MarkHidden("customer")

	return cmd
}
func (r *runners) archiveCustomer(cmd *cobra.Command, customers []string) error {
	if len(customers) == 0 {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	if r.hasApp() && r.appType == "platform" {
		return errors.New("archiving customers is not supported for platform applications")
	}

	for _, customer := range customers {
		c, err := r.resolveCustomer(customer, false)
		if err != nil {
			return err
		}

		if err := r.api.ArchiveCustomer(c.ID); err != nil {
			return errors.Wrapf(err, "archive customer %q", c.Name)
		}
	}

	return nil
}

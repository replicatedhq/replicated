package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalInviteCmd(parent *cobra.Command) *cobra.Command {
	var customer string

	cmd := &cobra.Command{
		Use:   "invite [EMAIL_ADDRESSES...]",
		Short: "Invite users to the Enterprise Portal",
		Long: `Invite one or more users to access the Enterprise Portal for a specific customer.

This command allows you to send invitation emails to users, granting them access
to the Enterprise Portal associated with a particular customer. You need to specify
the customer (by name or ID) and provide one or more email addresses of the users
you want to invite.

The command will generate and return unique invitation URLs for each email address.`,
		Example: `# Invite a single user to the Enterprise Portal for a customer
replicated enterprise-portal invite --customer "ACME Inc" user@example.com

# Invite multiple users to the Enterprise Portal for a customer
replicated enterprise-portal invite --customer "cus_abcdef123456" user1@example.com user2@example.com

# Invite a user and specify JSON output format
replicated enterprise-portal invite --customer "ACME Inc" --output json user@example.com

# Invite users to the Enterprise Portal for a specific app (if you have multiple apps)
replicated enterprise-portal invite --app myapp --customer "ACME Inc" user1@example.com user2@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.enterprisePortalInvite(cmd, r.appID, customer, args)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&customer, "customer", "", "The customer name or ID to invite")

	cmd.MarkFlagRequired("customer")

	return cmd
}

func (r *runners) enterprisePortalInvite(cmd *cobra.Command, appID string, customer string, emailAddresses []string) error {
	if appID == "" {
		return errors.Errorf("app required")
	}

	c, err := r.resolveCustomer(customer)
	if err != nil {
		return err
	}

	for _, emailAddress := range emailAddresses {
		url, err := r.kotsAPI.SendEnterprisePortalInvite(appID, c.ID, emailAddress)
		if err != nil {
			return err
		}

		fmt.Printf("%s: %s\n", emailAddress, url)
	}
	return nil
}

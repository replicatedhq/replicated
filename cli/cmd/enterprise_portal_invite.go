package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePortalInviteCmd(parent *cobra.Command) *cobra.Command {
	var customer string
	var outputFormat string

	cmd := &cobra.Command{
		Use: "invite",

		RunE: func(cmd *cobra.Command, args []string) error {
			return r.enterprisePortalInvite(cmd, r.appID, customer, args, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&customer, "customer", "", "The customer name or ID to invite")
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) enterprisePortalInvite(cmd *cobra.Command, appID string, customer string, emailAddresses []string, outputFormat string) error {
	var c *types.Customer

	// try to get the customer as if we have an id first
	cc, err := r.api.GetCustomerByID(customer)
	if err != nil && err != kotsclient.ErrNotFound {
		return errors.Wrapf(err, "find customer %q", customer)
	}
	if cc != nil {
		c = cc
	}

	if appID == "" {
		return errors.Errorf("app required")
	}

	if c == nil {
		// try to get the customer as if we have a name
		cc, err := r.api.GetCustomerByName(r.appID, customer)
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

	for _, emailAddress := range emailAddresses {
		err := r.kotsAPI.SendEnterprisePortalInvite(appID, c.ID, emailAddress)
		if err != nil {
			return err
		}

		fmt.Printf("Sent invitation to %s\n", emailAddress)
	}
	return nil
}

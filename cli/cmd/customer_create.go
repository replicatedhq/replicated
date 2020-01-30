package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersCreateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a customer",
		Long:  `create a customer`,
		RunE:  r.createCustomer,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerCreateName, "name", "", "The Customer Name")
	cmd.Flags().StringVar(&r.args.customerCreateChannel, "channel", "", "The Customer Channel")
	cmd.Flags().DurationVar(&r.args.customerCreateExpiryDuration, "expires-in", 0, "If set, an expiration date will be set on the license")
	cmd.Flags().BoolVar(&r.args.customerCreateEnsureChannel, "ensure-channel", false, "If set, channel will be created if it does not exist.")

	return cmd
}

func (r *runners) createCustomer(_ *cobra.Command, _ []string) error {

	channel, err := r.api.GetChannelByName(
		r.appID,
		r.appType,
		r.args.customerCreateChannel,
		"",
		r.args.customerCreateEnsureChannel,
	)
	if err != nil {
		return errors.Wrap(err, "get channel")
	}

	customer, err := r.api.CreateCustomer(r.appType, r.args.customerCreateName, channel.ID, r.args.customerCreateExpiryDuration)
	if err != nil {
		return errors.Wrap(err, "create customer")
	}

	// CreateCustomer query doesn't pull back the Channels, so bolt it on afterward for the printing
	customer.Channels = append(customer.Channels, *channel)

	return print.Customers(r.stdoutWriter, []types.Customer{*customer})
}

package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersCreateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create a customer",
		Long:         `create a customer`,
		RunE:         r.createCustomer,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerCreateName, "name", "", "Name of the customer")
	cmd.Flags().StringVar(&r.args.customerCreateChannel, "channel", "", "Release channel to which the customer should be assigned")
	cmd.Flags().DurationVar(&r.args.customerCreateExpiryDuration, "expires-in", 0, "If set, an expiration date will be set on the license. Supports Go durations like '72h' or '3600m'")
	cmd.Flags().BoolVar(&r.args.customerCreateEnsureChannel, "ensure-channel", false, "If set, channel will be created if it does not exist.")

	return cmd
}

func (r *runners) createCustomer(_ *cobra.Command, _ []string) error {

	channel, err := r.api.GetOrCreateChannelByName(
		r.appID,
		r.appType,
		r.appSlug,
		r.args.customerCreateChannel,
		"",
		r.args.customerCreateEnsureChannel,
	)
	if err != nil {
		return errors.Wrap(err, "get channel")
	}

	customer, err := r.api.CreateCustomer(r.appID, r.appType, r.args.customerCreateName, channel.ID, r.args.customerCreateExpiryDuration)
	if err != nil {
		return errors.Wrap(err, "create customer")
	}

	return print.Customers(r.w, []types.Customer{*customer})
}

package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
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
	cmd.Flags().BoolVar(&r.args.customerCreateIsAirgapEnabled, "airgap", false, "If set, the license will allow airgap installs.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsGitopsSupported, "gitops", false, "If set, the license will allow the GitOps usage.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsSnapshotSupported, "snapshot", false, "If set, the license will allow Snapshots.")
	cmd.Flags().StringVar(&r.args.customerCreateEmail, "email", "", "Email address of the customer that is to be created.")
	return cmd
}

func (r *runners) createCustomer(_ *cobra.Command, _ []string) error {

	channel, err := r.api.GetOrCreateChannelByName(
		r.appID,
		r.appType,
		r.args.customerCreateChannel,
		"",
		r.args.customerCreateEnsureChannel,
	)
	if err != nil {
		return errors.Wrap(err, "get channel")
	}

	opts := kotsclient.CreateCustomerOpts{
		Name:                r.args.customerCreateName,
		ChannelID:           channel.ID,
		AppID:               r.appID,
		ExpiresAt:           r.args.customerCreateExpiryDuration,
		IsAirgapEnabled:     r.args.customerCreateIsAirgapEnabled,
		IsGitopsSupported:   r.args.customerCreateIsGitopsSupported,
		IsSnapshotSupported: r.args.customerCreateIsSnapshotSupported,
		LicenseType:         "dev",
		Email:               r.args.customerCreateEmail,
	}

	customer, err := r.api.CreateCustomer(r.appType, opts)
	if err != nil {
		return errors.Wrap(err, "create customer")
	}

	return print.Customers(r.w, []types.Customer{*customer})
}

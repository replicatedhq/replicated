package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomerUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "update",
		Short:         "update a customer",
		Long:          `update a customer`,
		RunE:          r.updateCustomer,
		SilenceUsage:  false,
		SilenceErrors: true, // this command uses custom error printing
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerUpdateID, "customer", "", "The ID of the customer to update")
	_ = cmd.MarkFlagRequired("customer")
	cmd.Flags().StringVar(&r.args.customerUpdateName, "name", "", "Name of the customer")
	cmd.Flags().StringVar(&r.args.customerUpdateCustomID, "custom-id", "", "Set a custom customer ID to more easily tie this customer record to your external data systems")
	cmd.Flags().StringVar(&r.args.customerUpdateChannel, "channel", "", "Release channel to which the customer should be assigned")
	cmd.Flags().DurationVar(&r.args.customerUpdateExpiryDuration, "expires-in", 0, "If set, an expiration date will be set on the license. Supports Go durations like '72h' or '3600m'")
	cmd.Flags().BoolVar(&r.args.customerUpdateEnsureChannel, "ensure-channel", false, "If set, channel will be created if it does not exist.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsAirgapEnabled, "airgap", false, "If set, the license will allow airgap installs.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsGitopsSupported, "gitops", false, "If set, the license will allow the GitOps usage.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsSnapshotSupported, "snapshot", false, "If set, the license will allow Snapshots.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsKotInstallEnabled, "kots-install", true, "If set, the license will allow KOTS install. Otherwise license will allow Helm CLI installs only.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsEmbeddedClusterDownloadEnabled, "embedded-cluster-download", false, "If set, the license will allow embedded cluster downloads.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsGeoaxisSupported, "geo-axis", false, "If set, the license will allow Geo Axis usage.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsHelmVMDownloadEnabled, "helmvm-cluster-download", false, "If set, the license will allow helmvm cluster downloads.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsIdentityServiceSupported, "identity-service", false, "If set, the license will allow Identity Service usage.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsSupportBundleUploadEnabled, "support-bundle-upload", false, "If set, the license will allow uploading support bundles.")
	cmd.Flags().StringVar(&r.args.customerUpdateEmail, "email", "", "Email address of the customer that is to be updated.")
	cmd.Flags().StringVar(&r.args.customerUpdateType, "type", "dev", "The license type to update. One of: dev|trial|paid|community|test (default: dev)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	return cmd
}

func (r *runners) updateCustomer(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()
	if r.args.customerUpdateID == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	// all of the following validation occurs in the API also, but
	// we want to fail fast if the user has provided invalid input
	if err := validateCustomerType(r.args.customerUpdateType); err != nil {
		return errors.Wrap(err, "validate customer type")
	}
	if r.args.customerUpdateType == "test" && r.args.customerUpdateExpiryDuration > time.Hour*48 {
		return errors.New("test licenses cannot be updated with an expiration date greater than 48 hours")
	}
	if r.args.customerUpdateType == "paid" {
		r.args.customerUpdateType = "prod"
	}

	channelID := ""
	if r.args.customerUpdateChannel != "" {
		getOrCreateChannelOptions := client.GetOrCreateChannelOptions{
			AppID:          r.appID,
			AppType:        r.appType,
			NameOrID:       r.args.customerUpdateChannel,
			Description:    "",
			CreateIfAbsent: r.args.customerUpdateEnsureChannel,
		}

		channel, err := r.api.GetOrCreateChannelByName(getOrCreateChannelOptions)
		if err != nil {
			return errors.Wrap(err, "get channel")
		}

		channelID = channel.ID
	}

	opts := kotsclient.UpdateCustomerOpts{
		Name:                             r.args.customerUpdateName,
		CustomID:                         r.args.customerUpdateCustomID,
		ChannelID:                        channelID,
		AppID:                            r.appID,
		ExpiresAt:                        r.args.customerUpdateExpiryDuration,
		IsAirgapEnabled:                  r.args.customerUpdateIsAirgapEnabled,
		IsGitopsSupported:                r.args.customerUpdateIsGitopsSupported,
		IsSnapshotSupported:              r.args.customerUpdateIsSnapshotSupported,
		IsKotInstallEnabled:              r.args.customerUpdateIsKotInstallEnabled,
		IsEmbeddedClusterDownloadEnabled: r.args.customerUpdateIsEmbeddedClusterDownloadEnabled,
		IsGeoaxisSupported:               r.args.customerUpdateIsGeoaxisSupported,
		IsHelmVMDownloadEnabled:          r.args.customerUpdateIsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:       r.args.customerUpdateIsIdentityServiceSupported,
		IsSupportBundleUploadEnabled:     r.args.customerUpdateIsSupportBundleUploadEnabled,
		LicenseType:                      r.args.customerUpdateType,
		Email:                            r.args.customerUpdateEmail,
	}

	customer, err := r.api.UpdateCustomer(r.appType, r.args.customerUpdateID, opts)
	if err != nil {
		return errors.Wrap(err, "update customer")
	}

	err = print.Customer(r.outputFormat, r.w, customer)
	if err != nil {
		return errors.Wrap(err, "print customer")
	}

	return nil
}

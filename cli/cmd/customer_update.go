package cmd

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomerUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "update --customer <id> [options]",
		Short:         "update a customer",
		Long:          `update a customer`,
		RunE:          r.updateCustomer,
		SilenceUsage:  false,
		SilenceErrors: true, // this command uses custom error printing
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerUpdateID, "customer", "", "The ID of the customer to update")
	cmd.Flags().StringVar(&r.args.customerUpdateName, "name", "", "Name of the customer")
	cmd.Flags().StringVar(&r.args.customerUpdateCustomID, "custom-id", "", "Set a custom customer ID to more easily tie this customer record to your external data systems")
	cmd.Flags().StringArrayVar(&r.args.customerUpdateChannel, "channel", []string{}, "Release channel to which the customer should be assigned (can be specified multiple times)")
	cmd.Flags().StringVar(&r.args.customerUpdateDefaultChannel, "default-channel", "", "Which of the specified channels should be the default channel. if not set, the first channel specified will be the default channel.")
	cmd.Flags().DurationVar(&r.args.customerUpdateExpiryDuration, "expires-in", 0, "If set, an expiration date will be set on the license. Supports Go durations like '72h' or '3600m'")
	cmd.Flags().BoolVar(&r.args.customerUpdateEnsureChannel, "ensure-channel", false, "If set, channel will be created if it does not exist.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsAirgapEnabled, "airgap", false, "If set, the license will allow airgap installs.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsGitopsSupported, "gitops", false, "If set, the license will allow the GitOps usage.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsSnapshotSupported, "snapshot", false, "If set, the license will allow Snapshots.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsKotsInstallEnabled, "kots-install", true, "If set, the license will allow KOTS install. Otherwise license will allow Helm CLI installs only.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsEmbeddedClusterDownloadEnabled, "embedded-cluster-download", false, "If set, the license will allow embedded cluster downloads.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsGeoaxisSupported, "geo-axis", false, "If set, the license will allow Geo Axis usage.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsHelmVMDownloadEnabled, "helmvm-cluster-download", false, "If set, the license will allow helmvm cluster downloads.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsIdentityServiceSupported, "identity-service", false, "If set, the license will allow Identity Service usage.")
	cmd.Flags().BoolVar(&r.args.customerUpdateIsSupportBundleUploadEnabled, "support-bundle-upload", false, "If set, the license will allow uploading support bundles.")
	cmd.Flags().StringVar(&r.args.customerUpdateEmail, "email", "", "Email address of the customer that is to be updated.")
	cmd.Flags().StringVar(&r.args.customerUpdateType, "type", "dev", "The license type to update. One of: dev|trial|paid|community|test (default: dev)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.MarkFlagRequired("customer")
	cmd.MarkFlagRequired("channel")
	cmd.MarkFlagRequired("name") // until the API supports better patching, this is actually a required field

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

	channels := []kotsclient.CustomerChannel{}

	foundDefaultChannel := false
	for _, requestedChannel := range r.args.customerUpdateChannel {
		getOrCreateChannelOptions := client.GetOrCreateChannelOptions{
			AppID:          r.appID,
			AppType:        r.appType,
			NameOrID:       requestedChannel,
			Description:    "",
			CreateIfAbsent: r.args.customerCreateEnsureChannel,
		}

		channel, err := r.api.GetOrCreateChannelByName(getOrCreateChannelOptions)
		if err != nil {
			return errors.Wrap(err, "get channel")
		}

		customerChannel := kotsclient.CustomerChannel{
			ID: channel.ID,
		}

		if r.args.customerCreateDefaultChannel == requestedChannel {
			customerChannel.IsDefault = true
		}

		channels = append(channels, customerChannel)
	}

	if len(channels) == 0 {
		return errors.New("no channels found")
	}

	if !foundDefaultChannel {
		log := logger.NewLogger(os.Stdout)
		log.Info("No default channel specified, defaulting to the first channel specified.")
		firstChannel := channels[0]
		firstChannel.IsDefault = true
		channels[0] = firstChannel
	}

	opts := kotsclient.UpdateCustomerOpts{
		Name:                             r.args.customerUpdateName,
		CustomID:                         r.args.customerUpdateCustomID,
		Channels:                         channels,
		AppID:                            r.appID,
		ExpiresAtDuration:                r.args.customerUpdateExpiryDuration,
		IsAirgapEnabled:                  r.args.customerUpdateIsAirgapEnabled,
		IsGitopsSupported:                r.args.customerUpdateIsGitopsSupported,
		IsSnapshotSupported:              r.args.customerUpdateIsSnapshotSupported,
		IsKotsInstallEnabled:             r.args.customerUpdateIsKotsInstallEnabled,
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

package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersCreateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "create",
		Short:         "create a customer",
		Long:          `create a customer`,
		RunE:          r.createCustomer,
		SilenceUsage:  false,
		SilenceErrors: true, // this command uses custom error printing
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerCreateName, "name", "", "Name of the customer")
	cmd.Flags().StringVar(&r.args.customerCreateCustomID, "custom-id", "", "Set a custom customer ID to more easily tie this customer record to your external data systems")
	cmd.Flags().StringVar(&r.args.customerCreateChannel, "channel", "", "Release channel to which the customer should be assigned")
	cmd.Flags().DurationVar(&r.args.customerCreateExpiryDuration, "expires-in", 0, "If set, an expiration date will be set on the license. Supports Go durations like '72h' or '3600m'")
	cmd.Flags().BoolVar(&r.args.customerCreateEnsureChannel, "ensure-channel", false, "If set, channel will be created if it does not exist.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsAirgapEnabled, "airgap", false, "If set, the license will allow airgap installs.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsGitopsSupported, "gitops", false, "If set, the license will allow the GitOps usage.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsSnapshotSupported, "snapshot", false, "If set, the license will allow Snapshots.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsKotsInstallEnabled, "kots-install", true, "If set, the license will allow KOTS install. Otherwise license will allow Helm CLI installs only.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsEmbeddedClusterDownloadEnabled, "embedded-cluster-download", false, "If set, the license will allow embedded cluster downloads.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsGeoaxisSupported, "geo-axis", false, "If set, the license will allow Geo Axis usage.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsHelmVMDownloadEnabled, "helmvm-cluster-download", false, "If set, the license will allow helmvm cluster downloads.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsIdentityServiceSupported, "identity-service", false, "If set, the license will allow Identity Service usage.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsInstallerSupportEnabled, "installer-support", false, "If set, the license will allow installer support.")
	cmd.Flags().BoolVar(&r.args.customerCreateIsSupportBundleUploadEnabled, "support-bundle-upload", false, "If set, the license will allow uploading support bundles.")
	cmd.Flags().StringVar(&r.args.customerCreateEmail, "email", "", "Email address of the customer that is to be created.")
	cmd.Flags().StringVar(&r.args.customerCreateType, "type", "dev", "The license type to create. One of: dev|trial|paid|community|test (default: dev)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	return cmd
}

func (r *runners) createCustomer(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()

	// all of the following validation occurs in the API also, but
	// we want to fail fast if the user has provided invalid input
	if err := validateCustomerType(r.args.customerCreateType); err != nil {
		return errors.Wrap(err, "validate customer type")
	}
	if r.args.customerCreateType == "test" && r.args.customerCreateExpiryDuration > time.Hour*48 {
		return errors.New("test licenses cannot be created with an expiration date greater than 48 hours")
	}
	if r.args.customerCreateType == "paid" {
		r.args.customerCreateType = "prod"
	}

	channelID := ""
	if r.args.customerCreateChannel != "" {
		getOrCreateChannelOptions := client.GetOrCreateChannelOptions{
			AppID:          r.appID,
			AppType:        r.appType,
			NameOrID:       r.args.customerCreateChannel,
			Description:    "",
			CreateIfAbsent: r.args.customerCreateEnsureChannel,
		}

		channel, err := r.api.GetOrCreateChannelByName(getOrCreateChannelOptions)
		if err != nil {
			return errors.Wrap(err, "get channel")
		}

		channelID = channel.ID
	}

	opts := kotsclient.CreateCustomerOpts{
		Name:                             r.args.customerCreateName,
		CustomID:                         r.args.customerCreateCustomID,
		ChannelID:                        channelID,
		AppID:                            r.appID,
		ExpiresAtDuration:                r.args.customerCreateExpiryDuration,
		IsAirgapEnabled:                  r.args.customerCreateIsAirgapEnabled,
		IsGitopsSupported:                r.args.customerCreateIsGitopsSupported,
		IsSnapshotSupported:              r.args.customerCreateIsSnapshotSupported,
		IsKotsInstallEnabled:             r.args.customerCreateIsKotsInstallEnabled,
		IsEmbeddedClusterDownloadEnabled: r.args.customerCreateIsEmbeddedClusterDownloadEnabled,
		IsGeoaxisSupported:               r.args.customerCreateIsGeoaxisSupported,
		IsHelmVMDownloadEnabled:          r.args.customerCreateIsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:       r.args.customerCreateIsIdentityServiceSupported,
		IsInstallerSupportEnabled:        r.args.customerCreateIsInstallerSupportEnabled,
		IsSupportBundleUploadEnabled:     r.args.customerCreateIsSupportBundleUploadEnabled,
		LicenseType:                      r.args.customerCreateType,
		Email:                            r.args.customerCreateEmail,
	}

	customer, err := r.api.CreateCustomer(r.appType, opts)
	if err != nil {
		return errors.Wrap(err, "create customer")
	}

	err = print.Customer(r.outputFormat, r.w, customer)
	if err != nil {
		return errors.Wrap(err, "print customer")
	}

	return nil
}

func validateCustomerType(customerType string) error {
	switch customerType {
	case "dev", "trial", "paid", "community", "test":
		return nil
	default:
		return errors.Errorf("invalid customer type: %s", customerType)
	}
}

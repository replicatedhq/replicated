package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

type updateCustomerOpts struct {
	CustomerID                   string
	Name                         string
	CustomID                     string
	Channels                     []string
	DefaultChannel               string
	ExpiryDuration               time.Duration
	EnsureChannel                bool
	IsAirgapEnabled              bool
	IsGitopsSupported            bool
	IsSnapshotSupported          bool
	IsKotsInstallEnabled         bool
	IsEmbeddedClusterEnabled     bool
	IsGeoaxisSupported           bool
	IsHelmVMDownloadEnabled      bool
	IsIdentityServiceSupported   bool
	IsSupportBundleUploadEnabled bool
	IsDeveloperModeEnabled       bool
	Email                        string
	Type                         string
}

func (r *runners) InitCustomerUpdateCommand(parent *cobra.Command) *cobra.Command {
	opts := updateCustomerOpts{}
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "update --customer <id> --name <name> [options]",
		Short: "Update an existing customer",
		Long: `Update an existing customer's information and settings.

	This command allows you to modify various attributes of a customer, including their name,
	custom ID, assigned channels, license type, and feature flags. You can update expiration dates,
	enable or disable specific features, and change channel assignments.

	The --customer flag is required to specify which customer to update.`,
		Example: `  # Update a customer's name
	  replicated customer update --customer cus_abcdef123456 --name "New Company Name"

	  # Change a customer's channel and make it the default
	  replicated customer update --customer cus_abcdef123456 --channel stable --default-channel stable

	  # Enable airgap installations for a customer
	  replicated customer update --customer cus_abcdef123456 --airgap

	  # Update multiple attributes at once
	  replicated customer update --customer cus_abcdef123456 --name "Updated Corp" --type paid --channel enterprise --airgap --snapshot

	  # Set an expiration date for a customer's license
	  replicated customer update --customer cus_abcdef123456 --expires-in 8760h

	  # Update a customer and output the result in JSON format
	  replicated customer update --customer cus_abcdef123456 --name "JSON Corp" --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.updateCustomer(cmd, opts, outputFormat)
		},
		SilenceUsage:  false,
		SilenceErrors: false,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&opts.CustomerID, "customer", "", "The ID of the customer to update")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the customer")
	cmd.Flags().StringVar(&opts.CustomID, "custom-id", "", "Set a custom customer ID to more easily tie this customer record to your external data systems")
	cmd.Flags().StringArrayVar(&opts.Channels, "channel", []string{}, "Release channel to which the customer should be assigned (can be specified multiple times)")
	cmd.Flags().StringVar(&opts.DefaultChannel, "default-channel", "", "Which of the specified channels should be the default channel. if not set, the first channel specified will be the default channel.")
	cmd.Flags().DurationVar(&opts.ExpiryDuration, "expires-in", 0, "If set, an expiration date will be set on the license. Supports Go durations like '72h' or '3600m'")
	cmd.Flags().BoolVar(&opts.EnsureChannel, "ensure-channel", false, "If set, channel will be created if it does not exist.")
	cmd.Flags().BoolVar(&opts.IsAirgapEnabled, "airgap", false, "If set, the license will allow airgap installs.")
	cmd.Flags().BoolVar(&opts.IsGitopsSupported, "gitops", false, "If set, the license will allow the GitOps usage.")
	cmd.Flags().BoolVar(&opts.IsSnapshotSupported, "snapshot", false, "If set, the license will allow Snapshots.")
	cmd.Flags().BoolVar(&opts.IsKotsInstallEnabled, "kots-install", true, "If set, the license will allow KOTS install. Otherwise license will allow Helm CLI installs only.")
	cmd.Flags().BoolVar(&opts.IsEmbeddedClusterEnabled, "embedded-cluster-download", false, "If set, the license will allow embedded cluster downloads.")
	cmd.Flags().BoolVar(&opts.IsGeoaxisSupported, "geo-axis", false, "If set, the license will allow Geo Axis usage.")
	cmd.Flags().BoolVar(&opts.IsHelmVMDownloadEnabled, "helmvm-cluster-download", false, "If set, the license will allow helmvm cluster downloads.")
	cmd.Flags().BoolVar(&opts.IsIdentityServiceSupported, "identity-service", false, "If set, the license will allow Identity Service usage.")
	cmd.Flags().BoolVar(&opts.IsSupportBundleUploadEnabled, "support-bundle-upload", false, "If set, the license will allow uploading support bundles.")
	cmd.Flags().BoolVar(&opts.IsDeveloperModeEnabled, "developer-mode", false, "If set, Replicated SDK installed in dev mode will use mock data.")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Email address of the customer that is to be updated.")
	cmd.Flags().StringVar(&opts.Type, "type", "dev", "The license type to update. One of: dev|trial|paid|community|test (default: dev)")
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.MarkFlagRequired("customer")
	cmd.MarkFlagRequired("channel")
	cmd.MarkFlagRequired("name") // until the API supports better patching, this is actually a required field

	return cmd
}

func (r *runners) updateCustomer(cmd *cobra.Command, opts updateCustomerOpts, outputFormat string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()

	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if opts.CustomerID == "" {
		return errors.New("missing or invalid parameters: customer")
	}

	// all of the following validation occurs in the API also, but
	// we want to fail fast if the user has provided invalid input
	if err := validateCustomerType(opts.Type); err != nil {
		return errors.Wrap(err, "validate customer type")
	}
	if opts.Type == "test" && opts.ExpiryDuration > time.Hour*48 {
		return errors.New("test licenses cannot be updated with an expiration date greater than 48 hours")
	}
	if opts.Type == "paid" {
		opts.Type = "prod"
	}

	channels := []kotsclient.CustomerChannel{}

	foundDefaultChannel := false
	for _, requestedChannel := range opts.Channels {
		getOrCreateChannelOptions := client.GetOrCreateChannelOptions{
			AppID:          r.appID,
			AppType:        r.appType,
			NameOrID:       requestedChannel,
			Description:    "",
			CreateIfAbsent: opts.EnsureChannel,
		}

		channel, err := r.api.GetOrCreateChannelByName(getOrCreateChannelOptions)
		if err != nil {
			return errors.Wrap(err, "get channel")
		}

		customerChannel := kotsclient.CustomerChannel{
			ID: channel.ID,
		}

		if opts.DefaultChannel == requestedChannel {
			customerChannel.IsDefault = true
			foundDefaultChannel = true
		}

		channels = append(channels, customerChannel)
	}

	if len(channels) == 0 {
		return errors.New("no channels found")
	}

	if opts.DefaultChannel != "" && !foundDefaultChannel {
		return errors.New("default channel not found in specified channels")
	}

	if !foundDefaultChannel {
		if len(channels) > 1 {
			fmt.Fprintln(os.Stderr, "No default channel specified, defaulting to the first channel specified.")
		}
		firstChannel := channels[0]
		firstChannel.IsDefault = true
		channels[0] = firstChannel
	}

	updateOpts := kotsclient.UpdateCustomerOpts{
		Name:                             opts.Name,
		CustomID:                         opts.CustomID,
		Channels:                         channels,
		AppID:                            r.appID,
		ExpiresAtDuration:                opts.ExpiryDuration,
		IsAirgapEnabled:                  opts.IsAirgapEnabled,
		IsGitopsSupported:                opts.IsGitopsSupported,
		IsSnapshotSupported:              opts.IsSnapshotSupported,
		IsKotsInstallEnabled:             opts.IsKotsInstallEnabled,
		IsEmbeddedClusterDownloadEnabled: opts.IsEmbeddedClusterEnabled,
		IsGeoaxisSupported:               opts.IsGeoaxisSupported,
		IsHelmVMDownloadEnabled:          opts.IsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:       opts.IsIdentityServiceSupported,
		IsSupportBundleUploadEnabled:     opts.IsSupportBundleUploadEnabled,
		IsDeveloperModeEnabled:           opts.IsDeveloperModeEnabled,
		LicenseType:                      opts.Type,
		Email:                            opts.Email,
	}

	customer, err := r.api.UpdateCustomer(r.appType, opts.CustomerID, updateOpts)
	if err != nil {
		return errors.Wrap(err, "update customer")
	}

	err = print.Customer(outputFormat, r.w, customer)
	if err != nil {
		return errors.Wrap(err, "print customer")
	}

	return nil
}

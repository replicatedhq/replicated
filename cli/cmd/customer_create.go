package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

type createCustomerOpts struct {
	Name                              string
	CustomID                          string
	ChannelName                       string
	ExpiryDuration                    time.Duration
	EnsureChannel                     bool
	IsAirgapEnabled                   bool
	IsGitopsSupported                 bool
	IsSnapshotSupported               bool
	IsKotsInstallEnabled              bool
	IsHelmInstallEnabled              bool
	IsKurlInstallEnabled              bool
	IsEmbeddedClusterDownloadEnabled  bool
	IsEmbeddedClusterMultinodeEnabled bool
	IsGeoaxisSupported                bool
	IsHelmVMDownloadEnabled           bool
	IsIdentityServiceSupported        bool
	IsInstallerSupportEnabled         bool
	IsSupportBundleUploadEnabled      bool
	IsDeveloperModeEnabled            bool
	Email                             string
	CustomerType                      string
}

func (r *runners) InitCustomersCreateCommand(parent *cobra.Command) *cobra.Command {
	opts := createCustomerOpts{}
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new customer for the current application",
		Long: `Create a new customer for the current application with specified attributes.

This command allows you to create a customer record with various properties such as name,
custom ID, channel, license type, and feature flags. You can set expiration dates,
enable or disable specific features, and assign the customer to a channel.

The --app flag must be set to specify the target application.`,
		Example: `# Create a basic customer with a name and assigned to a channel
replicated customer create --app myapp --name "Acme Inc" --channel stable

# Create a paid customer with specific features enabled
replicated customer create --app myapp --name "Enterprise Ltd" --type paid --channel enterprise --airgap --snapshot

# Create a trial customer with an expiration date
replicated customer create --app myapp --name "Trial User" --type trial --channel stable --expires-in 720h

# Create a customer with all available options
replicated customer create --app myapp --name "Full Options Inc" --custom-id "FULL001" \
	--channel stable --type paid --email "contact@fulloptions.com" --expires-in 8760h \
	--airgap --snapshot --kots-install --embedded-cluster-download \
	--support-bundle-upload --ensure-channel`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.createCustomer(cmd, opts, outputFormat)
		},
		SilenceUsage:  false,
		SilenceErrors: false,
	}

	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the customer")
	cmd.Flags().StringVar(&opts.CustomID, "custom-id", "", "Set a custom customer ID to more easily tie this customer record to your external data systems")
	cmd.Flags().StringVar(&opts.ChannelName, "channel", "", "Release channel to which the customer should be assigned")
	cmd.Flags().DurationVar(&opts.ExpiryDuration, "expires-in", 0, "If set, an expiration date will be set on the license. Supports Go durations like '72h' or '3600m'")
	cmd.Flags().BoolVar(&opts.EnsureChannel, "ensure-channel", false, "If set, channel will be created if it does not exist.")
	cmd.Flags().BoolVar(&opts.IsAirgapEnabled, "airgap", false, "If set, the license will allow airgap installs.")
	cmd.Flags().BoolVar(&opts.IsGitopsSupported, "gitops", false, "If set, the license will allow the GitOps usage.")
	cmd.Flags().BoolVar(&opts.IsSnapshotSupported, "snapshot", false, "If set, the license will allow Snapshots.")
	cmd.Flags().BoolVar(&opts.IsKotsInstallEnabled, "kots-install", true, "If set, the license will allow KOTS install. Otherwise license will allow Helm CLI installs only.")
	cmd.Flags().BoolVar(&opts.IsHelmInstallEnabled, "helm-install", false, "If set, the license will allow Helm installs.")
	cmd.Flags().BoolVar(&opts.IsKurlInstallEnabled, "kurl-install", false, "If set, the license will allow kURL installs.")
	cmd.Flags().BoolVar(&opts.IsEmbeddedClusterDownloadEnabled, "embedded-cluster-download", false, "If set, the license will allow Embedded Cluster downloads.")
	cmd.Flags().BoolVar(&opts.IsEmbeddedClusterMultinodeEnabled, "embedded-cluster-multinode", true, "If set, users can add nodes to Embedded Cluster instances.")
	cmd.Flags().BoolVar(&opts.IsGeoaxisSupported, "geo-axis", false, "If set, the license will allow Geo Axis usage.")
	cmd.Flags().BoolVar(&opts.IsHelmVMDownloadEnabled, "helmvm-cluster-download", false, "If set, the license will allow helmvm cluster downloads.")
	cmd.Flags().BoolVar(&opts.IsIdentityServiceSupported, "identity-service", false, "If set, the license will allow Identity Service usage.")
	cmd.Flags().BoolVar(&opts.IsInstallerSupportEnabled, "installer-support", false, "If set, the license will allow installer support.")
	cmd.Flags().BoolVar(&opts.IsSupportBundleUploadEnabled, "support-bundle-upload", false, "If set, the license will allow uploading support bundles.")
	cmd.Flags().BoolVar(&opts.IsDeveloperModeEnabled, "developer-mode", false, "If set, Replicated SDK installed in dev mode will use mock data.")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Email address of the customer that is to be created.")
	cmd.Flags().StringVar(&opts.CustomerType, "type", "dev", "The license type to create. One of: dev|trial|paid|community|test (default: dev)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	cmd.MarkFlagRequired("channel")

	return cmd
}

func (r *runners) createCustomer(cmd *cobra.Command, opts createCustomerOpts, outputFormat string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()

	// Validation
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if err := validateCustomerType(opts.CustomerType); err != nil {
		return errors.Wrap(err, "validate customer type")
	}
	if opts.CustomerType == "test" && opts.ExpiryDuration > time.Hour*48 {
		return errors.New("test licenses cannot be created with an expiration date greater than 48 hours")
	}
	if opts.CustomerType == "paid" {
		opts.CustomerType = "prod"
	}

	getOrCreateChannelOptions := client.GetOrCreateChannelOptions{
		AppID:          r.appID,
		AppType:        r.appType,
		NameOrID:       opts.ChannelName,
		Description:    "",
		CreateIfAbsent: opts.EnsureChannel,
	}

	channel, err := r.api.GetOrCreateChannelByName(getOrCreateChannelOptions)
	if err != nil {
		return errors.Wrap(err, "get channel")
	}

	channels := []kotsclient.CustomerChannel{
		{
			ID:        channel.ID,
			IsDefault: true,
		},
	}

	createOpts := kotsclient.CreateCustomerOpts{
		Name:                         opts.Name,
		CustomID:                     opts.CustomID,
		Channels:                     channels,
		AppID:                        r.appID,
		ExpiresAtDuration:            opts.ExpiryDuration,
		IsAirgapEnabled:              opts.IsAirgapEnabled,
		IsGitopsSupported:            opts.IsGitopsSupported,
		IsSnapshotSupported:          opts.IsSnapshotSupported,
		IsKotsInstallEnabled:         opts.IsKotsInstallEnabled,
		IsGeoaxisSupported:           opts.IsGeoaxisSupported,
		IsHelmVMDownloadEnabled:      opts.IsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:   opts.IsIdentityServiceSupported,
		IsInstallerSupportEnabled:    opts.IsInstallerSupportEnabled,
		IsSupportBundleUploadEnabled: opts.IsSupportBundleUploadEnabled,
		IsDeveloperModeEnabled:       opts.IsDeveloperModeEnabled,
		LicenseType:                  opts.CustomerType,
		Email:                        opts.Email,
	}

	if cmd.Flags().Changed("helm-install") {
		createOpts.IsHelmInstallEnabled = &opts.IsHelmInstallEnabled
	}
	if cmd.Flags().Changed("kurl-install") {
		createOpts.IsKurlInstallEnabled = &opts.IsKurlInstallEnabled
	}
	if cmd.Flags().Changed("embedded-cluster-download") {
		createOpts.IsEmbeddedClusterDownloadEnabled = &opts.IsEmbeddedClusterDownloadEnabled
	}
	if cmd.Flags().Changed("embedded-cluster-multinode") {
		createOpts.IsEmbeddedClusterMultinodeEnabled = &opts.IsEmbeddedClusterMultinodeEnabled
	}

	customer, err := r.api.CreateCustomer(r.appType, createOpts)
	if err != nil {
		return errors.Wrap(err, "create customer")
	}

	err = print.Customer(outputFormat, r.w, customer)
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

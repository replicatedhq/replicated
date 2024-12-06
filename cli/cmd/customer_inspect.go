package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (r *runners) InitCustomersInspectCommand(parent *cobra.Command) *cobra.Command {
	var (
		customer     string
		outputFormat string
	)
	cmd := &cobra.Command{
		Use:   "inspect [flags]",
		Short: "Show detailed information about a specific customer",
		Long: `The inspect command provides comprehensive details about a customer.

	This command retrieves and displays full information about a specified customer,
	including their assigned channels, registry information, and other relevant attributes.
	It's useful for getting an in-depth view of a customer's configuration and status.

	You must specify the customer using either their name or ID with the --customer flag.`,
		Example: `  # Inspect a customer by ID
	  replicated customer inspect --customer cus_abcdef123456

	  # Inspect a customer by name
	  replicated customer inspect --customer "Acme Inc"

	  # Inspect a customer and output in JSON format
	  replicated customer inspect --customer cus_abcdef123456 --output json

	  # Inspect a customer for a specific app (if you have multiple apps)
	  replicated customer inspect --app myapp --customer "Acme Inc"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.inspectCustomer(cmd, customer, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&customer, "customer", "", "The Customer Name or ID")
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.MarkFlagRequired("customer")

	return cmd
}

func (r *runners) inspectCustomer(cmd *cobra.Command, customer string, outputFormat string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if customer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	c, err := r.api.GetCustomerByNameOrId(r.appType, r.appID, customer)
	if err != nil {
		return errors.Wrapf(err, "get customer %q", customer)
	}

	ch, err := r.customerChannel(c)
	if err != nil {
		return errors.Wrap(err, "get customer channel")
	}

	regHost, err := r.registryHostname(c, ch)
	if err != nil {
		return errors.Wrapf(err, "get registry hostname for customer %q", customer)
	}

	if err = print.CustomerAttrs(outputFormat, r.w, r.appType, r.appSlug, ch, regHost, c); err != nil {
		return errors.Wrap(err, "print customer attrs")
	}

	return nil
}

func (r *runners) customerChannel(customer *types.Customer) (*types.KotsChannel, error) {
	var ch *types.Channel
	if len(customer.Channels) == 0 {
		return nil, nil
	}
	if len(customer.Channels) > 1 {
		fmt.Fprintln(os.Stderr, "WARNING: customer has multiple channels, using first channel")
	}
	ch = &customer.Channels[0]

	channel, err := r.kotsAPI.GetKotsChannel(r.appID, ch.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "list channels for customer %q", customer.Name)
	}

	return channel, nil
}

func (r *runners) registryHostname(customer *types.Customer, ch *types.KotsChannel) (string, error) {
	if ch != nil && ch.CustomHostNameOverrides.Registry.Hostname != "" {
		return ch.CustomHostNameOverrides.Registry.Hostname, nil
	}

	defaultCustomHostName, err := r.getDefaultCustomRegistryHostName()
	if err != nil {
		return "", err
	}

	if defaultCustomHostName != "" {
		return defaultCustomHostName, nil
	}

	if ch != nil && ch.ReplicatedRegistryDomain != "" {
		return ch.ReplicatedRegistryDomain, nil
	}

	envHostName := os.Getenv("REPLICATED_REGISTRY_ORIGIN")
	if envHostName != "" {
		return envHostName, nil
	}

	return "registry.replicated.com", nil
}

func (r *runners) getDefaultCustomRegistryHostName() (string, error) {
	if r.appType != "kots" {
		return "", nil
	}

	customHostnames, err := r.kotsAPI.ListCustomHostnames(r.appID)
	if err != nil {
		return "", err
	}

	for _, h := range customHostnames.Registry {
		if h.IsDefault {
			return h.Hostname, nil
		}
	}

	return "", nil
}

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
	cmd := &cobra.Command{
		Use:          "inspect",
		Short:        "Show full details for a customer",
		Long:         `Show full details for a customer`,
		RunE:         r.inspectCustomer,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.customerInspectCustomer, "customer", "", "The Customer Name or ID")
	cmd.Flags().StringVar(&r.args.customerInspectOutputFormat, "output", "text", "The output format to use. One of: json|text (default: text)")

	return cmd
}

func (r *runners) inspectCustomer(_ *cobra.Command, _ []string) error {
	if r.args.customerInspectCustomer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	customer, err := r.api.GetCustomerByName(r.appType, r.appID, r.args.customerInspectCustomer)
	if err != nil {
		return errors.Wrapf(err, "get customer %q", r.args.customerInspectCustomer)
	}

	ch, err := r.customerChannel(customer)
	if err != nil {
		return err
	}

	regHost, err := r.registryHostname(customer, ch)
	if err != nil {
		return errors.Wrapf(err, "get registry hostname for customer %q", r.args.customerInspectCustomer)
	}

	if err = print.CustomerAttrs(r.args.customerInspectOutputFormat, r.w, r.appType, r.appSlug, ch, regHost, customer); err != nil {
		return err
	}

	return nil
}

func (r *runners) customerChannel(customer *types.Customer) (*types.KotsChannel, error) {
	var ch *types.Channel
	if len(customer.Channels) == 0 {
		return nil, nil
	}
	if len(customer.Channels) > 1 {
		fmt.Println("WARNING: customer has multiple channels, using first channel")
	}
	ch = &customer.Channels[0]

	// TODO: fix GET /v3/app/{appID}/customer/{customerID} and/or GET /v3/app/{appID}/channel/{channelID} to
	// include chartNames in channels, like List does. Also ?channelName selector wrongly excludes details
	channels, err := r.kotsAPI.ListKotsChannels(r.appID, "", false)
	if err != nil {
		return nil, errors.Wrapf(err, "list channels for customer %q", customer.Name)
	}

	for _, detailedChan := range channels {
		if ch.ID == detailedChan.Id {
			return detailedChan, nil
		}
	}

	return nil, nil
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

package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
	"time"
)

func (r *runners) InitCustomerInstancesInspectCommand(parent *cobra.Command) *cobra.Command {
	customerInstanceInspectCommand := &cobra.Command{
		Use:   "inspect CUSTOMER INSTANCE",
		Short: "inspect a customer instance",
		Long:  `inspect a customer instance`,
		RunE:  r.inspectCustomerInstance,
	}
	parent.AddCommand(customerInstanceInspectCommand)

	return customerInstanceInspectCommand
}

func (r *runners) inspectCustomerInstance(_ *cobra.Command, args []string) error {

	if r.appType != "kots" {
		return errors.Errorf("unsupported app type: %q, only kots applications are supported", r.appType)
	}

	if len(args) != 2 {
		return errors.Errorf("requires a customer name/id and exactly one instance ID or ID prefix")
	}
	customerNameOrID := args[0]

	customer, err := r.api.GetCustomerByNameOrID(r.appType, r.appID, customerNameOrID)
	if err != nil {
		return errors.Wrapf(err, "get customer %q", customerNameOrID)
	}

	instanceIDPrefix := args[1]

	instance, err := r.api.GetInstanceByIDPrefix(r.appType, *customer, instanceIDPrefix)
	if err != nil {
		return errors.Wrapf(err, "get instance %q", instanceIDPrefix)
	}
	uptimeInterval := 4 * time.Hour

	startDate := time.Now().UTC().Add(-3 * 24 * time.Hour)
	uptime, err := r.api.GetInstanceUptime(r.appType, *instance, startDate, uptimeInterval)
	if err != nil {
		return errors.Wrapf(err, "get instance uptime for %s", instance.Instance.ID)
	}

	return print.InstanceInspect(instance, uptime, uptimeInterval)
}

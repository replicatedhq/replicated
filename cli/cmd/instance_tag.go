package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitInstanceTagCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tag",
		Short:        "tag an instance",
		Long:         `remove or add instance tags`,
		RunE:         r.tagInstance,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.instanceTagCustomer, "customer", "", "Customer Name or ID")
	cmd.Flags().StringVar(&r.args.instanceTagInstacne, "instance", "", "Instance Name or ID")
	cmd.Flags().StringArrayVar(&r.args.instanceTagTags, "tag", []string{}, "Tags to apply to instance. Leave value empty to remove tag. Tags not specified will not be removed.")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) tagInstance(cmd *cobra.Command, _ []string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()

	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.args.instanceTagCustomer == "" {
		return errors.Errorf("missing or invalid parameters: customer")
	}

	if r.args.instanceTagInstacne == "" {
		return errors.Errorf("missing or invalid parameters: instance")
	}

	if len(r.args.instanceTagTags) == 0 {
		return errors.Errorf("missing or invalid parameters: tag")
	}

	tags, err := parseTags(r.args.instanceTagTags)
	if err != nil {
		return errors.Wrap(err, "parse tags")
	}

	customer, err := r.api.GetCustomerByNameOrId(r.appType, r.appID, r.args.instanceTagCustomer)
	if err != nil {
		return errors.Wrapf(err, "find customer %q", r.args.instanceTagCustomer)
	}

	instance, err := findInstanceByNameOrID(r.args.instanceTagInstacne, customer.Instances)
	if err != nil {
		return errors.Wrap(err, "find instance")
	}

	updatedInstance, err := r.api.SetInstanceTags(r.appID, r.appType, customer.ID, instance.InstanceID, tags)
	if err != nil {
		return errors.Wrap(err, "set instance tags")
	}

	// Using `ls` output here instead of `inspect` because tags API does not return version history
	if err := print.Instances(r.outputFormat, r.w, []types.Instance{*updatedInstance}); err != nil {
		return errors.Wrap(err, "print instance")
	}

	return nil
}

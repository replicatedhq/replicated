package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitCustomersCommand(parent *cobra.Command) *cobra.Command {
	customersCmd := &cobra.Command{
		Use:   "customer",
		Short: "Manage customers",
		Long:  `The customers command allows vendors to create, display, modify end customer records.`,
	}
	parent.AddCommand(customersCmd)

	return customersCmd
}

// resolveCustomer looks up a customer by ID first, then by name in the current app.
// includeArchived controls whether the name lookup also matches archived customers.
func (r *runners) resolveCustomer(nameOrID string, includeArchived bool) (*types.Customer, error) {
	c, err := r.api.GetCustomerByID(nameOrID)
	if err == nil {
		return c, nil
	}
	if errors.Cause(err) != platformclient.ErrNotFound {
		return nil, errors.Wrapf(err, "find customer %q", nameOrID)
	}

	if !r.hasApp() {
		return nil, errors.New("no app specified: app is required when looking up customers by name")
	}

	if includeArchived {
		c, err = r.api.GetCustomerByNameIncludeArchived(r.appID, nameOrID)
	} else {
		c, err = r.api.GetCustomerByName(r.appID, nameOrID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "find customer %q", nameOrID)
	}

	return c, nil
}

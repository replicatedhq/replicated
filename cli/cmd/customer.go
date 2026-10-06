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
func (r *runners) resolveCustomer(nameOrID string) (*types.Customer, error) {
	customers, err := r.resolveCustomers(nameOrID, func(appID, name string) ([]types.Customer, error) {
		c, err := r.api.GetCustomerByName(appID, name)
		if err != nil {
			return nil, err
		}
		return []types.Customer{*c}, nil
	})
	if err != nil {
		return nil, err
	}
	return &customers[0], nil
}

// resolveCustomers looks up a customer by ID first. If no customer has that ID, it falls back
// to byName, which looks up customers by name in the current app.
func (r *runners) resolveCustomers(nameOrID string, byName func(appID, name string) ([]types.Customer, error)) ([]types.Customer, error) {
	c, err := r.api.GetCustomerByID(nameOrID)
	if err == nil {
		return []types.Customer{*c}, nil
	}
	if errors.Cause(err) != platformclient.ErrNotFound {
		return nil, errors.Wrapf(err, "find customer %q", nameOrID)
	}

	if !r.hasApp() {
		return nil, errors.New("no app specified: app is required when looking up customers by name")
	}

	customers, err := byName(r.appID, nameOrID)
	if err != nil {
		return nil, errors.Wrapf(err, "find customer %q", nameOrID)
	}

	return customers, nil
}

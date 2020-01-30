package client

import (
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCustomers(appID string, appType string) ([]types.Customer, error) {

	if appType == "platform" {
		return nil, errors.New("listing customers is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("listing customers is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.ListCustomers(appID)
	}

	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) CreateCustomer(appType string, name string, channelID string, expiresIn time.Duration) (*types.Customer, error) {
	if appType == "platform" {
		return nil, errors.New("creating customers is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("creating customers is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.CreateCustomer(name, channelID, expiresIn)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}

func (c *Client) FindCustomerByNameOrID(appID, appType, customerNameOrID string) (*types.Customer, error) {
	customers, err := c.ListCustomers(appID, appType)
	if err != nil {
		return nil, errors.Wrap(err, "list customers")
	}

	var foundCustomers []types.Customer
	for _, customer := range customers {
		if customer.Name == customerNameOrID || customer.ID == customerNameOrID {
			foundCustomers = append(foundCustomers, customer)
		}
	}

	if len(foundCustomers) == 0 {
		return nil, errors.Errorf("customer with name or ID %q not found", customerNameOrID)
	}

	if len(foundCustomers) > 1 {
		return nil, fmt.Errorf("customer %q is ambiguous, please use customer ID", customerNameOrID)
	}

	return &foundCustomers[0], nil
}

func (c *Client) FetchLicense(appType string, appSlug string, customerID string) ([]byte, error) {

	if appType == "platform" {
		return nil, errors.New("fetching licenses is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("fetching licenses is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.FetchLicense(appSlug, customerID)
	}

	return nil, errors.Errorf("unknown app type %q", appType)
}

package client

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCustomers(appID string, appType string) ([]types.Customer, error) {

	if appType == "platform" {
		return nil, errors.New("listing customers is not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.ListCustomers(appID)
	}

	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) CreateCustomer(appType string, opts kotsclient.CreateCustomerOpts) (*types.Customer, error) {
	if appType == "platform" {
		return nil, errors.New("creating customers is not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.CreateCustomer(opts)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}

func (c *Client) GetCustomerByName(appType string, appID, name string) (*types.Customer, error) {
	if appType == "platform" {
		return nil, errors.New("listing customers is not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.GetCustomerByName(appID, name)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}

func (c *Client) DownloadLicense(appType string, appID string, customerID string) ([]byte, error) {
	if appType == "platform" {
		return nil, errors.New("downloading customer licenses is not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.DownloadLicense(appID, customerID)
	}
	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) ArchiveCustomer(appType string, customerID string) error {
	if appType == "platform" {
		return errors.New("archiving customer is not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.ArchiveCustomer(customerID)
	}
	return errors.Errorf("unknown app type %q", appType)
}

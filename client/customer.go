package client

import (
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

package client

import (
	"errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) CreateEntitlementSpec(name string, spec string, appID string) (*types.EntitlementSpec, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		return nil, errors.New("This feature is not supported for platform applications.")
	} else if appType == "ship" {
		c.ShipClient.CreateEntitlementSpec(appID, name, spec)
	} else if appType == "kots" {
		return nil, errors.New("This feature is not supported for kots applications.")
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) SetDefaultEntitlementSpec(specID string) error {
	return c.ShipClient.SetDefaultEntitlementSpec(specID)
}

func (c *Client) SetEntitlementValue(customerID string, specID string, key string, value string, datatype string, appID string) (*types.EntitlementValue, error) {
	return c.ShipClient.SetEntitlementValue(customerID, specID, key, value, datatype, appID)
}

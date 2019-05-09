package client

import (
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) CreateEntitlementSpec(name string, spec string, appID string) (*types.EntitlementSpec, error) {
	return c.ShipClient.CreateEntitlementSpec(appID, name, spec)
}

func (c *Client) SetDefaultEntitlementSpec(specID string) error {
	return c.ShipClient.SetDefaultEntitlementSpec(specID)
}

func (c *Client) SetEntitlementValue(customerID string, specID string, key string, value string, datatype string, appID string) (*types.EntitlementValue, error) {
	return c.ShipClient.SetEntitlementValue(customerID, specID, key, value, datatype, appID)
}

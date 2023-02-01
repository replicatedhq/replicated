package client

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"time"
)

func (c *Client) ListCustomerInstances(appType string, customer types.Customer) ([]types.CustomerInstance, error) {
	if appType == "platform" {
		return nil, errors.New("customer instance inspection is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("customer instance inspection is not supported for ship applications")
	} else if appType == "kots" {
		return c.InstancesClient.ListInstances(customer)
	}
	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) GetInstanceByIDPrefix(appType string, customer types.Customer, idPrefix string) (*types.CustomerInstance, error) {
	if appType == "platform" {
		return nil, errors.New("listing instances is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("listing instances is not supported for ship applications")
	} else if appType == "kots" {
		return c.InstancesClient.GetInstanceByIDPrefix(customer, idPrefix)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}
func (c *Client) GetInstanceUptime(appType string, instance types.CustomerInstance, startTime time.Time, uptimeInterval time.Duration) (*types.CustomerInstanceUptime, error) {
	if appType == "platform" {
		return nil, errors.New("listing instances is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("listing instances is not supported for ship applications")
	} else if appType == "kots" {
		return c.InstancesClient.GetInstanceUptime(instance, startTime, uptimeInterval)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}

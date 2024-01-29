package client

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCollectors(appID string, appType string) ([]types.CollectorSpec, error) {
	if appType == "kots" {
		return nil, errors.New("On a kots application, users must modify the support-bundle.yaml file in the release")
	}

	shipappCollectors, err := c.PlatformClient.ListCollectors(appID, appType)
	if err != nil {
		return nil, err
	}

	return shipappCollectors, nil
}

func (c *Client) UpdateCollector(appID string, specID string, yaml string, name string, isArchived bool) error {
	return c.PlatformClient.UpdateCollector(appID, specID, yaml, name, isArchived)
}

func (c *Client) CreateCollector(appID string, appType string, name string, yaml string) (*types.CollectorSpec, error) {
	if appType == "kots" {
		return nil, errors.New("On a kots application, users must modify the support-bundle.yaml file in the release")
	}

	return c.PlatformClient.CreateCollector(appID, name, yaml)

}

func (c *Client) GetCollector(appID string, specID string) (*types.CollectorSpec, error) {
	return c.PlatformClient.GetCollector(appID, specID)

}

func (c *Client) PromoteCollector(appID string, appType string, specID string, channelIDs ...string) error {

	if appType == "platform" {
		return c.PlatformClient.PromoteCollector(appID, specID, channelIDs...)
	}

	return errors.New("unknown app type")
}

package client

import (
	"github.com/pkg/errors"
	collectors "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCollectors(appID string, appType string) ([]types.CollectorInfo, error) {

	return nil, errors.New("On a kots application, users must modify the support-bundle.yaml file in the release")
}

// func (c *Client) CreateCollector(appID string, name string, yaml string) (*collectors.AppCollectorInfo, error) {
func (c *Client) CreateCollector(appID string, appType string, name string, yaml string) (*collectors.AppCollectorInfo, error) {

	return nil, errors.New("On a kots application, users must modify the support-bundle.yaml file in the release")
}

func (c *Client) PromoteCollector(appID string, appType string, specID string, channelIDs ...string) error {

	if appType == "platform" {
		return c.PlatformClient.PromoteCollector(appID, specID, channelIDs...)
	}

	return errors.New("unknown app type")
}

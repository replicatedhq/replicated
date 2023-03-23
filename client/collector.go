package client

import (
	"github.com/pkg/errors"
	collectors "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCollectors(appID string, appType string) ([]types.CollectorInfo, error) {

	if appType == "kots" {
		return nil, errors.New("On a kots application, users must modify the support-bundle.yaml file in the release")
	}

	shipappCollectors, err := c.ShipClient.ListCollectors(appID, appType)
	if err != nil {
		return nil, err
	}

	shipCollectorInfos := make([]types.CollectorInfo, 0, 0)
	for _, shipappCollector := range shipappCollectors {
		activeChannels := make([]types.Channel, 0, 0)
		for _, shipappCollectorChannel := range shipappCollector.ActiveChannels {
			activeChannel := types.Channel{
				ID:   shipappCollectorChannel.ID,
				Name: shipappCollectorChannel.Name,
			}

			activeChannels = append(activeChannels, activeChannel)
		}
		shipCollectorInfo := types.CollectorInfo{
			AppID:          shipappCollector.AppID,
			CreatedAt:      shipappCollector.CreatedAt,
			Name:           shipappCollector.Name,
			ActiveChannels: activeChannels,
			SpecID:         shipappCollector.SpecID,
		}

		shipCollectorInfos = append(shipCollectorInfos, shipCollectorInfo)
	}

	return shipCollectorInfos, nil
}

func (c *Client) UpdateCollector(appID string, specID string, yaml string) (interface{}, error) {

	return c.ShipClient.UpdateCollector(appID, specID, yaml)
}

func (c *Client) UpdateCollectorName(appID string, specID string, name string) (interface{}, error) {

	return c.ShipClient.UpdateCollectorName(appID, specID, name)

}

// func (c *Client) CreateCollector(appID string, name string, yaml string) (*collectors.AppCollectorInfo, error) {
func (c *Client) CreateCollector(appID string, appType string, name string, yaml string) (*collectors.AppCollectorInfo, error) {

	if appType == "kots" {
		return nil, errors.New("On a kots application, users must modify the support-bundle.yaml file in the release")
	}
	return c.ShipClient.CreateCollector(appID, name, yaml)

}

func (c *Client) GetCollector(appID string, specID string) (*collectors.AppCollectorInfo, error) {
	return c.ShipClient.GetCollector(appID, specID)

}

func (c *Client) PromoteCollector(appID string, appType string, specID string, channelIDs ...string) error {

	if appType == "platform" {
		return c.PlatformClient.PromoteCollector(appID, specID, channelIDs...)
	}

	return errors.New("unknown app type")
}

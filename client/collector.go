package client

import (
	"errors"

	"github.com/replicatedhq/replicated/pkg/types"
	// collectors "github.com/replicatedhq/replicated/gen/go/v1"
)

func (c *Client) ListCollectors(appID string) ([]types.CollectorInfo, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		platformCollectors, err := c.PlatformClient.ListCollectors(appID)
		if err != nil {
			return nil, err
		}

		collectorInfos := make([]types.CollectorInfo, 0, 0)
		for _, platformCollector := range platformCollectors {
			activeChannels := make([]types.Channel, 0, 0)
			for _, platformCollectorChannel := range platformCollector.ActiveChannels {
				activeChannel := types.Channel{
					ID:   platformCollectorChannel.Id,
					Name: platformCollectorChannel.Name,
				}

				activeChannels = append(activeChannels, activeChannel)
			}
			collectorInfo := types.CollectorInfo{
				AppID:          platformCollector.AppId,
				CreatedAt:      platformCollector.CreatedAt,
				Name:           platformCollector.Name,
				ActiveChannels: activeChannels,
				SpecID:         platformCollector.SpecId,
			}

			collectorInfos = append(collectorInfos, collectorInfo)
		}

		return collectorInfos, nil
	} else if appType == "ship" {
		shipappCollectors, err := c.ShipClient.ListCollectors(appID)
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

	return nil, errors.New("unknown app type")
}

func (c *Client) UpdateCollector(appID string, name string, collectorOptions interface{}) error {
	return nil
}

func (c *Client) GetCollector(appID string, id string) (interface{}, error) {
	return nil, nil
}

func (c *Client) PromoteCollector(appID string, specID string, channelIDs ...string) error {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return err
	}

	if appType == "platform" {
		return c.PlatformClient.PromoteCollector(appID, specID, channelIDs...)
	} else if appType == "ship" {
		return c.ShipClient.PromoteCollector(appID, specID, channelIDs...)
	}

	return errors.New("unknown app type")
}

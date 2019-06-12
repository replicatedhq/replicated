package client

import (
	"errors"

	collectors "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCollectors(appID string, appType string) ([]types.CollectorInfo, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		platformCollectors, err := c.PlatformClient.ListCollectors(appID, appType)
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
		shipappCollectors, err := c.ShipClient.ListCollectors(appID, appType)
		if err != nil {
			return nil, err
		}

		shipCollectorInfos := make([]types.CollectorInfo, 0, 0)
		for _, shipappCollector := range shipappCollectors {
			activeChannels := make([]types.Channel, 0, 0)
			for _, shipappCollectorChannel := range shipappCollector.ActiveChannels {
				activeChannel := types.Channel{
					ID:   shipappCollectorChannel.Id,
					Name: shipappCollectorChannel.Name,
				}

				activeChannels = append(activeChannels, activeChannel)
			}
			shipCollectorInfo := types.CollectorInfo{
				AppID:          shipappCollector.AppId,
				CreatedAt:      shipappCollector.CreatedAt,
				Name:           shipappCollector.Name,
				ActiveChannels: activeChannels,
				SpecID:         shipappCollector.SpecId,
			}

			shipCollectorInfos = append(shipCollectorInfos, shipCollectorInfo)
		}

		return shipCollectorInfos, nil
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) UpdateCollector(appID string, appType string, specID string, yaml string) (interface{}, error) {
	if appType == "platform" {
		return c.PlatformClient.UpdateCollector(appID, appType, specID, yaml)
	} else if appType == "ship" {
		return c.ShipClient.UpdateCollector(appID, appType, specID, yaml)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) UpdateCollectorName(appID string, appType string, specID string, name string) (interface{}, error) {
	if appType == "platform" {
		return c.PlatformClient.UpdateCollectorName(appID, appType, specID, name)
	} else if appType == "ship" {
		return c.ShipClient.UpdateCollectorName(appID, appType, specID, name)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) CreateCollector(appID string, appType string, yaml string) (*collectors.AppCollectorInfo, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		return c.PlatformClient.CreateCollector(appID, appType, yaml)
	} else if appType == "ship" {
		return c.ShipClient.CreateCollector(appID, appType, yaml)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) GetCollector(appID string, appType string, specID string) (interface{}, error) {
	return nil, nil
}

func (c *Client) PromoteCollector(appID string, appType string, specID string, channelIDs ...string) error {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return err
	}

	if appType == "platform" {
		return c.PlatformClient.PromoteCollector(appID, appType, specID, channelIDs...)
	} else if appType == "ship" {
		return c.ShipClient.PromoteCollector(appID, appType, specID, channelIDs...)
	}

	return errors.New("unknown app type")
}

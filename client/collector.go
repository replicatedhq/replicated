package client

import (
	"errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCollectors(appID string) ([]types.CollectorInfo, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		platformReleases, err := c.PlatformClient.ListCollectors(appID)
		if err != nil {
			return nil, err
		}

		collectorInfos := make([]types.CollectorInfo, 0, 0)
		for _, platformRelease := range platformReleases {
			activeChannels := make([]types.Channel, 0, 0)
			for _, platformReleaseChannel := range platformRelease.ActiveChannels {
				activeChannel := types.Channel{
					ID:          platformReleaseChannel.Id,
					Name:        platformReleaseChannel.Name,
					Description: platformReleaseChannel.Description,
				}

				activeChannels = append(activeChannels, activeChannel)
			}
			collectorInfo := types.CollectorInfo{
				AppID:          platformRelease.AppId,
				CreatedAt:      platformRelease.CreatedAt,
				Name:           platformRelease.Name,
				ActiveChannels: activeChannels,
			}

			collectorInfos = append(collectorInfos, collectorInfo)
		}

		return collectorInfos, nil
	} else if appType == "ship" {
		return c.ShipClient.ListCollectors(appID)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) CreateCollector(appID string, name string, yaml string) (*types.CollectorInfo, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		platformReleaseInfo, err := c.PlatformClient.CreateCollector(appID, name, yaml)
		if err != nil {
			return nil, err
		}

		activeChannels := make([]types.Channel, 0, 0)

		for _, platformReleaseChannel := range platformReleaseInfo.ActiveChannels {
			activeChannel := types.Channel{
				ID:          platformReleaseChannel.Id,
				Name:        platformReleaseChannel.Name,
				Description: platformReleaseChannel.Description,
			}

			activeChannels = append(activeChannels, activeChannel)
		}
		return &types.CollectorInfo{
			AppID:          platformReleaseInfo.AppId,
			CreatedAt:      platformReleaseInfo.CreatedAt,
			Name:           platformReleaseInfo.Name,
			ActiveChannels: activeChannels,
		}, nil
	} else if appType == "ship" {
		return c.ShipClient.CreateCollector(appID, name, yaml)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) UpdateCollector(appID string, name string, collectorOptions interface{}) error {
	return nil
}

func (c *Client) GetCollector(appID string, name string) (interface{}, error) {
	return nil, nil
}

func (c *Client) PromoteCollector(appID string, name string, channelIDs ...string) error {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return err
	}

	if appType == "platform" {
		return c.PlatformClient.PromoteCollector(appID, name, channelIDs...)
	} else if appType == "ship" {
		return c.ShipClient.PromoteCollector(appID, name, channelIDs...)
	}

	return errors.New("unknown app type")
}

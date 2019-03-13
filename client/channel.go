package client

import (
	"errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListChannels(appID string) ([]types.Channel, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		platformChannels, err := c.PlatformClient.ListChannels(appID)
		if err != nil {
			return nil, err
		}

		channels := make([]types.Channel, 0, 0)
		for _, platformChannel := range platformChannels {
			channel := types.Channel{
				ID:              platformChannel.Id,
				Name:            platformChannel.Name,
				Description:     platformChannel.Description,
				ReleaseSequence: platformChannel.ReleaseSequence,
				ReleaseLabel:    platformChannel.ReleaseLabel,
			}

			channels = append(channels, channel)
		}

		return channels, nil
	} else if appType == "ship" {
		return c.ShipClient.ListChannels(appID)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) GetChannel(appID string, channelID string) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func (c *Client) ArchiveChannel(appID string, channelID string) error {
	return nil
}

func (c *Client) CreateChannel(appID string, name string, description string) ([]types.Channel, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		if err := c.PlatformClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.ListChannels(appID)
	} else if appType == "ship" {
		if err := c.ShipClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.ShipClient.ListChannels(appID)
	}

	return nil, errors.New("unknown app type")
}

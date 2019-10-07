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
	} else if appType == "kots" {
		return c.KotsClient.ListChannels(appID)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) GetChannel(appID string, channelID string) (interface{}, interface{}, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, nil, err
	}

	if appType == "platform" {
		return c.PlatformClient.GetChannel(appID, channelID)
	} else if appType == "ship" {
		return c.ShipClient.GetChannel(appID, channelID)
	} else if appType == "kots" {
		// return c.KotsClient.GetChannel(appID, channelID)
		return nil, nil, nil
	}
	return nil, nil, errors.New("unknown app type")

}

func (c *Client) ArchiveChannel(appID string, channelID string) error {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return err
	}

	if appType == "platform" {
		return c.PlatformClient.ArchiveChannel(appID, channelID)
	} else if appType == "ship" {
		// return c.ShipClient.ArchiveChannel(appID, channelID)
		return nil
	} else if appType == "kots" {
		// return c.KotsClient.ArchiveChannel(appID, channelID)
		return nil
	}
	return errors.New("unknown app type")

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
	} else if appType == "kots" {
		if err := c.KotsClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.KotsClient.ListChannels(appID)
	}

	return nil, errors.New("unknown app type")
}

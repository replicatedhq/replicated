package client

import (
	"errors"

	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListChannels(appID string, appType string) ([]types.Channel, error) {

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

func (c *Client) GetChannel(appID string, appType string, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error) {

	if appType == "platform" {
		return c.PlatformClient.GetChannel(appID, channelID)
	} else if appType == "ship" {
		return c.ShipClient.GetChannel(appID, channelID)
	} else if appType == "kots" {
		return c.KotsClient.GetChannel(appID, channelID)
	}
	return nil, nil, errors.New("unknown app type")

}

func (c *Client) ArchiveChannel(appID string, appType string, channelID string) error {

	if appType == "platform" {
		return c.PlatformClient.ArchiveChannel(appID, channelID)
	} else if appType == "ship" {
		return errors.New("This feature is not currently supported for Ship applications.")
	} else if appType == "kots" {
		return errors.New("This feature is not currently supported for Kots applications.")
	}
	return errors.New("unknown app type")

}

func (c *Client) CreateChannel(appID string, appType string, name string, description string) ([]types.Channel, error) {

	if appType == "platform" {
		if err := c.PlatformClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.ListChannels(appID, appType)
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

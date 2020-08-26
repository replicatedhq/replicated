package client

import (
	"fmt"
	"github.com/pkg/errors"

	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListChannels(appID string, appType string, appSlug string) ([]types.Channel, error) {

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
		return c.KotsHTTPClient.ListChannels(appID, appSlug)
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

func (c *Client) CreateChannel(appID string, appType string, appSlug string, name string, description string) ([]types.Channel, error) {

	if appType == "platform" {
		if err := c.PlatformClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.ListChannels(appID, appType, appSlug)
	} else if appType == "ship" {
		if _, err := c.ShipClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.ShipClient.ListChannels(appID)
	} else if appType == "kots" {
		if _, err := c.KotsClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.KotsHTTPClient.ListChannels(appID, appSlug)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) GetOrCreateChannelByName(appID string, appType string, appSlug string, name string, description string, createIfAbsent bool) (*types.Channel, error) {
	allChannels, err := c.ListChannels(appID, appType, appSlug)
	if err != nil {
		return nil, err
	}

	foundChannel, numMatching, err := c.findChannel(allChannels, name)

	if numMatching == 0 && createIfAbsent {
		updatedListOfChannels, err := c.CreateChannel(appID, appType, appSlug, name, description)
		if err != nil {
			return nil, errors.Wrapf(err, "create channel %q ", name)
		}
		// for some reason CreateChannel returns the list of all channels,
		// so now we gotta go find the channel we just created
		channel, _, err := c.findChannel(updatedListOfChannels, name)
		return channel, errors.Wrapf(err, "find channel %q")
	}

	return foundChannel, errors.Wrapf(err, "find channel %q")
}

func (c *Client) GetChannelByName(appID string, appType string, appSlug string, name string) (*types.Channel, error) {
	return c.GetOrCreateChannelByName(appID, appType, appSlug, name, "", false)
}

func (c *Client) findChannel(channels []types.Channel, name string) (*types.Channel, int, error) {

	matchingChannels := make([]*types.Channel, 0)
	for _, channel := range channels {
		if channel.ID == name || channel.Name == name {
			matchingChannels = append(matchingChannels, &channel)
		}
	}
	if len(matchingChannels) == 0 {
		return nil, 0, errors.Errorf("No channel %q ", name)
	}

	if len(matchingChannels) > 1 {
		return nil, len(matchingChannels), fmt.Errorf("channel %q is ambiguous, please use channel ID", name)
	}
	return matchingChannels[0], 1, nil
}

package client

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListChannels(appID string, appType string, channelName string) ([]types.Channel, error) {

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
	} else if appType == "kots" {
		return c.KotsClient.ListChannels(appID, channelName)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) GetChannel(appID string, appType string, channelID string) (*types.Channel, error) {
	if appType == "platform" {
		platformChannel, _, err := c.PlatformClient.GetChannel(appID, channelID)
		if err != nil {
			return nil, err
		}
		channel := types.Channel{
			ID:              platformChannel.Id,
			Name:            platformChannel.Name,
			Description:     platformChannel.Description,
			ReleaseSequence: platformChannel.ReleaseSequence,
			ReleaseLabel:    platformChannel.ReleaseLabel,
		}
		return &channel, nil
	} else if appType == "kots" {
		return c.KotsClient.GetChannel(appID, channelID)
	}
	return nil, errors.New("unknown app type")
}

func (c *Client) ArchiveChannel(appID string, appType string, channelID string) error {
	if appType == "platform" {
		return c.PlatformClient.ArchiveChannel(appID, channelID)
	} else if appType == "kots" {
		return c.KotsClient.ArchiveChannel(appID, channelID)
	}
	return errors.New("unknown app type")

}

func (c *Client) CreateChannel(appID string, appType string, name string, description string) ([]types.Channel, error) {

	if appType == "platform" {
		if err := c.PlatformClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.ListChannels(appID, appType, name)
	} else if appType == "kots" {
		if _, err := c.KotsClient.CreateChannel(appID, name, description); err != nil {
			return nil, err
		}
		return c.KotsClient.ListChannels(appID, name)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) GetOrCreateChannelByName(appID string, appType string, nameOrID string, description string, createIfAbsent bool) (*types.Channel, error) {
	gqlNotFoundErr := fmt.Sprintf("channel %s not found", nameOrID)
	channel, err := c.GetChannel(appID, appType, nameOrID)
	if err == nil {
		return channel, nil
	} else if !strings.Contains(err.Error(), gqlNotFoundErr) && !errors.Is(err, platformclient.ErrNotFound) {
		return nil, errors.Wrap(err, "get channel")
	}

	allChannels, err := c.ListChannels(appID, appType, nameOrID)
	if err != nil {
		return nil, err
	}

	foundChannel, numMatching, err := c.findChannel(allChannels, nameOrID)

	if numMatching == 0 && createIfAbsent {
		updatedListOfChannels, err := c.CreateChannel(appID, appType, nameOrID, description)
		if err != nil {
			return nil, errors.Wrapf(err, "create channel %q ", nameOrID)
		}
		// for some reason CreateChannel returns the list of all channels,
		// so now we gotta go find the channel we just created
		channel, _, err := c.findChannel(updatedListOfChannels, nameOrID)
		return channel, errors.Wrapf(err, "find channel %q", nameOrID)
	}

	return foundChannel, errors.Wrapf(err, "find channel %q", nameOrID)
}

func (c *Client) GetChannelByName(appID string, appType string, name string) (*types.Channel, error) {
	return c.GetOrCreateChannelByName(appID, appType, name, "", false)
}

func (c *Client) findChannel(channels []types.Channel, name string) (*types.Channel, int, error) {
	matchingChannels := make([]*types.Channel, 0)
	for _, channel := range channels {
		if channel.ID == name || channel.Name == name {
			matchingChannels = append(matchingChannels, channel.Copy())
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

func (c *Client) UpdateSemanticVersioningForChannel(appType string, appID string, chanID string, enableSemver bool) error {
	if appType == "platform" {
		return errors.New("This feature is not currently supported for Platform applications.")
	} else if appType == "kots" {
		channel, err := c.KotsClient.GetChannel(appID, chanID)
		if err != nil {
			return err
		}
		err = c.KotsClient.UpdateSemanticVersioning(appID, channel, enableSemver)
		return err
	}

	return errors.New("unknown app type")
}

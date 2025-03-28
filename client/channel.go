package client

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListChannels(appID string, appType string, channelName string) ([]*types.Channel, error) {

	if appType == "platform" {
		platformChannels, err := c.PlatformClient.ListChannels(appID)
		if err != nil {
			return nil, err
		}

		channels := make([]*types.Channel, 0, 0)
		for _, platformChannel := range platformChannels {
			channel := &types.Channel{
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

	return nil, errors.Errorf("unknown app type %q", appType)
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
	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) ArchiveChannel(appID string, appType string, channelID string) error {
	if appType == "platform" {
		return c.PlatformClient.ArchiveChannel(appID, channelID)
	} else if appType == "kots" {
		return c.KotsClient.ArchiveChannel(appID, channelID)
	}
	return errors.Errorf("unknown app type %q", appType)

}

func (c *Client) CreateChannel(appID string, appType string, name string, description string) ([]*types.Channel, error) {

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

	return nil, errors.Errorf("unknown app type %q", appType)
}

type GetOrCreateChannelOptions struct {
	AppID          string
	AppType        string
	NameOrID       string
	Description    string
	CreateIfAbsent bool
}

func (c *Client) GetOrCreateChannelByName(opts GetOrCreateChannelOptions) (*types.Channel, error) {
	gqlNotFoundErr := fmt.Sprintf("channel %s not found", opts.NameOrID)
	// The backing api endpoint technically only accepts a channel id, and names will always 404.
	// We rely on this to fall through to the list channels call when passed as a name. Additionally,,
	// in some rbac configurations, we get a 403 instead of a 404 when a channel name is passed
	// even if the user has access to the channel id. In exchange for still getting a early return
	// when we're passed a channel id that exists and is accessible we need to fall through in both cases.
	channel, err := c.GetChannel(opts.AppID, opts.AppType, opts.NameOrID)
	if err == nil {
		return channel, nil
	} else if !strings.Contains(err.Error(), gqlNotFoundErr) && !errors.Is(err, platformclient.ErrNotFound) && !errors.Is(err, platformclient.ErrForbidden) {
		return nil, errors.Wrap(err, "get channel")
	}

	allChannels, err := c.ListChannels(opts.AppID, opts.AppType, opts.NameOrID)
	if err != nil {
		return nil, err
	}

	foundChannel, numMatching, err := c.findChannel(allChannels, opts.NameOrID)

	if numMatching == 0 && opts.CreateIfAbsent {
		updatedListOfChannels, err := c.CreateChannel(opts.AppID, opts.AppType, opts.NameOrID, opts.Description)
		if err != nil {
			return nil, errors.Wrapf(err, "create channel %q ", opts.NameOrID)
		}
		// for some reason CreateChannel returns the list of all channels,
		// so now we gotta go find the channel we just created
		channel, _, err := c.findChannel(updatedListOfChannels, opts.NameOrID)
		return channel, errors.Wrapf(err, "find channel %q", opts.NameOrID)
	}

	return foundChannel, errors.Wrapf(err, "find channel %q", opts.NameOrID)
}

func (c *Client) GetChannelByName(appID string, appType string, name string) (*types.Channel, error) {
	opts := GetOrCreateChannelOptions{
		AppID:          appID,
		AppType:        appType,
		NameOrID:       name,
		CreateIfAbsent: false,
	}
	return c.GetOrCreateChannelByName(opts)
}

func (c *Client) findChannel(channels []*types.Channel, nameOrID string) (*types.Channel, int, error) {
	matchingChannels := make([]*types.Channel, 0)
	for _, channel := range channels {
		if channel.ID == nameOrID || channel.Name == nameOrID {
			matchingChannels = append(matchingChannels, channel)
		}
	}
	if len(matchingChannels) == 0 {
		return nil, 0, errors.Errorf("No channel %q ", nameOrID)
	}

	if len(matchingChannels) > 1 {
		return nil, len(matchingChannels), fmt.Errorf("channel %q is ambiguous, please use channel ID", nameOrID)
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

	return errors.Errorf("unknown app type %q", appType)
}

func (c *Client) ChannelReleaseDemote(appID string, appType string, channelID string, channelSequence int64) (*types.ChannelRelease, error) {
	if appType == "platform" {
		return nil, errors.New("This feature is not currently supported for Platform applications.")
	} else if appType == "kots" {
		return c.KotsClient.DemoteChannelRelease(appID, channelID, channelSequence)
	}
	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) ChannelReleaseUnDemote(appID string, appType string, channelID string, channelSequence int64) (*types.ChannelRelease, error) {
	if appType == "platform" {
		return nil, errors.New("This feature is not currently supported for Platform applications.")
	} else if appType == "kots" {
		return c.KotsClient.UnDemoteChannelRelease(appID, channelID, channelSequence)
	}
	return nil, errors.Errorf("unknown app type %q", appType)
}

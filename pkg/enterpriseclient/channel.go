package enterpriseclient

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

func (c HTTPClient) ListChannels() ([]*enterprisetypes.Channel, error) {
	enterpriseChannels := []*enterprisetypes.Channel{}
	err := c.doJSON("GET", "/v1/channels", 200, nil, &enterpriseChannels)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channels")
	}

	return enterpriseChannels, nil
}

func (c HTTPClient) CreateChannel(name string, description string) (*enterprisetypes.Channel, error) {
	type CreateChannelRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	createChannelRequest := CreateChannelRequest{
		Name:        name,
		Description: description,
	}

	enterpriseChannel := enterprisetypes.Channel{}
	err := c.doJSON("POST", "/v1/channel", 201, createChannelRequest, &enterpriseChannel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create channel")
	}

	return &enterpriseChannel, nil
}

func (c HTTPClient) UpdateChannel(id string, name string, description string) (*enterprisetypes.Channel, error) {
	type UpdateChannelRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	updateChannelRequest := UpdateChannelRequest{
		Name:        name,
		Description: description,
	}

	enterpriseChannel := enterprisetypes.Channel{}

	err := c.doJSON("PUT", fmt.Sprintf("/v1/channel/%s", id), 200, updateChannelRequest, &enterpriseChannel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update channel")
	}

	return &enterpriseChannel, nil
}

func (c HTTPClient) RemoveChannel(id string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/v1/channel/%s", id), 204, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete channel")
	}

	return nil
}

func (c HTTPClient) AssignChannel(channelID string, teamID string) error {
	err := c.doJSON("POST", fmt.Sprintf("/v1/teamchannel/%s/%s", channelID, teamID), 204, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to assign channel")
	}

	return nil
}

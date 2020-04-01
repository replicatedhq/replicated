package enterpriseclient

import (
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

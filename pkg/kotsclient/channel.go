package kotsclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type ListChannelsResponse struct {
	Channels []*types.KotsChannel `json:"channels"`
}

func (c *VendorV3Client) ListKotsChannels(appID string, channelName string, excludeDetails bool) ([]*types.KotsChannel, error) {
	var response = ListChannelsResponse{}
	v := url.Values{}
	if channelName != "" {
		v.Set("channelName", channelName)
	}
	if excludeDetails {
		v.Set("excludeDetail", "true")
	}

	url := fmt.Sprintf("/v3/app/%s/channels?%s", appID, v.Encode())
	err := c.DoJSON(context.TODO(), "GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list channels")
	}

	return response.Channels, nil
}

func (c *VendorV3Client) ListChannels(appID string, channelName string) ([]*types.Channel, error) {
	kotsChannels, err := c.ListKotsChannels(appID, channelName, true)
	if err != nil {
		return nil, err
	}

	channels := make([]*types.Channel, 0)
	for _, kotsChannel := range kotsChannels {
		channels = append(channels, kotsChannel.ToChannel())
	}

	return channels, nil
}

func (c *VendorV3Client) CreateChannel(appID, name, description string) (*types.Channel, error) {
	request := types.CreateChannelRequest{
		Name:        name,
		Description: description,
	}

	type createChannelResponse struct {
		Channel types.KotsChannel `json:"channel"`
	}
	var response createChannelResponse

	url := fmt.Sprintf("/v3/app/%s/channel", appID)
	err := c.DoJSON(context.TODO(), "POST", url, http.StatusCreated, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "create channel")
	}

	return response.Channel.ToChannel(), nil
}

func (c *VendorV3Client) GetChannel(appID string, channelID string) (*types.Channel, error) {
	type getChannelResponse struct {
		Channel types.KotsChannel `json:"channel"`
	}

	response := getChannelResponse{}
	url := fmt.Sprintf("/v3/app/%s/channel/%s", appID, url.QueryEscape(channelID))
	err := c.DoJSON(context.TODO(), "GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "get app channel")
	}

	return response.Channel.ToChannel(), nil
}

func (c *VendorV3Client) ArchiveChannel(appID, channelID string) error {
	url := fmt.Sprintf("/v3/app/%s/channel/%s", appID, url.QueryEscape(channelID))

	err := c.DoJSON(context.TODO(), "DELETE", url, http.StatusOK, nil, nil)
	if err != nil {
		return errors.Wrap(err, "archive app channel")
	}

	return nil
}

func (c *VendorV3Client) UpdateSemanticVersioning(appID string, channel *types.Channel, enableSemver bool) error {
	request := types.PatchChannelRequest{
		SemverRequired: &enableSemver,
	}

	response := struct {
		Channel types.KotsChannel `json:"channel"`
	}{}

	url := fmt.Sprintf("/v3/app/%s/channel/%s", appID, channel.ID)
	err := c.DoJSON(context.TODO(), "PATCH", url, http.StatusOK, request, &response)
	if err != nil {
		return errors.Wrap(err, "edit semantic versioning for channel")
	}

	return nil
}

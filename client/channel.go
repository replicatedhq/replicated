package client

import (
	"fmt"
	"net/http"

	channels "github.com/replicatedhq/replicated/gen/go/channels"
)

// ListChannels returns all channels for an app.
func (c *Client) ListChannels(appID string) ([]channels.AppChannel, error) {
	path := fmt.Sprintf("/v1/app/%s/channels", appID)
	appChannels := make([]channels.AppChannel, 0)
	err := c.doJSON("GET", path, http.StatusOK, nil, &appChannels)
	if err != nil {
		return nil, fmt.Errorf("ListChannels: %v", err)
	}
	return appChannels, nil
}

// CreateChannel adds a channel to an app.
func (c *Client) CreateChannel(appID, name, desc string) ([]channels.AppChannel, error) {
	path := fmt.Sprintf("/v1/app/%s/channel", appID)
	body := &channels.Body{
		Name:        name,
		Description: desc,
	}
	appChannels := make([]channels.AppChannel, 0)
	err := c.doJSON("POST", path, http.StatusOK, body, &appChannels)
	if err != nil {
		return nil, fmt.Errorf("CreateChannel: %v", err)
	}
	return appChannels, nil
}

// ArchiveChannel archives a channel.
func (c *Client) ArchiveChannel(appID, channelID string) error {
	endpoint := fmt.Sprintf("%s/v1/app/%s/channel/%s/archive", c.apiOrigin, appID, channelID)
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.apiKey)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("ArchiveChannel (%s %s): %v", req.Method, endpoint, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("ArchiveChannel (%s %s): status %d", req.Method, endpoint, resp.StatusCode)
	}
	return nil
}

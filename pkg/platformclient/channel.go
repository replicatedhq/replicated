package platformclient

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	channels "github.com/replicatedhq/replicated/gen/go/v1"
)

// AppChannels sorts []channels.AppChannel by Channel.Position
type AppChannels []channels.AppChannel

func (acs AppChannels) Len() int {
	return len(acs)
}

func (acs AppChannels) Swap(i, j int) {
	acs[i], acs[j] = acs[j], acs[i]
}

func (acs AppChannels) Less(i, j int) bool {
	return acs[i].Position < acs[j].Position
}

// ListChannels returns all channels for an app.
func (c *HTTPClient) ListChannels(appID string) ([]channels.AppChannel, error) {
	path := fmt.Sprintf("/v1/app/%s/channels", appID)
	appChannels := make([]channels.AppChannel, 0)
	err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &appChannels)
	if err != nil {
		return nil, fmt.Errorf("ListChannels: %w", err)
	}
	sort.Sort(AppChannels(appChannels))

	return appChannels, nil
}

// CreateChannel adds a channel to an app.
func (c *HTTPClient) CreateChannel(appID string, name string, description string) error {
	path := fmt.Sprintf("/v1/app/%s/channel", appID)
	body := &channels.BodyCreateChannel{
		Name:        name,
		Description: description,
	}
	appChannels := make([]channels.AppChannel, 0)
	err := c.DoJSON(context.TODO(), "POST", path, http.StatusOK, body, &appChannels)
	if err != nil {
		return fmt.Errorf("CreateChannel: %w", err)
	}
	return nil
}

// ArchiveChannel archives a channel.
func (c *HTTPClient) ArchiveChannel(appID, channelID string) error {
	endpoint := fmt.Sprintf("%s/v1/app/%s/channel/%s/archive", c.apiOrigin, appID, channelID)
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ArchiveChannel (%s %s): %w", req.Method, endpoint, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("ArchiveChannel (%s %s): status %d", req.Method, endpoint, resp.StatusCode)
	}
	return nil
}

// ChannelReleases sorts []channels.ChannelRelease newest to oldest
type ChannelReleases []channels.ChannelRelease

func (crs ChannelReleases) Len() int {
	return len(crs)
}

func (crs ChannelReleases) Swap(i, j int) {
	crs[i], crs[j] = crs[j], crs[i]
}

func (crs ChannelReleases) Less(i, j int) bool {
	return crs[i].ChannelSequence > crs[j].ChannelSequence
}

// GetChannel returns channel details and release history
func (c *HTTPClient) GetChannel(appID, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error) {
	path := fmt.Sprintf("/v1/app/%s/channel/%s/releases", appID, channelID)
	respBody := channels.GetChannelInlineResponse200{}
	err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &respBody)
	if err != nil {
		return nil, nil, fmt.Errorf("GetChannel: %w", err)
	}
	sort.Sort(ChannelReleases(respBody.Releases))
	return respBody.Channel, respBody.Releases, nil
}

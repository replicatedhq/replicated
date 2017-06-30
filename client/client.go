// Package client manages channels and releases through the Replicated Vendor API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	channels "github.com/replicatedhq/replicated/gen/go/channels"
	releases "github.com/replicatedhq/replicated/gen/go/releases"
)

type Client interface {
	ListChannels(appID string) ([]channels.AppChannel, error)
	CreateChannel(appID, name, desc string) ([]channels.AppChannel, error)
	ArchiveChannel(appID, channelID string) error

	ListReleases(appID string) ([]releases.AppReleaseInfo, error)
	CreateRelease(appID string) (*releases.AppReleaseInfo, error)
	UpdateRelease(appID string, sequence int64, yaml string) error
	GetRelease(appID string, sequence int64) (*releases.AppReleaseInfo, error)
	PromoteRelease(
		appID string,
		sequence int64,
		label string,
		notes string,
		required bool,
		channelIDs ...string) error
}

// A Client communicates with the Replicated Vendor HTTP API.
type HTTPClient struct {
	apiKey    string
	apiOrigin string
}

// New returns a new  HTTP client.
func New(origin string, apiKey string) Client {
	c := &HTTPClient{
		apiKey:    apiKey,
		apiOrigin: origin,
	}

	return c
}

func (c *HTTPClient) doJSON(method, path string, successStatus int, reqBody, respBody interface{}) error {
	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)
	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return fmt.Errorf("%s %s: %v", method, endpoint, err)
		}
	}
	req, err := http.NewRequest(method, endpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %v", method, endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != successStatus {
		return fmt.Errorf("%s %s: status %d", method, endpoint, resp.StatusCode)
	}
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("%s %s response decoding: %v", method, endpoint, err)
		}
	}
	return nil
}

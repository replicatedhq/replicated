package client

import (
	"fmt"
	"net/http"
	"strings"

	releases "github.com/replicatedhq/replicated/gen/go/releases"
)

// ListReleases lists all releases for an app.
func (c *HTTPClient) ListReleases(appID string) ([]releases.AppReleaseInfo, error) {
	path := fmt.Sprintf("/v1/app/%s/releases", appID)
	releases := make([]releases.AppReleaseInfo, 0)
	if err := c.doJSON("GET", path, http.StatusOK, nil, &releases); err != nil {
		return nil, fmt.Errorf("ListReleases: %v", err)
	}
	return releases, nil
}

// CreateRelease adds a release to an app.
func (c *HTTPClient) CreateRelease(appID string) (*releases.AppReleaseInfo, error) {
	path := fmt.Sprintf("/v1/app/%s/release", appID)
	body := &releases.Body{
		Source: "latest",
	}
	release := &releases.AppReleaseInfo{}
	if err := c.doJSON("POST", path, http.StatusCreated, body, release); err != nil {
		return nil, fmt.Errorf("CreateRelease: %v", err)
	}
	return release, nil
}

// UpdateRelease updates a release's yaml.
func (c *HTTPClient) UpdateRelease(appID string, sequence int64, yaml string) error {
	endpoint := fmt.Sprintf("/v1/app/%s/%d/raw", appID, sequence)
	req, err := http.NewRequest("PUT", endpoint, strings.NewReader(yaml))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/yaml")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateRelease (%s %s): %v", req.Method, endpoint, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("UpdateRelease (%s %s): status %d", req.Method, endpoint, resp.StatusCode)
	}
	return nil
}

// GetRelease returns a release's properties.
func (c *HTTPClient) GetRelease(appID string, sequence int64) (*releases.AppReleaseInfo, error) {
	path := fmt.Sprintf("%s/v1/app/%s/release/%d/properties", c.apiOrigin, appID, sequence)
	release := &releases.AppReleaseInfo{}
	if err := c.doJSON("GET", path, http.StatusOK, nil, release); err != nil {
		return nil, fmt.Errorf("GetRelease: %v", err)
	}
	return release, nil
}

// PromoteRelease points the specified channels at a release sequence.
func (c *HTTPClient) PromoteRelease(appID string, sequence int64, label, notes string, required bool, channelIDs ...string) error {
	path := fmt.Sprintf("/v1/app/%s/%d/promote", appID, sequence)
	body := &releases.Body1{
		Label:        label,
		ReleaseNotes: notes,
		Required:     required,
		Channels:     channelIDs,
	}
	if err := c.doJSON("POST", path, http.StatusNoContent, body, nil); err != nil {
		return fmt.Errorf("PromoteRelease: %v", err)
	}
	return nil
}

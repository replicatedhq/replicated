package platformclient

import (
	"fmt"
	"net/http"
	"strings"

	releases "github.com/replicatedhq/replicated/gen/go/v1"
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
func (c *HTTPClient) CreateRelease(appID string, yaml string) (*releases.AppReleaseInfo, error) {
	path := fmt.Sprintf("/v1/app/%s/release", appID)
	body := &releases.BodyCreateRelease{
		Source: "latest",
	}
	release := &releases.AppReleaseInfo{}
	if err := c.doJSON("POST", path, http.StatusCreated, body, release); err != nil {
		return nil, fmt.Errorf("CreateRelease: %v", err)
	}
	// API does not accept yaml in create operation, so first create then udpate
	if yaml != "" {
		if err := c.UpdateRelease(appID, release.Sequence, yaml); err != nil {
			return nil, fmt.Errorf("CreateRelease with YAML: %v", err)
		}
	}
	return release, nil
}

// UpdateRelease updates a release's yaml.
func (c *HTTPClient) UpdateRelease(appID string, sequence int64, yaml string) error {
	endpoint := fmt.Sprintf("%s/v1/app/%s/%d/raw", c.apiOrigin, appID, sequence)
	req, err := http.NewRequest("PUT", endpoint, strings.NewReader(yaml))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/yaml")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateRelease: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if badRequestErr, err := unmarshalBadRequest(resp.Body); err == nil {
			return badRequestErr
		}
		return fmt.Errorf("UpdateRelease (%s %s): status %d", req.Method, endpoint, resp.StatusCode)
	}
	return nil
}

// GetRelease returns a release's properties.
func (c *HTTPClient) GetRelease(appID string, sequence int64) (*releases.AppRelease, error) {
	path := fmt.Sprintf("/v1/app/%s/%d/properties", appID, sequence)
	release := &releases.AppRelease{}
	if err := c.doJSON("GET", path, http.StatusOK, nil, release); err != nil {
		return nil, fmt.Errorf("GetRelease: %v", err)
	}
	return release, nil
}

// PromoteRelease points the specified channels at a release sequence.
func (c *HTTPClient) PromoteRelease(appID string, sequence int64, label, notes string, required bool, channelIDs ...string) error {
	path := fmt.Sprintf("/v1/app/%s/%d/promote", appID, sequence)
	body := &releases.BodyPromoteRelease{
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

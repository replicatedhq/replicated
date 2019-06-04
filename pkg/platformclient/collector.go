package platformclient

import (
	"fmt"
	"net/http"
	"strings"

	collectors "github.com/replicatedhq/replicated/gen/go/v1"
)

// ListCollectors lists all collectors for an app.
func (c *HTTPClient) ListCollectors(appID string) ([]collectors.AppCollectorInfo, error) {
	path := fmt.Sprintf("/v1/app/%s/collectors", appID)
	collectors := make([]collectors.AppCollectorInfo, 0)
	if err := c.doJSON("GET", path, http.StatusOK, nil, &collectors); err != nil {
		return nil, fmt.Errorf("ListCollectors: %v", err)
	}
	return collectors, nil
}

// CreateCollector adds a release to an app.
func (c *HTTPClient) CreateCollector(appID string, name string, yaml string) (*collectors.AppCollectorInfo, error) {
	path := fmt.Sprintf("/v1/app/%s/collector/%d", appID, name)
	body := &collectors.BodyCreateCollector{
		Source: "latest",
	}
	collector := &collectors.AppCollectorInfo{}
	if err := c.doJSON("POST", path, http.StatusCreated, body, collector); err != nil {
		return nil, fmt.Errorf("CreateCollector: %v", err)
	}
	// API does not accept yaml in create operation, so first create then udpate
	if yaml != "" {
		if err := c.UpdateCollector(appID, collector.Name, yaml); err != nil {
			return nil, fmt.Errorf("CreateCollector with YAML: %v", err)
		}
	}
	return collector, nil
}

// UpdateCollector updates a collector's yaml.
func (c *HTTPClient) UpdateCollector(appID string, name string, yaml string) error {
	endpoint := fmt.Sprintf("%s/v1/app/%s/collectors/%d/raw", c.apiOrigin, appID, name)
	req, err := http.NewRequest("PUT", endpoint, strings.NewReader(yaml))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/yaml")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateCollector: %v", err)
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

// GetCollector returns a collector's properties.
func (c *HTTPClient) GetCollector(appID string, name string) (*collectors.AppCollector, error) {
	path := fmt.Sprintf("/v1/app/%s/collectors/%d/properties", appID, name)
	collector := &collectors.AppCollector{}
	if err := c.doJSON("GET", path, http.StatusOK, nil, collector); err != nil {
		return nil, fmt.Errorf("GetCollector: %v", err)
	}
	return collector, nil
}

// PromoteCollector points the specified channels at a named collector.
func (c *HTTPClient) PromoteCollector(appID string, name string, channelIDs ...string) error {
	path := fmt.Sprintf("/v1/app/%s/collectors/%d/promote?dry_run=true", appID, name)
	body := &collectors.BodyPromoteCollector{
		Channels: channelIDs,
	}
	if err := c.doJSON("POST", path, http.StatusNoContent, body, nil); err != nil {
		return fmt.Errorf("PromoteCollector: %v", err)
	}
	return nil
}

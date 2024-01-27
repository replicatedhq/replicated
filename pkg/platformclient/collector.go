package platformclient

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *HTTPClient) ListCollectors(appID string, appType string) ([]types.CollectorSpec, error) {
	if appType != "platform" {
		return nil, errors.Errorf("unknown app type %s", appType)
	}

	params := url.Values{}
	params.Add("appId", appID)

	collectors := struct {
		Specs []types.CollectorSpec `json:"specs"`
	}{}

	url := fmt.Sprintf("/v1/collector/specs?%s", params.Encode())
	if err := c.DoJSON("GET", url, http.StatusOK, nil, &collectors); err != nil {
		return nil, fmt.Errorf("list specs: %w", err)
	}

	return collectors.Specs, nil
}

func (c *HTTPClient) GetCollector(appID string, specID string) (*types.CollectorSpec, error) {
	collector := struct {
		Spec types.CollectorSpec `json:"spec"`
	}{}

	if err := c.DoJSON("GET", fmt.Sprintf("/v1/collector/spec/%s", specID), http.StatusOK, nil, &collector); err != nil {
		return nil, fmt.Errorf("get collector: %w", err)
	}

	return &collector.Spec, nil
}

// Vendor-API: PromoteCollector points the specified channels at a named collector.
func (c *HTTPClient) PromoteCollector(appID string, specID string, channelIDs ...string) error {
	path := fmt.Sprintf("/v1/app/%s/collector/%s/promote", appID, specID)
	body := &v1.BodyPromoteCollector{
		ChannelIDs: channelIDs,
	}
	if err := c.DoJSON("POST", path, http.StatusOK, body, nil); err != nil {
		return fmt.Errorf("PromoteCollector: %w", err)
	}
	return nil
}

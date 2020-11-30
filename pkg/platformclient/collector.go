package platformclient

import (
	"fmt"
	"net/http"

	v1 "github.com/replicatedhq/replicated/gen/go/v1"
)

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

package kotsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) GetNetworkReport(id string) (*types.NetworkReport, error) {
	return c.GetNetworkReportAfter(id, nil)
}

func (c *VendorV3Client) GetNetworkReportAfter(id string, after *time.Time) (*types.NetworkReport, error) {
	urlPath := fmt.Sprintf("/v3/network/%s/report", id)
	if after != nil {
		v := url.Values{}
		v.Set("after", after.Format(time.RFC3339Nano))
		urlPath = fmt.Sprintf("%s?%s", urlPath, v.Encode())
	}

	// Get raw response as map
	var rawResponse map[string]interface{}
	err := c.DoJSON(context.TODO(), "GET", urlPath, http.StatusOK, nil, &rawResponse)
	if err != nil {
		return nil, err
	}

	// Extract events array
	eventsRaw, ok := rawResponse["events"].([]interface{})
	if !ok {
		return &types.NetworkReport{Events: []*types.NetworkEventData{}}, nil
	}

	// Parse each event using json.Unmarshal
	var events []*types.NetworkEventData
	for _, eventRaw := range eventsRaw {
		// Convert to JSON bytes
		eventBytes, err := json.Marshal(eventRaw)
		if err != nil {
			continue // Skip malformed events
		}

		// Unmarshal into NetworkEventData
		var eventData types.NetworkEventData
		if err := json.Unmarshal(eventBytes, &eventData); err != nil {
			continue // Skip malformed events
		}

		events = append(events, &eventData)
	}

	return &types.NetworkReport{Events: events}, nil
}

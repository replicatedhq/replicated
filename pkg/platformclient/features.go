package platformclient

import (
	"context"
	"net/http"
)

// Feature represents a single feature flag from the vendor API
type Feature struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// FeaturesResponse represents the response from the /v1/user/features endpoint
type FeaturesResponse struct {
	FutureFeatures []Feature `json:"futureFeatures"`
	Features       []Feature `json:"features"`
}

// GetFeatures fetches the feature flags for the authenticated user
func (c *HTTPClient) GetFeatures(ctx context.Context) (*FeaturesResponse, error) {
	var resp FeaturesResponse
	if err := c.DoJSON(ctx, "GET", "/v1/user/features", http.StatusOK, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFeatureValue returns the value of a specific feature flag, or an empty string if not found
func (fr *FeaturesResponse) GetFeatureValue(key string) string {
	for _, feature := range fr.Features {
		if feature.Key == key {
			return feature.Value
		}
	}
	for _, feature := range fr.FutureFeatures {
		if feature.Key == key {
			return feature.Value
		}
	}
	return ""
}

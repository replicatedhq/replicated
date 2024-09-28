package types

import "time"

type VM struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Distribution string       `json:"distribution"`
	Version      string       `json:"version"`
	NodeGroups   []*NodeGroup `json:"node_groups"`

	Status    ClusterStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`

	TTL string `json:"ttl"`

	CreditsPerHourPerVM int64 `json:"credits_per_hour_per_vm"`
	FlatFee             int64 `json:"flat_fee"`
	TotalCredits        int64 `json:"total_credits"`

	EstimatedCost int64 `json:"estimated_cost"` // Represents estimated credits for this vm based on the TTL

	OverlayEndpoint string `json:"overlay_endpoint,omitempty"`
	OverlayToken    string `json:"overlay_token,omitempty"`

	Tags []Tag `json:"tags"`
}

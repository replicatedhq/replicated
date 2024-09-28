package types

import "time"

type VM struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Distribution string `json:"distribution"`
	Version      string `json:"version"`

	Status    ClusterStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`

	OverlayEndpoint string `json:"overlay_endpoint,omitempty"`
	OverlayToken    string `json:"overlay_token,omitempty"`

	Tags []Tag `json:"tags"`
}

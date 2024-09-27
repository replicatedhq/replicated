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
}

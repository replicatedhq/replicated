package types

import "time"

type Cluster struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Distribution string    `json:"distribution"`
	Version      string    `json:"version"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

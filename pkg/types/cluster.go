package types

import "time"

type Cluster struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Distribution string    `json:"distribution"`
	Version      string    `json:"version"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expire_at"`
}

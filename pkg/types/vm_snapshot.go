package types

import "time"

type VMSnapshot struct {
	ID        string     `json:"id"`
	VMID      string     `json:"resource_id"`
	ShortID   string     `json:"short_id"`
	TeamID    string     `json:"team_id"`
	Status    string     `json:"status"`
	SizeBytes *int64     `json:"size_bytes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ReadyAt   *time.Time `json:"ready_at,omitempty"`
	Error     string     `json:"error,omitempty"`
}

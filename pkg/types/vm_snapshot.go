package types

import "time"

type VMSnapshot struct {
	ID        string     `json:"id"`
	VMID      string     `json:"resource_id"`
	ShortID   string     `json:"short_id"`
	Name      string     `json:"name"`
	TeamID    string     `json:"team_id"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	ReadyAt   *time.Time `json:"ready_at,omitempty"`
}

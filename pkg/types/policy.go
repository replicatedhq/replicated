package types

import "time"

// Policy represents a team RBAC policy as returned by the vendor-api.
type Policy struct {
	ID          string     `json:"id"`
	TeamID      string     `json:"teamId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Definition  string     `json:"definition"`
	CreatedAt   time.Time  `json:"createdAt"`
	ModifiedAt  *time.Time `json:"modifiedAt"`
	ReadOnly    bool       `json:"readOnly"`
}

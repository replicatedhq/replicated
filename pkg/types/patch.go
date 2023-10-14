package types

import "time"

type Patch struct {
	ID          string     `json:"id"`
	AppID       string     `json:"appId,omitempty"`
	BundleID    string     `json:"bundleId,omitempty"`
	Description string     `json:"description"`
	Patch       string     `json:"patch"`
	CreatedAt   time.Time  `json:"createdAt"`
	AppliedAt   *time.Time `json:"appliedAt,omitempty"`
	RejectedAt  *time.Time `json:"rejectedAt,omitempty"`
}

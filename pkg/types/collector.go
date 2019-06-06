package types

import "time"

type CollectorInfo struct {
	ActiveChannels []Channel
	AppID          string
	CreatedAt      time.Time
	EditedAt       time.Time
	Name           string
	SpecID         string
	Config         string
}

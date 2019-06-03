package types

import "time"

type CollectorInfo struct {
	ActiveChannels []Channel
	AppID          string
	CreatedAt      time.Time
	Name           string
}

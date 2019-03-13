package types

import "time"

type ReleaseInfo struct {
	ActiveChannels []Channel
	AppID          string
	CreatedAt      time.Time
	EditedAt       time.Time
	Editable       bool
	Sequence       int64
	Version        string
}

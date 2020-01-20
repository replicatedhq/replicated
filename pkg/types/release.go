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

type LintMessage struct {
	Rule      string          `json:"rule"`
	Type      string          `json:"type"`
	Message   string          `json:"message"`
	Positions []*LintPosition `json:"positions"`
}

type LintPosition struct {
	Path  string            `json:"path"`
	Start LintLinePosition `json:"start"`
	End   LintLinePosition `json:"end"`
}

type LintLinePosition struct {
	Position int64 `json:"position"`
	Line     int64 `json:"line"`
	Column   int64 `json:"column"`
}

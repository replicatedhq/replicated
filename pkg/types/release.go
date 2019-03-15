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
	Rule      string
	Type      string
	Positions []*LintPosition
}

type LintPosition struct {
	Path  string
	Start *LintLinePosition
	End   *LintLinePosition
}

type LintLinePosition struct {
	Position int64
	Line     int64
	Column   int64
}

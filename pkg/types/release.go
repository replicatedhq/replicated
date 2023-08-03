package types

import "time"

type ChartStatus string

var (
	ChartStatusUnknown ChartStatus = "unknown"
	ChartStatusPushing ChartStatus = "pushing"
	ChartStatusPushed  ChartStatus = "pushed"
	ChartStatusError   ChartStatus = "error"
)

type ReleaseInfo struct {
	ActiveChannels []Channel
	AppID          string
	CreatedAt      time.Time
	EditedAt       time.Time
	Editable       bool
	Sequence       int64
	Version        string
	Charts         []Chart
}

type LintMessage struct {
	Rule      string          `json:"rule"`
	Type      string          `json:"type"`
	Path      string          `json:"path"`
	Message   string          `json:"message"`
	Positions []*LintPosition `json:"positions"`
}

type LintPosition struct {
	Path  string           `json:"path"`
	Start LintLinePosition `json:"start"`
	End   LintLinePosition `json:"end"`
}

type LintLinePosition struct {
	Position int64 `json:"position"`
	Line     int64 `json:"line"`
	Column   int64 `json:"column"`
}

type KotsCreateReleaseRequest struct {
	SpecGzip []byte `json:"spec_gzip"`
}

type KotsGetReleaseResponse struct {
	Release KotsAppRelease `json:"release"`
}

type KotsTestReleaseResponse struct {
}

type KotsUpdateReleaseRequest struct {
	SpecGzip []byte `json:"spec_gzip"`
}

// KotsListReleasesResponse contains the JSON releases list
type KotsListReleasesResponse struct {
	Releases []*KotsAppRelease `json:"releases"`
}

type KotsPromoteReleaseRequest struct {
	ReleaseNotes   string   `json:"releaseNotes"`
	VersionLabel   string   `json:"versionLabel"`
	IsRequired     bool     `json:"isRequired"`
	ChannelIDs     []string `json:"channelIds"`
	IgnoreWarnings bool     `json:"ignoreWarnings"`
}

type KotsAppRelease struct {
	AppID                string     `json:"appId"`
	Sequence             int64      `json:"sequence"`
	CreatedAt            time.Time  `json:"created"`
	IsArchived           bool       `json:"isArchived"`
	Spec                 string     `json:"spec"`
	ReleaseNotes         string     `json:"releaseNotes"`
	IsReleaseNotEditable bool       `json:"isReleaseNotEditable"`
	Channels             []*Channel `json:"channels"`
	Charts               []Chart    `json:"charts"`
}

type Chart struct {
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	Status    ChartStatus `json:"status"`
	Error     string      `json:"error,omitempty"`
	UpdatedAt *time.Time  `json:"updatedAt,omitempty"`
}

type EntitlementValue struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type AppRelease struct {
	Config    string    `json:"config,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	Editable  bool      `json:"editable,omitempty"`
	EditedAt  time.Time `json:"editedAt,omitempty"`
	Sequence  int64     `json:"sequence,omitempty"`
	Charts    []Chart   `json:"charts,omitempty"`
}

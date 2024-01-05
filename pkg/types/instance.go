package types

import (
	"strings"
	"time"
)

type Instance struct {
	LicenseID      string           `json:"licenseId,omitempty"`
	InstanceID     string           `json:"instanceId,omitempty"`
	ClusterID      string           `json:"clusterId,omitempty"`
	CreatedAt      time.Time        `json:"createdAt,omitempty"`
	LastActive     time.Time        `json:"lastActive,omitempty"`
	AppStatus      string           `json:"appStatus,omitempty"`
	Active         bool             `json:"active,omitempty"`
	VersionHistory []VersionHistory `json:"versionHistory,omitempty"`
	Tags           Tags             `json:"tags,omitempty"` // must be Tags type for template evaluation
}

// Used for template evaluation
func (i Instance) LatestVersion() string {
	if len(i.VersionHistory) == 0 {
		return ""
	}
	// API returns a sorted list with latest being first
	return i.VersionHistory[0].VersionLabel
}

// Used for template evaluation
func (i Instance) Name() string {
	for _, tag := range i.Tags {
		if strings.ToLower(tag.Key) == "name" {
			return tag.Value
		}
	}
	return ""
}

type VersionHistory struct {
	InstanceID                string    `json:"instanceId,omitempty"`
	ClusterID                 string    `json:"clusterId,omitempty"`
	VersionLabel              string    `json:"versionLabel,omitempty"`
	DownStreamChannelID       string    `json:"downstreamChannelId,omitempty"`
	DownStreamReleaseSequence int32     `json:"downstreamReleaseSequence,omitempty"`
	IntervalStart             time.Time `json:"intervalStart,omitempty"`
	IntervalLast              time.Time `json:"intervallast,omitempty"`
	RepHelmCount              int32     `json:"repHelmCount,omitempty"`
	NativeHelmCount           int32     `json:"nativeHelmCount,omitempty"`
}

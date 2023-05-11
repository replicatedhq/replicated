package types

import "time"

type Instance struct {
	LicenseId      string           `json:"licenseId,omitempty"`
	InstanceId     string           `json:"instanceId,omitempty"`
	ClusterId      string           `json:"clusterId,omitempty"`
	CreatedAt      time.Time        `json:"createdAt,omitempty"`
	LastActive     time.Time        `json:"lastActive,omitempty"`
	AppStatus      string           `json:"appStatus,omitempty"`
	Active         bool             `json:"active,omitempty"`
	VersionHistory []VersionHistory `json:"versionHistory,omitempty"`
}

type VersionHistory struct {
	InstanceId                string    `json:"instanceId,omitempty"`
	ClusterId                 string    `json:"clusterId,omitempty"`
	VersionLabel              string    `json:"versionLabel,omitempty"`
	DownStreamChannelId       string    `json:"downstreamChannelId,omitempty"`
	DownStreamReleaseSequence int32     `json:"downstreamReleaseSequence,omitempty"`
	IntervalStart             time.Time `json:"intervalStart,omitempty"`
	IntervalLast              time.Time `json:"intervallast,omitempty"`
	RepHelmCount              int32     `json:"repHelmCount,omitempty"`
	NativeHelmCount           int32     `json:"nativeHelmCount,omitempty"`
}

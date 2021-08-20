package types

import (
	"github.com/replicatedhq/replicated/pkg/util"
)

type InstallerSpec struct {
	AppID           string    `json:"appId"`
	KurlInstallerID string    `json:"kurlInstallerID"`
	Sequence        int64     `json:"sequence"`
	YAML            string    `json:"yaml"`
	ActiveChannels  []Channel `json:"channels"`
	CreatedAt       util.Time `json:"created"`
	Immutable       bool      `json:"isInstallerNotEditable"`
}

type InstallerSpecResponse struct {
	Body InstallerSpec `json:"installer"`
}

type CreateInstallerRequest struct {
	Yaml string `json:"yaml"`
}

type PromoteInstallerRequest struct {
	Sequence     int64  `json:"sequence"`
	VersionLabel string `json:"versionLabel"`
	ChannelID    string `json:"channelId"`
}

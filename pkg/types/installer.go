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

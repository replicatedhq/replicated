package types

import (
	"time"

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

type EnterpriseChannel struct {
	Description    string               `json:"description,omitempty"`
	EnterpriseName string               `json:"enterprise_name,omitempty"`
	Id             string               `json:"id,omitempty"`
	Installer      *EnterpriseInstaller `json:"installer,omitempty"`
	Name           string               `json:"name,omitempty"`
	Policies       []EnterprisePolicy   `json:"policies,omitempty"`
}

type EnterpriseInstaller struct {
	Id        string `json:"id,omitempty"`
	PartnerId string `json:"partner_id,omitempty"`
	Yaml      string `json:"yaml,omitempty"`
}

type EnterprisePolicy struct {
	CreatedAt   time.Time `json:"created_at,omitempty"`
	Description string    `json:"description,omitempty"`
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	PartnerId   string    `json:"partner_id,omitempty"`
	Policy      string    `json:"policy,omitempty"`
}

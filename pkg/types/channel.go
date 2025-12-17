package types

import (
	"time"
)

type KotsChannel struct {
	AdoptionRate             []CustomerAdoption            `json:"adoptionRate,omitempty"`
	AppId                    string                        `json:"appId,omitempty"`
	BuildAirgapAutomatically bool                          `json:"buildAirgapAutomatically,omitempty"`
	ChannelIcon              string                        `json:"channelIcon,omitempty"`
	ChannelSequence          int32                         `json:"channelSequence,omitempty"`
	ChannelSlug              string                        `json:"channelSlug,omitempty"`
	Created                  time.Time                     `json:"created,omitempty"`
	CurrentVersion           string                        `json:"currentVersion,omitempty"`
	Customers                *TotalActiveInactiveCustomers `json:"customers,omitempty"`
	Description              string                        `json:"description,omitempty"`
	Id                       string                        `json:"id,omitempty"`
	IsArchived               bool                          `json:"isArchived,omitempty"`
	IsDefault                bool                          `json:"isDefault,omitempty"`
	Name                     string                        `json:"name,omitempty"`
	NumReleases              int32                         `json:"numReleases,omitempty"`
	IsHelmOnly               bool                          `json:"isHelmOnly,omitempty"`
	ReleaseNotes             string                        `json:"releaseNotes,omitempty"`
	// TODO: set these (see kotsChannelToSchema function)
	ReleaseSequence          int32                   `json:"releaseSequence,omitempty"`
	Releases                 []ChannelRelease        `json:"releases,omitempty"`
	Updated                  time.Time               `json:"updated,omitempty"`
	ReplicatedRegistryDomain string                  `json:"replicatedRegistryDomain"`
	CustomHostNameOverrides  CustomHostNameOverrides `json:"customHostNameOverrides"`
	ChartReleases            []ChartRelease          `json:"chartReleases"`
}

func (c *KotsChannel) ToChannel() *Channel {
	return &Channel{
		ID:              c.Id,
		Name:            c.Name,
		Description:     c.Description,
		Slug:            c.ChannelSlug,
		ReleaseSequence: int64(c.ReleaseSequence),
		ChannelSequence: int64(c.ChannelSequence),
		ReleaseLabel:    c.CurrentVersion,
		IsArchived:      c.IsArchived,
		IsHelmOnly:      c.IsHelmOnly,
	}
}

type ChartRelease struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Weight           int    `json:"weight"`
	Error            string `json:"error"`
	HasPreflightSpec bool   `json:"hasPreflightSpec"`
}

type CustomerAdoption struct {
	ChannelId       string  `json:"channelId,omitempty"`
	Count           int32   `json:"count,omitempty"`
	Percent         float32 `json:"percent,omitempty"`
	ReleaseSequence int32   `json:"releaseSequence,omitempty"`
	Semver          string  `json:"semver,omitempty"`
	TotalOnChannel  int64   `json:"totalOnChannel,omitempty"`
}

type ChannelRelease struct {
	AirgapBuildError    string            `json:"airgapBuildError,omitempty"`
	AirgapBuildStatus   string            `json:"airgapBuildStatus,omitempty"`
	AirgapBundleImages  []string          `json:"airgapBundleImages,omitempty"`
	ChannelIcon         string            `json:"channelIcon,omitempty"`
	ChannelId           string            `json:"channelId,omitempty"`
	ChannelName         string            `json:"channelName,omitempty"`
	ChannelSequence     int32             `json:"channelSequence,omitempty"`
	Created             time.Time         `json:"created,omitempty"`
	ProxyRegistryDomain string            `json:"proxyRegistryDomain,omitempty"`
	RegistrySecret      string            `json:"registrySecret,omitempty"`
	ReleaseNotes        string            `json:"releaseNotes,omitempty"`
	ReleasedAt          time.Time         `json:"releasedAt,omitempty"`
	Semver              string            `json:"semver,omitempty"`
	Sequence            int32             `json:"sequence,omitempty"`
	Updated             time.Time         `json:"updated,omitempty"`
	InstallationTypes   InstallationTypes `json:"installationTypes,omitempty"`
}

type CreateChannelRequest struct {
	// Description of the channel that is to be created.
	Description string `json:"description,omitempty"`
	Name        string `json:"name"`
}

type PatchChannelRequest struct {
	SemverRequired *bool `json:"semverRequired,omitempty"`
}

type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"channelSlug"`

	ReleaseSequence int64  `json:"releaseSequence"`
	ReleaseLabel    string `json:"releaseLabel"`

	ChannelSequence int64 `json:"channelSequence"`

	IsArchived bool `json:"isArchived"`
	IsHelmOnly bool `json:"isHelmOnly"`
}

type CustomHostNameOverrides struct {
	Registry struct {
		Hostname string `json:"hostname"`
	} `json:"registry"`

	Proxy struct {
		Hostname string `json:"hostname"`
	} `json:"proxy"`

	DownloadPortal struct {
		Hostname string `json:"hostname"`
	} `json:"downloadPortal"`

	ReplicatedApp struct {
		Hostname string `json:"hostname"`
	} `json:"replicatedApp"`
}

type InstallationTypes struct {
	EmbeddedCluster EmbeddedCluster `json:"embeddedCluster,omitempty"`
}

type EmbeddedCluster struct {
	Version             string `json:"version,omitempty"`
	ReplicatedAppDomain string `json:"replicatedAppDomain,omitempty"`
	ProxyRegistryDomain string `json:"proxyRegistryDomain,omitempty"`
}

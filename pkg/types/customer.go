package types

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/util"
)

type Customer struct {
	ID                                 string        `json:"id"`
	CustomID                           string        `json:"customId"`
	Name                               string        `json:"name"`
	Email                              string        `json:"email"`
	Channels                           []Channel     `json:"channels"`
	Type                               string        `json:"type"`
	Expires                            *util.Time    `json:"expiresAt"`
	Instances                          []Instance    `json:"instances"`
	InstallationID                     string        `json:"installationId"`
	Entitlements                       []Entitlement `json:"entitlements"`
	IsAirgapEnabled                    bool          `json:"airgap"`
	IsEmbeddedClusterDownloadEnabled   bool          `json:"isEmbeddedClusterDownloadEnabled"`
	IsEmbeddedClusterMultinodeDisabled bool          `json:"isEmbeddedClusterMultinodeDisabled"`
	IsGeoaxisSupported                 bool          `json:"isGeoaxisSupported"`
	IsHelmVMDownloadEnabled            bool          `json:"isHelmVmDownloadEnabled"`
	IsIdentityServiceSupported         bool          `json:"isIdentityServiceSupported"`
	IsInstallerSupportEnabled          bool          `json:"isInstallerSupportEnabled"`
	IsKotsInstallEnabled               bool          `json:"isKotsInstallEnabled"`
	IsSnapshotSupported                bool          `json:"isSnapshotSupported"`
	IsSupportBundleUploadEnabled       bool          `json:"isSupportBundleUploadEnabled"`
	IsGitopsSupported                  bool          `json:"isGitopsSupported"`
}

func (c Customer) WithExpiryTime(expiryTime string) (Customer, error) {
	if expiryTime != "" {
		parsed, err := util.ParseTime(expiryTime)
		if err != nil {
			return Customer{}, errors.Wrapf(err, "parse expiresAt timestamp %q", expiryTime)
		}
		c.Expires = &util.Time{Time: parsed}
	}
	return c, nil
}

type TotalActiveInactiveCustomers struct {
	ActiveCustomers   int64 `json:"activeCustomers,omitempty"`
	InactiveCustomers int64 `json:"inactiveCustomers,omitempty"`
	TotalCustomers    int64 `json:"totalCustomers,omitempty"`
}

package kotsclient

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type EntitlementValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CustomerChannel struct {
	ID                    string `json:"channel_id"`
	PinnedChannelSequence *int64 `json:"pinned_channel_sequence"`
	IsDefault             bool   `json:"is_default_for_customer"`
}

type CreateCustomerRequest struct {
	Name                             string             `json:"name"`
	Channels                         []CustomerChannel  `json:"channels"`
	CustomID                         string             `json:"custom_id"`
	AppID                            string             `json:"app_id"`
	Type                             string             `json:"type"`
	ExpiresAt                        string             `json:"expires_at"`
	IsAirgapEnabled                  bool               `json:"is_airgap_enabled"`
	IsGitopsSupported                bool               `json:"is_gitops_supported"`
	IsSnapshotSupported              bool               `json:"is_snapshot_supported"`
	IsKotsInstallEnabled             bool               `json:"is_kots_install_enabled"`
	IsEmbeddedClusterDownloadEnabled bool               `json:"is_embedded_cluster_download_enabled"`
	IsGeoaxisSupported               bool               `json:"is_geoaxis_supported"`
	IsHelmVMDownloadEnabled          bool               `json:"is_helm_vm_download_enabled"`
	IsIdentityServiceSupported       bool               `json:"is_identity_service_supported"`
	IsInstallerSupportEnabled        bool               `json:"is_installer_support_enabled"`
	IsSupportBundleUploadEnabled     bool               `json:"is_support_bundle_upload_enabled"`
	IsDeveloperModeEnabled           bool               `json:"is_dev_mode_enabled"`
	Email                            string             `json:"email,omitempty"`
	EntitlementValues                []EntitlementValue `json:"entitlementValues"`

	// These fields were added after the "built in" fields feature was released.
	// If they are not pointer types, they will override the defaults.
	IsHelmInstallEnabled *bool `json:"is_helm_install_enabled,omitempty"`
	IsKurlInstallEnabled *bool `json:"is_kurl_install_enabled,omitempty"`
}

type CreateCustomerResponse struct {
	Customer *types.Customer `json:"customer"`
}

type CreateCustomerOpts struct {
	Name                             string
	CustomID                         string
	Channels                         []CustomerChannel
	AppID                            string
	ExpiresAt                        string
	ExpiresAtDuration                time.Duration
	IsAirgapEnabled                  bool
	IsGitopsSupported                bool
	IsSnapshotSupported              bool
	IsKotsInstallEnabled             bool
	IsHelmInstallEnabled             *bool
	IsKurlInstallEnabled             *bool
	IsEmbeddedClusterDownloadEnabled bool
	IsGeoaxisSupported               bool
	IsHelmVMDownloadEnabled          bool
	IsIdentityServiceSupported       bool
	IsInstallerSupportEnabled        bool
	IsSupportBundleUploadEnabled     bool
	IsDeveloperModeEnabled           bool
	LicenseType                      string
	Email                            string
	EntitlementValues                []EntitlementValue
}

func (c *VendorV3Client) CreateCustomer(opts CreateCustomerOpts) (*types.Customer, error) {
	request := &CreateCustomerRequest{
		Name:                             opts.Name,
		CustomID:                         opts.CustomID,
		Channels:                         opts.Channels,
		AppID:                            opts.AppID,
		Type:                             opts.LicenseType,
		IsAirgapEnabled:                  opts.IsAirgapEnabled,
		IsGitopsSupported:                opts.IsGitopsSupported,
		IsSnapshotSupported:              opts.IsSnapshotSupported,
		IsKotsInstallEnabled:             opts.IsKotsInstallEnabled,
		IsHelmInstallEnabled:             opts.IsHelmInstallEnabled,
		IsKurlInstallEnabled:             opts.IsKurlInstallEnabled,
		IsEmbeddedClusterDownloadEnabled: opts.IsEmbeddedClusterDownloadEnabled,
		IsGeoaxisSupported:               opts.IsGeoaxisSupported,
		IsHelmVMDownloadEnabled:          opts.IsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:       opts.IsIdentityServiceSupported,
		IsInstallerSupportEnabled:        opts.IsInstallerSupportEnabled,
		IsSupportBundleUploadEnabled:     opts.IsSupportBundleUploadEnabled,
		IsDeveloperModeEnabled:           opts.IsDeveloperModeEnabled,
		Email:                            opts.Email,
		EntitlementValues:                opts.EntitlementValues,
	}

	// if expiresAtDuration is set, calculate the expiresAt time
	if opts.ExpiresAtDuration > 0 {
		request.ExpiresAt = (time.Now().UTC().Add(opts.ExpiresAtDuration)).Format(time.RFC3339)
	} else {
		request.ExpiresAt = opts.ExpiresAt
	}
	var response CreateCustomerResponse
	err := c.DoJSON(context.TODO(), "POST", "/v3/customer", http.StatusCreated, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "create customer")
	}

	return response.Customer, nil
}

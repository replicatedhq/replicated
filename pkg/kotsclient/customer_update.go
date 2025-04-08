package kotsclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type UpdateCustomerRequest struct {
	Name                               string             `json:"name"`
	Channels                           []CustomerChannel  `json:"channels"`
	CustomID                           string             `json:"custom_id"`
	AppID                              string             `json:"app_id"`
	Type                               string             `json:"type"`
	ExpiresAt                          string             `json:"expires_at"`
	IsAirgapEnabled                    bool               `json:"is_airgap_enabled"`
	IsGitopsSupported                  bool               `json:"is_gitops_supported"`
	IsSnapshotSupported                bool               `json:"is_snapshot_supported"`
	IsKotsInstallEnabled               bool               `json:"is_kots_install_enabled"`
	IsEmbeddedClusterDownloadEnabled   bool               `json:"is_embedded_cluster_download_enabled"`
	IsEmbeddedClusterMultinodeEnabled  bool               `json:"is_embedded_cluster_multinode_enabled"`
	IsGeoaxisSupported                 bool               `json:"is_geoaxis_supported"`
	IsHelmVMDownloadEnabled            bool               `json:"is_helm_vm_download_enabled"`
	IsIdentityServiceSupported         bool               `json:"is_identity_service_supported"`
	IsSupportBundleUploadEnabled       bool               `json:"is_support_bundle_upload_enabled"`
	IsDeveloperModeEnabled             bool               `json:"is_dev_mode_enabled"`
	Email                              string             `json:"email,omitempty"`
	EntitlementValues                  []EntitlementValue `json:"entitlementValues"`

	// These fields were added after the "built in" fields feature was released.
	// If they are not pointer types, they will override the defaults.
	IsHelmInstallEnabled *bool `json:"is_helm_install_enabled,omitempty"`
	IsKurlInstallEnabled *bool `json:"is_kurl_install_enabled,omitempty"`
}

type UpdateCustomerResponse struct {
	Customer *types.Customer `json:"customer"`
}

type UpdateCustomerOpts struct {
	Name                               string
	CustomID                           string
	Channels                           []CustomerChannel
	AppID                              string
	ExpiresAt                          string
	ExpiresAtDuration                  time.Duration
	IsAirgapEnabled                    bool
	IsGitopsSupported                  bool
	IsSnapshotSupported                bool
	IsKotsInstallEnabled               bool
	IsHelmInstallEnabled               *bool
	IsKurlInstallEnabled               *bool
	IsEmbeddedClusterDownloadEnabled   bool
	IsEmbeddedClusterMultinodeEnabled  bool
	IsGeoaxisSupported                 bool
	IsHelmVMDownloadEnabled            bool
	IsIdentityServiceSupported         bool
	IsSupportBundleUploadEnabled       bool
	IsDeveloperModeEnabled             bool
	LicenseType                        string
	Email                              string
	EntitlementValues                  []EntitlementValue
}

func (c *VendorV3Client) UpdateCustomer(customerID string, opts UpdateCustomerOpts) (*types.Customer, error) {
	request := &UpdateCustomerRequest{
		Name:                               opts.Name,
		CustomID:                           opts.CustomID,
		Channels:                           opts.Channels,
		AppID:                              opts.AppID,
		Type:                               opts.LicenseType,
		IsAirgapEnabled:                    opts.IsAirgapEnabled,
		IsGitopsSupported:                  opts.IsGitopsSupported,
		IsSnapshotSupported:                opts.IsSnapshotSupported,
		IsKotsInstallEnabled:               opts.IsKotsInstallEnabled,
		IsHelmInstallEnabled:               opts.IsHelmInstallEnabled,
		IsKurlInstallEnabled:               opts.IsKurlInstallEnabled,
		IsEmbeddedClusterDownloadEnabled:   opts.IsEmbeddedClusterDownloadEnabled,
		IsEmbeddedClusterMultinodeEnabled:  opts.IsEmbeddedClusterMultinodeEnabled,
		IsGeoaxisSupported:                 opts.IsGeoaxisSupported,
		IsHelmVMDownloadEnabled:            opts.IsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:         opts.IsIdentityServiceSupported,
		IsSupportBundleUploadEnabled:       opts.IsSupportBundleUploadEnabled,
		IsDeveloperModeEnabled:             opts.IsDeveloperModeEnabled,
		Email:                              opts.Email,
		EntitlementValues:                  opts.EntitlementValues,
	}

	// If duration is set, calculate the expiry time
	if opts.ExpiresAtDuration > 0 {
		request.ExpiresAt = (time.Now().UTC().Add(opts.ExpiresAtDuration)).Format(time.RFC3339)
	} else {
		request.ExpiresAt = opts.ExpiresAt
	}
	var response UpdateCustomerResponse
	endpoint := fmt.Sprintf("/v3/customer/%s", customerID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusOK, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "update customer")
	}

	return response.Customer, nil
}

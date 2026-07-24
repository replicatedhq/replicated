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
	Name           *string           `json:"name,omitempty"`
	AddChannels    []CustomerChannel `json:"add_channels,omitempty"`
	RemoveChannels []string          `json:"remove_channels,omitempty"`
	CustomID       *string           `json:"custom_id,omitempty"`
	Type           string            `json:"type,omitempty"`
	ExpiresAt      *string           `json:"expires_at,omitempty"`

	// PATCH requests must distinguish an omitted flag from an explicit false value.
	IsAirgapEnabled                   *bool   `json:"is_airgap_enabled,omitempty"`
	IsGitopsSupported                 *bool   `json:"is_gitops_supported,omitempty"`
	IsSnapshotSupported               *bool   `json:"is_snapshot_supported,omitempty"`
	IsKotsInstallEnabled              *bool   `json:"is_kots_install_enabled,omitempty"`
	IsHelmInstallEnabled              *bool   `json:"is_helm_install_enabled,omitempty"`
	IsKurlInstallEnabled              *bool   `json:"is_kurl_install_enabled,omitempty"`
	IsEmbeddedClusterDownloadEnabled  *bool   `json:"is_embedded_cluster_download_enabled,omitempty"`
	IsEmbeddedClusterMultinodeEnabled *bool   `json:"is_embedded_cluster_multinode_enabled,omitempty"`
	IsGeoaxisSupported                *bool   `json:"is_geoaxis_supported,omitempty"`
	IsHelmVMDownloadEnabled           *bool   `json:"is_helm_vm_download_enabled,omitempty"`
	IsIdentityServiceSupported        *bool   `json:"is_identity_service_supported,omitempty"`
	IsSupportBundleUploadEnabled      *bool   `json:"is_support_bundle_upload_enabled,omitempty"`
	IsDeveloperModeEnabled            *bool   `json:"is_dev_mode_enabled,omitempty"`
	Email                             *string `json:"email,omitempty"`
}

type UpdateCustomerResponse struct {
	Customer *types.Customer `json:"customer"`
}

type UpdateCustomerOpts struct {
	Name                              *string
	CustomID                          *string
	AddChannels                       []CustomerChannel
	RemoveChannels                    []string
	ExpiresAt                         *string
	ExpiresAtDuration                 *time.Duration
	IsAirgapEnabled                   *bool
	IsGitopsSupported                 *bool
	IsSnapshotSupported               *bool
	IsKotsInstallEnabled              *bool
	IsHelmInstallEnabled              *bool
	IsKurlInstallEnabled              *bool
	IsEmbeddedClusterDownloadEnabled  *bool
	IsEmbeddedClusterMultinodeEnabled *bool
	IsGeoaxisSupported                *bool
	IsHelmVMDownloadEnabled           *bool
	IsIdentityServiceSupported        *bool
	IsSupportBundleUploadEnabled      *bool
	IsDeveloperModeEnabled            *bool
	LicenseType                       string
	Email                             *string
}

func (c *VendorV3Client) UpdateCustomer(customerID string, opts UpdateCustomerOpts) (*types.Customer, error) {
	request := &UpdateCustomerRequest{
		Name:                              opts.Name,
		CustomID:                          opts.CustomID,
		AddChannels:                       opts.AddChannels,
		RemoveChannels:                    opts.RemoveChannels,
		Type:                              opts.LicenseType,
		IsAirgapEnabled:                   opts.IsAirgapEnabled,
		IsGitopsSupported:                 opts.IsGitopsSupported,
		IsSnapshotSupported:               opts.IsSnapshotSupported,
		IsKotsInstallEnabled:              opts.IsKotsInstallEnabled,
		IsHelmInstallEnabled:              opts.IsHelmInstallEnabled,
		IsKurlInstallEnabled:              opts.IsKurlInstallEnabled,
		IsEmbeddedClusterDownloadEnabled:  opts.IsEmbeddedClusterDownloadEnabled,
		IsEmbeddedClusterMultinodeEnabled: opts.IsEmbeddedClusterMultinodeEnabled,
		IsGeoaxisSupported:                opts.IsGeoaxisSupported,
		IsHelmVMDownloadEnabled:           opts.IsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:        opts.IsIdentityServiceSupported,
		IsSupportBundleUploadEnabled:      opts.IsSupportBundleUploadEnabled,
		IsDeveloperModeEnabled:            opts.IsDeveloperModeEnabled,
		Email:                             opts.Email,
	}

	// If duration is set, calculate the expiry time
	if opts.ExpiresAtDuration != nil {
		expiresAt := time.Now().UTC().Add(*opts.ExpiresAtDuration).Format(time.RFC3339)
		request.ExpiresAt = &expiresAt
	} else {
		request.ExpiresAt = opts.ExpiresAt
	}
	var response UpdateCustomerResponse
	endpoint := fmt.Sprintf("/v3/customer/%s", customerID)
	err := c.DoJSON(context.TODO(), http.MethodPatch, endpoint, http.StatusOK, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "update customer")
	}

	return response.Customer, nil
}

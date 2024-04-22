package kotsclient

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type EntitlementValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CreateCustomerRequest struct {
	Name                             string             `json:"name"`
	ChannelID                        string             `json:"channel_id"`
	CustomID                         string             `json:"custom_id"`
	AppID                            string             `json:"app_id"`
	Type                             string             `json:"type"`
	ExpiresAt                        string             `json:"expires_at"`
	IsAirgapEnabled                  bool               `json:"is_airgap_enabled"`
	IsGitopsSupported                bool               `json:"is_gitops_supported"`
	IsSnapshotSupported              bool               `json:"is_snapshot_supported"`
	IsKotInstallEnabled              bool               `json:"is_kots_install_enabled"`
	IsEmbeddedClusterDownloadEnabled bool               `json:"is_embedded_cluster_download_enabled"`
	IsGeoaxisSupported               bool               `json:"is_geoaxis_supported"`
	IsHelmVMDownloadEnabled          bool               `json:"is_helm_vm_download_enabled"`
	IsIdentityServiceSupported       bool               `json:"is_identity_service_supported"`
	IsInstallerSupportEnabled        bool               `json:"is_installer_support_enabled"`
	IsSupportBundleUploadEnabled     bool               `json:"is_support_bundle_upload_enabled"`
	Email                            string             `json:"email,omitempty"`
	EntitlementValues                []EntitlementValue `json:"entitlementValues"`
}

type CreateCustomerResponse struct {
	Customer *types.Customer `json:"customer"`
}

type CreateCustomerOpts struct {
	Name                             string
	CustomID                         string
	ChannelID                        string
	AppID                            string
	ExpiresAt                        time.Duration
	IsAirgapEnabled                  bool
	IsGitopsSupported                bool
	IsSnapshotSupported              bool
	IsKotsInstallEnabled             bool
	IsEmbeddedClusterDownloadEnabled bool
	IsGeoaxisSupported               bool
	IsHelmVMDownloadEnabled          bool
	IsIdentityServiceSupported       bool
	IsInstallerSupportEnabled        bool
	IsSupportBundleUploadEnabled     bool
	LicenseType                      string
	Email                            string
	EntitlementValues                []EntitlementValue
}

func (c *VendorV3Client) CreateCustomer(opts CreateCustomerOpts) (*types.Customer, error) {
	request := &CreateCustomerRequest{
		Name:                             opts.Name,
		CustomID:                         opts.CustomID,
		ChannelID:                        opts.ChannelID,
		AppID:                            opts.AppID,
		Type:                             opts.LicenseType,
		IsAirgapEnabled:                  opts.IsAirgapEnabled,
		IsGitopsSupported:                opts.IsGitopsSupported,
		IsSnapshotSupported:              opts.IsSnapshotSupported,
		IsKotInstallEnabled:              opts.IsKotsInstallEnabled,
		IsEmbeddedClusterDownloadEnabled: opts.IsEmbeddedClusterDownloadEnabled,
		IsGeoaxisSupported:               opts.IsGeoaxisSupported,
		IsHelmVMDownloadEnabled:          opts.IsHelmVMDownloadEnabled,
		IsIdentityServiceSupported:       opts.IsIdentityServiceSupported,
		IsInstallerSupportEnabled:        opts.IsInstallerSupportEnabled,
		IsSupportBundleUploadEnabled:     opts.IsSupportBundleUploadEnabled,
		Email:                            opts.Email,
		EntitlementValues:                opts.EntitlementValues,
	}

	if opts.ExpiresAt > 0 {
		request.ExpiresAt = (time.Now().UTC().Add(opts.ExpiresAt)).Format(time.RFC3339)
	}
	var response CreateCustomerResponse
	err := c.DoJSON("POST", "/v3/customer", http.StatusCreated, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "create customer")
	}

	return response.Customer, nil
}

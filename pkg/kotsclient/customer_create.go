package kotsclient

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateCustomerRequest struct {
	Name                string `json:"name"`
	ChannelID           string `json:"channel_id"`
	AppID               string `json:"app_id"`
	Type                string `json:"type"`
	ExpiresAt           string `json:"expires_at"`
	IsAirgapEnabled     bool   `json:"is_airgap_enabled"`
	IsGitopsSupported   bool   `json:"is_gitops_supported"`
	IsSnapshotSupported bool   `json:"is_snapshot_supported"`
	IsKotInstallEnabled bool   `json:"is_kots_install_enabled"`
	Email               string `json:"email,omitempty"`
}

type CreateCustomerResponse struct {
	Customer *types.Customer `json:"customer"`
}

type CreateCustomerOpts struct {
	Name                string
	ChannelID           string
	AppID               string
	ExpiresAt           time.Duration
	IsAirgapEnabled     bool
	IsGitopsSupported   bool
	IsSnapshotSupported bool
	IsKotInstallEnabled bool
	LicenseType         string
	Email               string
}

func (c *VendorV3Client) CreateCustomer(opts CreateCustomerOpts) (*types.Customer, error) {
	request := &CreateCustomerRequest{
		Name:                opts.Name,
		ChannelID:           opts.ChannelID,
		AppID:               opts.AppID,
		Type:                opts.LicenseType,
		IsAirgapEnabled:     opts.IsAirgapEnabled,
		IsGitopsSupported:   opts.IsGitopsSupported,
		IsSnapshotSupported: opts.IsSnapshotSupported,
		IsKotInstallEnabled: opts.IsKotInstallEnabled,
		Email:               opts.Email,
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

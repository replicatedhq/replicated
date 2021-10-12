package kotsclient

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateCustomerRequest struct {
	Name              string `json:"name"`
	ChannelID         string `json:"channel_id"`
	AppID             string `json:"app_id"`
	Type              string `json:"type"`
	ExpiresAt         string `json:"expires_at"`
	IsAirgapEnabled   bool   `json:"is_airgap_enabled"`
	IsGitopsSupported bool   `json:"is_gitops_supported"`
}

type CreateCustomerResponse struct {
	Customer *types.Customer `json:"customer"`
}

func (c *VendorV3Client) CreateCustomer(name string, appID string, channelID string, expiresIn time.Duration, airgap, gitops bool) (*types.Customer, error) {
	request := &CreateCustomerRequest{
		Name:              name,
		ChannelID:         channelID,
		AppID:             appID,
		IsAirgapEnabled:   airgap,
		IsGitopsSupported: gitops,
		Type:              "dev", // hardcode for now
	}

	if expiresIn > 0 {
		request.ExpiresAt = (time.Now().UTC().Add(expiresIn)).Format(time.RFC3339)
	}
	var response CreateCustomerResponse
	err := c.DoJSON("POST", "/v3/customer", http.StatusCreated, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "create customer")
	}

	return response.Customer, nil
}

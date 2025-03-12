package kotsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type UpdateNetworkPolicyRequest struct {
	Policy string `json:"policy"`
}

type UpdateNetworkPolicyResponse struct {
	Network *types.Network `json:"network"`
	Errors  []string       `json:"errors"`
}

type UpdateNetworkPolicyOpts struct {
	Policy string `json:"policy"`
}

type UpdateNetworkPolicyErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	}
}

// UpdateNetworkPolicy updates a network's policy setting
func (c *VendorV3Client) UpdateNetworkPolicy(networkID string, opts UpdateNetworkPolicyOpts) (*types.Network, error) {
	req := UpdateNetworkPolicyRequest{
		Policy: opts.Policy,
	}

	return c.doUpdateNetworkPolicyRequest(networkID, req)
}

func (c *VendorV3Client) doUpdateNetworkPolicyRequest(networkID string, req UpdateNetworkPolicyRequest) (*types.Network, error) {
	resp := UpdateNetworkPolicyResponse{}
	endpoint := fmt.Sprintf("/v3/network/%s/policy", networkID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				errResp := &UpdateNetworkPolicyErrorResponse{}
				if jsonErr := json.Unmarshal(apiErr.Body, errResp); jsonErr != nil {
					return nil, fmt.Errorf("unmarshal error response: %w", err)
				}
				return nil, errors.New(errResp.Error.Message)
			}
		}
		return nil, err
	}

	return resp.Network, nil
}

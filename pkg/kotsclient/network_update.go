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

type UpdateNetworkRequest struct {
	Policy        string `json:"policy"`
	CollectReport *bool  `json:"collect_report,omitempty"`
}

type UpdateNetworkResponse struct {
	Network *types.Network `json:"network"`
	Errors  []string       `json:"errors"`
}

type UpdateNetworkOpts struct {
	Policy        string `json:"policy"`
	CollectReport *bool  `json:"collect_report,omitempty"`
}

type UpdateNetworkErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	}
}

// UpdateNetworkPolicy updates a network's policy setting
func (c *VendorV3Client) UpdateNetwork(networkID string, opts UpdateNetworkOpts) (*types.Network, error) {
	req := UpdateNetworkRequest{
		Policy:        opts.Policy,
		CollectReport: opts.CollectReport,
	}
	return c.doUpdateNetworkRequest(networkID, req)
}

func (c *VendorV3Client) doUpdateNetworkRequest(networkID string, req UpdateNetworkRequest) (*types.Network, error) {
	resp := UpdateNetworkResponse{}
	endpoint := fmt.Sprintf("/v3/network/%s/update", networkID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				errResp := &UpdateNetworkErrorResponse{}
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

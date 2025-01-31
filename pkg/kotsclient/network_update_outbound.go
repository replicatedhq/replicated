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

type UpdateNetworkOutboundRequest struct {
	Outbound string `json:"outbound"`
}

type UpdateNetworkOutboundResponse struct {
	Network *types.Network `json:"network"`
	Errors  []string       `json:"errors"`
}

type UpdateNetworkOutboundOpts struct {
	Outbound string
}

type UpdateNetworkOutboundErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	}
}

func (c *VendorV3Client) UpdateNetworkOutbound(networkID string, opts UpdateNetworkOutboundOpts) (*types.Network, error) {
	req := UpdateNetworkOutboundRequest{
		Outbound: opts.Outbound,
	}

	return c.doUpdateNetworkOutboundRequest(networkID, req)
}

func (c *VendorV3Client) doUpdateNetworkOutboundRequest(networkID string, req UpdateNetworkOutboundRequest) (*types.Network, error) {
	resp := UpdateNetworkOutboundResponse{}
	endpoint := fmt.Sprintf("/v3/network/%s/outbound", networkID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				errResp := &UpdateNetworkOutboundErrorResponse{}
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

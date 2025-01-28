package kotsclient

import (
	"context"
	"fmt"
	"net/http"

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
		return nil, err
	}

	return resp.Network, nil
}

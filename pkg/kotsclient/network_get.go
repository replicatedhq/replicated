package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type GetNetworkResponse struct {
	Network *types.Network `json:"network"`
	Error   string         `json:"error"`
}

func (c *VendorV3Client) GetNetwork(id string) (*types.Network, error) {
	network := GetNetworkResponse{}

	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/network/%s", id), http.StatusOK, nil, &network)
	if err != nil {
		return nil, err
	}

	return network.Network, nil
}

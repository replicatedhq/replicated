package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListVMPortsResponse struct {
	Ports []*types.VMPort `json:"ports"`
}

func (c *VendorV3Client) ListVMPorts(vmID string) ([]*types.VMPort, error) {
	resp := ListVMPortsResponse{}
	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/vm/%s/ports", vmID), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Ports, nil
}

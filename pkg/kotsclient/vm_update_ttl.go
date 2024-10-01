package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type UpdateVMTTLRequest struct {
	TTL string `json:"ttl"`
}

type UpdateVMTTLResponse struct {
	VM     *types.VM `json:"vm"`
	Errors []string  `json:"errors"`
}

type UpdateVMTTLOpts struct {
	TTL string
}

func (c *VendorV3Client) UpdateVMTTL(vmID string, opts UpdateVMTTLOpts) (*types.VM, error) {
	req := UpdateVMTTLRequest{
		TTL: opts.TTL,
	}

	return c.doUpdateVMTTLRequest(vmID, req)
}

func (c *VendorV3Client) doUpdateVMTTLRequest(vmID string, req UpdateVMTTLRequest) (*types.VM, error) {
	resp := UpdateVMTTLResponse{}
	endpoint := fmt.Sprintf("/v3/vm/%s/ttl", vmID)
	err := c.DoJSON("PUT", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.VM, nil
}

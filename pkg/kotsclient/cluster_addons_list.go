package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClusterAddOnsResponse struct {
	AddOns []*types.ClusterAddOn `json:"addons"`
}

func (c *VendorV3Client) ListClusterAddOns() ([]*types.ClusterAddOn, error) {
	resp := ListClusterAddOnsResponse{}
	err := c.DoJSON("GET", "/v3/cluster/addons", http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.AddOns, nil
}

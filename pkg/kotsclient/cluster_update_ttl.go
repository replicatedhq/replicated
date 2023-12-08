package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type UpdateClusterTTLRequest struct {
	TTL string `json:"ttl"`
}

type UpdateClusterTTLResponse struct {
	Cluster *types.Cluster `json:"cluster"`
	Errors  []string       `json:"errors"`
}

type UpdateClusterTTLOpts struct {
	TTL string
}

func (c *VendorV3Client) UpdateClusterTTL(clusterID string, opts UpdateClusterTTLOpts) (*types.Cluster, error) {
	req := UpdateClusterTTLRequest{
		TTL: opts.TTL,
	}

	return c.doUpdateClusterTTLRequest(clusterID, req)
}

func (c *VendorV3Client) doUpdateClusterTTLRequest(clusterID string, req UpdateClusterTTLRequest) (*types.Cluster, error) {
	resp := UpdateClusterTTLResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/ttl", clusterID)
	err := c.DoJSON("PUT", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Cluster, nil
}

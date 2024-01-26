package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterIngressOpts struct {
	ClusterID string
	Target    string
}

type CreateClusterIngressRequest struct {
	Target string `json:"target"`
}

type CreateClusterIngressResponse struct {
	Ingress *types.ClusterAddOn `json:"ingress"`
}

func (c *VendorV3Client) CreateClusterIngress(opts CreateClusterIngressOpts) (*types.ClusterAddOn, error) {
	req := CreateClusterIngressRequest{
		Target: opts.Target,
	}

	return c.doCreateClusterIngressRequest(opts.ClusterID, req)
}

func (c *VendorV3Client) doCreateClusterIngressRequest(clusterID string, req CreateClusterIngressRequest) (*types.ClusterAddOn, error) {
	resp := CreateClusterIngressResponse{}
	endpoint := "/v3/cluster/" + clusterID + "/addons/ingress"
	err := c.DoJSON("POST", endpoint, http.StatusCreated, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Ingress, nil
}

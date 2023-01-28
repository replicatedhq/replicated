package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClustersRequest struct {
}

type ListClustersResponse struct {
	Clusters []*types.Cluster `json:"clusters"`
}

func (c *VendorV3Client) ListClusters() ([]*types.Cluster, error) {
	reqBody := &ListClustersRequest{}
	clusters := ListClustersResponse{}
	err := c.DoJSON("GET", "/v3/clusters", http.StatusOK, reqBody, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters.Clusters, nil
}

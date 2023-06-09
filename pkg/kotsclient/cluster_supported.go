package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListSupportedClustersRequest struct {
}

type ListSupportedClustersResponse struct {
	Clusters []*types.SupportedCluster `json:"supported-clusters"`
}

func (c *VendorV3Client) ListSupportedClusters() ([]*types.SupportedCluster, error) {
	reqBody := &ListSupportedClustersRequest{}
	clusters := ListSupportedClustersResponse{}
	err := c.DoJSON("GET", "/v3/supported-clusters", http.StatusOK, reqBody, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters.Clusters, nil
}

package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClusterVersionsResponse struct {
	Clusters []*types.ClusterVersion `json:"cluster-versions"`
}

func (c *VendorV3Client) ListClusterVersions() ([]*types.ClusterVersion, error) {
	clusters := ListClusterVersionsResponse{}
	err := c.DoJSON("GET", "/v3/cluster/versions", http.StatusOK, nil, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters.Clusters, nil
}

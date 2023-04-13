package kotsclient

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClustersRequest struct {
}

type ListClustersResponse struct {
	Clusters []*types.Cluster `json:"clusters"`
}

func (c *VendorV3Client) ListClusters(includeTerminated bool) ([]*types.Cluster, error) {
	reqBody := &ListClustersRequest{}
	clusters := ListClustersResponse{}
	err := c.DoJSON("GET", fmt.Sprintf("/v3/clusters?include-terminated=%s", strconv.FormatBool(includeTerminated)), http.StatusOK, reqBody, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters.Clusters, nil
}

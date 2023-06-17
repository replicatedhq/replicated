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

func (c *VendorV3Client) ListClusters(includeTerminated bool, startTime string, endTime string) ([]*types.Cluster, error) {
	reqBody := &ListClustersRequest{}
	clusters := ListClustersResponse{}
	err := c.DoJSON("GET", fmt.Sprintf("/v3/clusters?include-terminated=%s&start-time=%s&end-time=%s", strconv.FormatBool(includeTerminated), startTime, endTime),
		http.StatusOK, reqBody, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters.Clusters, nil
}

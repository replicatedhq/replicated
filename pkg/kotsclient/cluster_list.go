package kotsclient

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClusterRequest struct{}

type ListClustersResponse struct {
	Clusters []*types.Cluster `json:"clusters"`
}

type ListClustersOpts struct {
	IncludeTerminated bool
	StartTime         time.Time
	EndTime           time.Time
}

func (c *VendorV3Client) ListClusters(opts ListClustersOpts) ([]*types.Cluster, error) {
	reqBody := ListClusterRequest{}
	clusters := ListClustersResponse{}
	err := c.DoJSON("GET", fmt.Sprintf("/v3/clusters?include-terminated=%s&start-time=%s&end-time=%s", strconv.FormatBool(opts.IncludeTerminated), strconv.FormatInt(opts.StartTime.Unix(), 10), strconv.FormatInt(opts.EndTime.Unix(), 10)), http.StatusOK, reqBody, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters.Clusters, nil
}

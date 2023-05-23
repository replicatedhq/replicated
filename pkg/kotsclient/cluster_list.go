package kotsclient

import (
	"net/http"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClustersRequest struct {
	IncludeTerminated bool     `json:"include_terminated"`
	ClusterIDs        []string `json:"cluster_ids"`
	StartTime         int64    `json:"start_time"`
	EndTime           int64    `json:"end_time"`
}

type ListClustersResponse struct {
	Continue bool             `json:"continue"`
	Clusters []*types.Cluster `json:"clusters"`
}

type ListClustersOpts struct {
	IncludeTerminated bool
	ClusterIDs        []string
	StartTime         time.Time
	EndTime           time.Time
}

func (c *VendorV3Client) ListClusters(opts ListClustersOpts) ([]*types.Cluster, error) {
	reqBody := &ListClustersRequest{
		IncludeTerminated: opts.IncludeTerminated,
		ClusterIDs:        opts.ClusterIDs,
		StartTime:         opts.StartTime.Unix(),
		EndTime:           opts.EndTime.Unix(),
	}
	clusters := ListClustersResponse{}
	err := c.DoJSON("GET", "/v3/clusters", http.StatusOK, reqBody, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters.Clusters, nil
}

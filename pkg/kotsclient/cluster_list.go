package kotsclient

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClustersRequest struct {
}

type ListClustersResponse struct {
	Clusters      []*types.Cluster `json:"clusters"`
	TotalClusters int              `json:"totalClusters"`
}

func (c *VendorV3Client) ListClusters(includeTerminated bool, startTime string, endTime string) ([]*types.Cluster, error) {
	reqBody := &ListClustersRequest{}

	allClusters := []*types.Cluster{}
	page := 0
	for {
		clusters := ListClustersResponse{}
		err := c.DoJSON("GET", fmt.Sprintf("/v3/clusters?currentPage=%d&show-terminated=%s&start-time=%s&end-time=%s", page, strconv.FormatBool(includeTerminated), startTime, endTime),
			http.StatusOK, reqBody, &clusters)
		if err != nil {
			return nil, errors.Wrapf(err, "list clusters page %d", page)
		}

		if len(allClusters) == clusters.TotalClusters || len(clusters.Clusters) == 0 {
			break
		}

		allClusters = append(allClusters, clusters.Clusters...)

		page = page + 1
	}
	return allClusters, nil
}

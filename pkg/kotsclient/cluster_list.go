package kotsclient

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClustersResponse struct {
	Clusters      []*types.Cluster `json:"clusters"`
	TotalClusters int              `json:"totalClusters"`
}

func (c *VendorV3Client) ListClusters(includeTerminated bool, startTime string, endTime string) ([]*types.Cluster, error) {
	allClusters := []*types.Cluster{}
	page := 0
	for {
		clusters := ListClustersResponse{}

		v := url.Values{}
		if startTime != "" {
			v.Set("start-time", startTime)
		}
		if endTime != "" {
			v.Set("end-time", endTime)
		}
		v.Set("currentPage", strconv.Itoa(page))
		v.Set("show-terminated", strconv.FormatBool(includeTerminated))
		url := fmt.Sprintf("/v3/clusters?%s", v.Encode())
		err := c.DoJSON("GET", url, http.StatusOK, nil, &clusters)
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

package kotsclient

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClustersResponse struct {
	Clusters      []*types.Cluster `json:"clusters"`
	TotalClusters int              `json:"totalClusters"`
}

func (c *VendorV3Client) ListClusters(includeTerminated bool, startTime *time.Time, endTime *time.Time) ([]*types.Cluster, error) {
	allClusters := []*types.Cluster{}
	page := 0
	for {
		clusters := ListClustersResponse{}

		v := url.Values{}
		const longForm = "2006-01-02T15:04:05Z"
		if startTime != nil {
			v.Set("start-time", startTime.Format(longForm))
		}
		if endTime != nil {
			v.Set("end-time", endTime.Format(longForm))
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

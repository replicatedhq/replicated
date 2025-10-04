package kotsclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListNetworksResponse struct {
	Networks      []*types.Network `json:"networks"`
	TotalNetworks int              `json:"totalNetworks"`
}

func (c *VendorV3Client) ListNetworks(includeTerminated bool, startTime *time.Time, endTime *time.Time) ([]*types.Network, error) {
	allNetworks := []*types.Network{}
	page := 0
	for {
		networks := ListNetworksResponse{}

		v := url.Values{}
		if startTime != nil {
			v.Set("start-time", startTime.Format(time.RFC3339))
		}
		if endTime != nil {
			v.Set("end-time", endTime.Format(time.RFC3339))
		}
		v.Set("currentPage", strconv.Itoa(page))
		v.Set("show-terminated", strconv.FormatBool(includeTerminated))
		url := fmt.Sprintf("/v3/networks?%s", v.Encode())
		err := c.DoJSON(context.TODO(), "GET", url, http.StatusOK, nil, &networks)
		if err != nil {
			return nil, errors.Wrapf(err, "list networks page %d", page)
		}

		if len(allNetworks) == networks.TotalNetworks || len(networks.Networks) == 0 {
			break
		}

		allNetworks = append(allNetworks, networks.Networks...)

		page = page + 1
	}
	return allNetworks, nil
}

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

type ListVMsResponse struct {
	VMs      []*types.VM `json:"vms"`
	TotalVMs int         `json:"totalVMs"`
}

func (c *VendorV3Client) ListVMs(includeTerminated bool, startTime *time.Time, endTime *time.Time) ([]*types.VM, error) {
	allVMs := []*types.VM{}
	page := 0
	for {
		vms := ListVMsResponse{}

		v := url.Values{}
		if startTime != nil {
			v.Set("start-time", startTime.Format(time.RFC3339))
		}
		if endTime != nil {
			v.Set("end-time", endTime.Format(time.RFC3339))
		}
		v.Set("currentPage", strconv.Itoa(page))
		v.Set("show-terminated", strconv.FormatBool(includeTerminated))
		url := fmt.Sprintf("/v3/vms?%s", v.Encode())
		err := c.DoJSON(context.TODO(), "GET", url, http.StatusOK, nil, &vms)
		if err != nil {
			return nil, errors.Wrapf(err, "list vms page %d", page)
		}

		if len(allVMs) == vms.TotalVMs || len(vms.VMs) == 0 {
			break
		}

		allVMs = append(allVMs, vms.VMs...)

		page = page + 1
	}
	return allVMs, nil
}

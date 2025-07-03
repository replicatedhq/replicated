package kotsclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) GetNetworkReport(id string) (*types.NetworkReport, error) {
	return c.GetNetworkReportAfter(id, nil)
}

func (c *VendorV3Client) GetNetworkReportAfter(id string, after *time.Time) (*types.NetworkReport, error) {
	report := &types.NetworkReport{}

	urlPath := fmt.Sprintf("/v3/network/%s/report", id)
	if after != nil {
		v := url.Values{}
		v.Set("after", after.Format(time.RFC3339Nano))
		urlPath = fmt.Sprintf("%s?%s", urlPath, v.Encode())
	}

	err := c.DoJSON(context.TODO(), "GET", urlPath, http.StatusOK, nil, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

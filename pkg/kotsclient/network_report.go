package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) GetNetworkReport(id string) (*types.NetworkReport, error) {
	report := &types.NetworkReport{}

	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/network/%s/report", id), http.StatusOK, nil, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

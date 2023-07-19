package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CustomHostnamesListResponse struct {
	Body types.KotsAppCustomHostnames
}

func (c *VendorV3Client) ListCustomHostnames(appID string) (*types.KotsAppCustomHostnames, error) {
	resp := CustomHostnamesListResponse{}
	path := fmt.Sprintf("/v3/app/%s/custom-hostnames", appID)
	err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "list custom hostnames appId %s", appID)
	}

	return &resp.Body, nil
}

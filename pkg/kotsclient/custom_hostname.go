package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) ListCustomHostnames(appID string) (*types.KotsAppCustomHostnames, error) {
	resp := types.KotsAppCustomHostnames{}
	path := fmt.Sprintf("/v3/app/%s/custom-hostnames", appID)
	err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "list custom hostnames appId %s", appID)
	}

	return &resp, nil
}

func (c *VendorV3Client) ListDefaultHostnames(appID string) (*types.DefaultHostnames, error) {
	resp := types.DefaultHostnames{}
	path := fmt.Sprintf("/v3/app/%s/default-hostnames", appID)
	err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "list default hostnames appId %s", appID)
	}

	return &resp, nil
}

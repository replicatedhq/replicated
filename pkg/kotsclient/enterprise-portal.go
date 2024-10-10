package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type GetEnterprisePortalStatusResponse struct {
	Status string `json:"status"`
}

func (c *VendorV3Client) GetEnterprisePortalStatus(appID string) (string, error) {
	var response = GetEnterprisePortalStatusResponse{}

	err := c.DoJSON("GET", fmt.Sprintf("/v3/app/%s/enterprise-portal/status", appID), http.StatusOK, nil, &response)
	if err != nil {
		return "", errors.Wrap(err, "get enterprise portal status")
	}

	return response.Status, nil
}

func (c *VendorV3Client) UpdateEnterprisePortalStatus(appID string, status string) (string, error) {
	var response = GetEnterprisePortalStatusResponse{}

	err := c.DoJSON("PUT", fmt.Sprintf("/v3/app/%s/enterprise-portal/status", appID), http.StatusOK, status, &response)
	if err != nil {
		return "", errors.Wrap(err, "update enterprise portal status")
	}

	return response.Status, nil
}

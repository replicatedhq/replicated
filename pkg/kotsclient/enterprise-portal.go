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

type InviteEnterprisePortalResponse struct {
}

type InviteEnterprisePortalRequest struct {
	CustomerID   string `json:"customer_id"`
	EmailAddress string `json:"email_address"`
}

func (c *VendorV3Client) SendEnterprisePortalInvite(appID string, customerID string, emailAddress string) error {
	var response = InviteEnterprisePortalResponse{}

	var request = InviteEnterprisePortalRequest{
		CustomerID:   customerID,
		EmailAddress: emailAddress,
	}

	err := c.DoJSON("POST", fmt.Sprintf("/v3/app/%s/enterprise-portal/customer-user", appID), http.StatusCreated, request, &response)
	if err != nil {
		return errors.Wrap(err, "send enterprise portal invite")
	}

	return nil
}

type EnterprisePortalUser struct {
	Email string `json:"email"`
}

func (c *VendorV3Client) ListEnterprisePortalUsers(appID string, includeInvites bool) ([]EnterprisePortalUser, error) {
	var response = struct {
		Users []EnterprisePortalUser `json:"users"`
	}{}

	endpoint := fmt.Sprintf("/v3/app/%s/enterprise-portal/customer-users", appID)
	if includeInvites {
		endpoint = fmt.Sprintf("%s?includeInvites=true", endpoint)
	}

	err := c.DoJSON("GET", endpoint, http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list enterprise portal users")
	}

	return response.Users, nil
}

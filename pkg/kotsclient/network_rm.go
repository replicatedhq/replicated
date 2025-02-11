package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

type RemoveNetworkRequest struct {
	ID string `json:"id"`
}

type RemoveNetworkResponse struct {
	Error string `json:"error"`
}

func (c *VendorV3Client) RemoveNetwork(id string) error {
	resp := RemoveNetworkResponse{}

	url := fmt.Sprintf("/v3/network/%s", id)
	err := c.DoJSON(context.TODO(), "DELETE", url, http.StatusOK, nil, &resp)
	if err != nil {
		return err
	}
	return nil
}

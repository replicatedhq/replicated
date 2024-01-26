package kotsclient

import (
	"fmt"
	"net/http"
)

type DeleteClusterAddOnRequest struct {
	ID string `json:"id"`
}

type DeleteClusterAddOnResponse struct {
	Error string `json:"error"`
}

func (c *VendorV3Client) DeleteClusterAddOn(id string) error {
	resp := DeleteClusterAddOnResponse{}

	url := fmt.Sprintf("/v3/cluster/addons/%s", id)
	err := c.DoJSON("DELETE", url, http.StatusOK, nil, &resp)
	if err != nil {
		return err
	}

	return nil
}

package kotsclient

import (
	"fmt"
	"net/http"
)

type RemoveClusterRequest struct {
	ID string `json:"id"`
}

type RemoveClusterResponse struct {
	Error string `json:"error"`
}

func (c *VendorV3Client) RemoveCluster(id string, force bool) error {
	resp := RemoveClusterResponse{}

	url := fmt.Sprintf("/v3/cluster/%s", id)
	if force {
		url = fmt.Sprintf("%s?force=true", url)
	}

	err := c.DoJSON("DELETE", url, http.StatusOK, nil, &resp)
	if err != nil {
		return err
	}
	return nil
}

package kotsclient

import (
	"fmt"
	"net/http"
)

type DeleteClusterAddonRequest struct {
	ID string `json:"id"`
}

func (c *VendorV3Client) DeleteClusterAddon(id string) error {
	endpoint := fmt.Sprintf("/v3/cluster/addons/%s", id)
	err := c.DoJSON("DELETE", endpoint, http.StatusNoContent, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

package kotsclient

import (
	"fmt"
	"net/http"
)

type DeleteClusterAddonRequest struct {
	ID string `json:"id"`
}

func (c *VendorV3Client) DeleteClusterAddon(clusterID, addonID string) error {
	endpoint := fmt.Sprintf("/v3/cluster/%s/addons/%s", clusterID, addonID)
	err := c.DoJSON("DELETE", endpoint, http.StatusNoContent, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

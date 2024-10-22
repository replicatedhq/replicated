package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

func (c *VendorV3Client) DeleteClusterAddon(clusterID, addonID string) error {
	endpoint := fmt.Sprintf("/v3/cluster/%s/addons/%s", clusterID, addonID)
	err := c.DoJSON(context.TODO(), "DELETE", endpoint, http.StatusNoContent, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

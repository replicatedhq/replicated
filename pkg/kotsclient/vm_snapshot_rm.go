package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

func (c *VendorV3Client) DeleteVMSnapshot(snapshotID string) error {
	endpoint := fmt.Sprintf("/v3/snapshots/%s", snapshotID)
	err := c.DoJSON(context.TODO(), "DELETE", endpoint, http.StatusOK, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

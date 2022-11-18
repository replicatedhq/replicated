package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *VendorV3Client) DownloadLicense(appID string, customerID string) ([]byte, error) {
	path := fmt.Sprintf("/v3/app/%s/customer/%s/license-download", appID, customerID)
	licenseBytes, err := c.HTTPGet(path, http.StatusOK)
	if err != nil {
		return nil, errors.Wrapf(err, "download license")
	}
	return licenseBytes, nil
}

package kotsclient

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

func (c HTTPClient) DownloadLicense(appID string, customerID string) ([]byte, error) {
	path := fmt.Sprintf("/v3/app/%s/customer/%s/license-download", appID, customerID)
	licenseBytes, err := c.HTTPGet(path, http.StatusOK)
	if err != nil {
		return nil, errors.Wrapf(err, "list channels")
	}
	return licenseBytes, nil
}


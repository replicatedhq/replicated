package kotsclient

import (
	"fmt"
	"github.com/pkg/errors"
)

func (c *HybridClient) FetchLicense(appSlug string, customerID string) ([]byte, error) {
	bytes, err := c.doRawHTTP("GET", fmt.Sprintf("/kots/license/download/%s/%s?authorization=%s", appSlug, customerID, c.apiKey), 200, nil)
	if err != nil {
		return nil, errors.Wrap(err, "download license")
	}

	return bytes, nil
}

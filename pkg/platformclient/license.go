package platformclient

import (
	"fmt"
	"net/http"

	v2 "github.com/replicatedhq/replicated/gen/go/v2"
)

// CreateLicense creates a new License.
func (c *HTTPClient) CreateLicense(license *v2.LicenseV2) (*v2.LicenseV2, error) {
	created := &v2.LicenseV2{}
	if err := c.DoJSON("POST", "/v2/license", http.StatusCreated, license, created); err != nil {
		return nil, fmt.Errorf("CreateLicense: %w", err)
	}
	return created, nil
}

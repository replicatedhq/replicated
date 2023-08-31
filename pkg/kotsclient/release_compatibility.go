package kotsclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) ReportReleaseCompatibility(appID string, sequence int64, distribution string, version string, success bool, notes string) error {
	request := types.CompatibilityResult{
		Distribution: distribution,
		Version:      version,
	}

	now := time.Now()
	if success {
		request.SuccessAt = &now
		request.SuccessNotes = notes
	} else {
		request.FailureAt = &now
		request.FailureNotes = notes
	}

	path := fmt.Sprintf("/v3/app/%s/release/%v/compatibility", appID, sequence)
	err := c.DoJSON("POST", path, http.StatusCreated, request, nil)
	if err != nil {
		return err
	}

	return nil
}

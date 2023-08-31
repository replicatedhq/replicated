package kotsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type ReportReleaseCompatibilityErrorResponseBody struct {
	Error ReportReleaseCompatibilityErrorResponse `json:"Error"`
}
type ReportReleaseCompatibilityErrorResponse struct {
	Message         string           `json:"message"`
	ValidationError *ValidationError `json:"validationError,omitempty"`
}

func (c *VendorV3Client) ReportReleaseCompatibility(appID string, sequence int64, distribution string, version string, success bool, notes string) (*ValidationError, error) {
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
		if err, ok := err.(platformclient.APIError); ok && err.StatusCode == http.StatusBadRequest {
			// parse the error response body
			respBody := &ReportReleaseCompatibilityErrorResponseBody{}
			if err := json.NewDecoder(bytes.NewReader(err.Body)).Decode(respBody); err != nil {
				return nil, fmt.Errorf("Error decoding error response body: %w", err)
			}
			if respBody.Error.ValidationError != nil {
				return respBody.Error.ValidationError, nil
			}
		}
		return nil, err
	}

	return nil, nil
}

package kotsclient

import (
	"fmt"
	"net/http"
)

type ReportClusterCompatibilityRequest struct {
	SupportBundleID string  `json:"support_bundle_id"`
	ApplicationID   *string `json:"application_id"`
	Sequence        *int64  `json:"sequence"`
	AppVersion      *string `json:"app_version"`
}

type ValidateVersion struct {
	ClusterGUID     string `json:"cluster_guid"`
	SupportBundleID string `json:"support_bundle_id"`
}

type ListClusterValidationsResponse struct {
	ValidateVersion []ValidateVersion `json:"validation_versions"`
}

func (c *VendorV3Client) ReportClusterCompatibility(clusterID string, bundleID string, appID string, sequence int64, appVersion string) error {
	request := ReportClusterCompatibilityRequest{
		SupportBundleID: bundleID,
	}
	if appVersion != "" {
		request.AppVersion = &appVersion
	} else {
		request.ApplicationID = &appID
		request.Sequence = &sequence
	}
	path := fmt.Sprintf("/v3/cluster/%s/validate", clusterID)
	err := c.DoJSON("POST", path, http.StatusCreated, request, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *VendorV3Client) ListClusterValidations(appID string, sequence int64, appVersion string) ([]ValidateVersion, error) {
	resp := ListClusterValidationsResponse{}
	var path string
	if appID != "" && sequence != -1 {
		path = fmt.Sprintf("/v3/cluster/validations?appId=%s&sequence=%d", appID, sequence)
	} else {
		path = fmt.Sprintf("/v3/cluster/validations?appVersion=%s", appVersion)
	}
	err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.ValidateVersion, nil
}

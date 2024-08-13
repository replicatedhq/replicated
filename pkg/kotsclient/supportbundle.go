package kotsclient

import (
	"fmt"
	"net/http"
)

type UploadResponse struct {
	BundleID string `json:"bundleId"`
	URL      string `json:"url"`
}

func (c *VendorV3Client) GetSupportBundleUploadURL() (*UploadResponse, error) {
	uploadResponse := UploadResponse{}

	err := c.DoJSON("GET", "/v3/supportbundle/upload-url", http.StatusOK, nil, &uploadResponse)
	if err != nil {
		return nil, err
	}

	return &uploadResponse, nil
}

type UploadedResponse struct {
	Slug string `json:"slug"`
}

func (c *VendorV3Client) SupportBundleUploaded(bundleID string) error {
	resp := UploadedResponse{}
	req := map[string]interface{}{
		"app_id":    "",
		"issue_url": "",
	}
	endpoint := fmt.Sprintf("/v3/supportbundle/%s/uploaded", bundleID)
	err := c.DoJSON("POST", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return err
	}

	return nil
}

type SupportBundleInsight struct {
	Level           string      `json:"level"`
	Primary         string      `json:"primary"`
	Key             string      `json:"key"`
	Detail          string      `json:"detail"`
	Icon            string      `json:"icon"`
	IconKey         string      `json:"icon_key"`
	DesiredPosition int         `json:"desiredPosition"`
	InvolvedObject  interface{} `json:"involvedObject"`
}

type SupportBundleStatus string

const (
	SupportBundleStatusPending  SupportBundleStatus = "pending"
	SupportBundleStatusUploaded SupportBundleStatus = "uploaded"
)

type SupportBundle struct {
	ID       string                  `json:"id"`
	Status   SupportBundleStatus     `json:"status"`
	Insights []*SupportBundleInsight `json:"insights"`
}

type GetSupportBundleResponse struct {
	Bundle   *SupportBundle          `json:"bundle"`
	Insights []*SupportBundleInsight `json:"insights"`
}

func (c *VendorV3Client) GetSupportBundle(bundleID string) (*SupportBundle, error) {
	resp := GetSupportBundleResponse{}

	err := c.DoJSON("GET", fmt.Sprintf("/v3/supportbundle/%s", bundleID), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Bundle, nil
}

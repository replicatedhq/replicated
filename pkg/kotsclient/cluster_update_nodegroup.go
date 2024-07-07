package kotsclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type UpdateClusterNodegroupRequest struct {
	Count    int64  `json:"count"`
	MinCount *int64 `json:"minCount,omitempty"`
	MaxCount *int64 `json:"maxCount,omitempty"`
}

type UpdateClusterNodegroupResponse struct {
	Cluster *types.Cluster `json:"cluster"`
	Errors  []string       `json:"errors"`
}

type UpdateClusterNodegroupOpts struct {
	Count    int64
	MinCount *int64
	MaxCount *int64
}

type UpdateClusterNodegroupErrorResponse struct {
	Error UpdateClusterNodegroupErrorError `json:"Error"`
}

type UpdateClusterNodegroupErrorError struct {
	Message         string                           `json:"message"`
	ValidationError *ClusterNodegroupValidationError `json:"validationError,omitempty"`
}

type ClusterNodegroupValidationError struct {
	Errors []string `json:"errors"`
}

func (c *VendorV3Client) UpdateClusterNodegroup(clusterID string, nodegroupID string, opts UpdateClusterNodegroupOpts) (*types.Cluster, *UpdateClusterNodegroupErrorError, error) {
	req := UpdateClusterNodegroupRequest{
		Count:    opts.Count,
		MinCount: opts.MinCount,
		MaxCount: opts.MaxCount,
	}

	return c.doUpdateClusterNodegroupRequest(clusterID, nodegroupID, req)
}

func (c *VendorV3Client) doUpdateClusterNodegroupRequest(clusterID string, nodegroupID string, req UpdateClusterNodegroupRequest) (*types.Cluster, *UpdateClusterNodegroupErrorError, error) {
	resp := UpdateClusterNodegroupResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/nodegroup/%s", clusterID, nodegroupID)
	err := c.DoJSON("PUT", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				veResp := &UpdateClusterNodegroupErrorResponse{}
				if jsonErr := json.Unmarshal(apiErr.Body, veResp); jsonErr != nil {
					return nil, nil, fmt.Errorf("unmarshal validation error response: %w", err)
				}
				return nil, &veResp.Error, nil
			}
		}

		return nil, nil, err
	}

	return resp.Cluster, nil, nil
}

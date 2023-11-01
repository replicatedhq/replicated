package kotsclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterNodeGroupRequest struct {
	Name         string `json:"name"`
	NodeCount    int    `json:"node_count"`
	DiskGiB      int64  `json:"disk_gib"`
	InstanceType string `json:"instance_type"`
}

type CreateClusterNodeGroupResponse struct {
	ClusterNodeGroup *types.ClusterNodeGroup `json:"clusterNodeGroup"`
	Errors           []string                `json:"errors"`
}

type CreateClusterNodeGroupOpts struct {
	ClusterID    string
	Name         string
	NodeCount    int
	DiskGiB      int64
	InstanceType string
}

type CreateClusterNodeGroupErrorResponse struct {
	Error []string `json:"errors"`
}

type CreateClusterNodeGroupError struct {
	Message         string                           `json:"message"`
	MaxDiskGiB      int64                            `json:"maxDiskGiB,omitempty"`
	ValidationError *ClusterNodeGroupValidationError `json:"validationError,omitempty"`
}

type ClusterNodeGroupValidationError struct {
	Errors                 []string `json:"errors"`
	SupportedInstanceTypes []string `json:"supportedInstanceTypes"`
}

func (c *VendorV3Client) CreateClusterNodeGroup(clusterID string, opts CreateClusterNodeGroupOpts) (*types.ClusterNodeGroup, *CreateClusterNodeGroupError, error) {
	req := CreateClusterNodeGroupRequest{
		Name:         opts.Name,
		NodeCount:    opts.NodeCount,
		InstanceType: opts.InstanceType,
		DiskGiB:      opts.DiskGiB,
	}

	// all defaults are applied in the

	return c.doCreateClusterNodeGroupRequest(clusterID, req)
}

func (c *VendorV3Client) doCreateClusterNodeGroupRequest(clusterID string, req CreateClusterNodeGroupRequest) (*types.ClusterNodeGroup, *CreateClusterNodeGroupError, error) {
	resp := CreateClusterNodeGroupResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/nodegroup", clusterID)

	err := c.DoJSON("POST", endpoint, http.StatusCreated, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				veResp := &CreateClusterNodeGroupErrorResponse{}
				if jsonErr := json.Unmarshal(apiErr.Body, veResp); jsonErr != nil {
					return nil, nil, fmt.Errorf("unmarshal validation error response: %w", err)
				}

				return nil, &CreateClusterNodeGroupError{
					Message: veResp.Error[0],
					ValidationError: &ClusterNodeGroupValidationError{
						Errors: veResp.Error,
					},
				}, nil
			}
		}

		return nil, nil, err
	}

	return resp.ClusterNodeGroup, nil, nil
}

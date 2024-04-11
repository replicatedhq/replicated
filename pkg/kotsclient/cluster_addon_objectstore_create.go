package kotsclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterAddonObjectStoreOpts struct {
	ClusterID string
	Bucket    string
	DryRun    bool
}

type CreateClusterAddonObjectStoreRequest struct {
	Bucket string `json:"bucket"`
}

type CreateClusterAddonObjectStoreResponse struct {
	Addon *types.ClusterAddon `json:"addon"`
}

type CreateClusterAddonErrorResponse struct {
	Message string `json:"message"`
}

func (c *VendorV3Client) CreateClusterAddonObjectStore(opts CreateClusterAddonObjectStoreOpts) (*types.ClusterAddon, error) {
	req := CreateClusterAddonObjectStoreRequest{
		Bucket: opts.Bucket,
	}
	return c.doCreateClusterAddonObjectStoreRequest(opts.ClusterID, req, opts.DryRun)
}

func (c *VendorV3Client) doCreateClusterAddonObjectStoreRequest(clusterID string, req CreateClusterAddonObjectStoreRequest, dryRun bool) (*types.ClusterAddon, error) {
	resp := CreateClusterAddonObjectStoreResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/addons/objectstore", clusterID)
	if dryRun {
		endpoint = fmt.Sprintf("%s?dry-run=true", endpoint)
	}
	err := c.DoJSON("POST", endpoint, http.StatusCreated, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				errResp := &CreateClusterAddonErrorResponse{}
				if jsonErr := json.Unmarshal(apiErr.Body, errResp); jsonErr != nil {
					return nil, fmt.Errorf("unmarshal error response: %w", err)
				}
				return nil, errors.New(errResp.Message)
			}
		}

		return nil, err
	}

	return resp.Addon, nil
}

package kotsclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterAddonPostgresOpts struct {
	ClusterID    string
	Version      string
	DiskSize     int64
	InstanceType string
	DryRun       bool
}

type CreateClusterAddonPostgresRequest struct {
	Version      string `json:"version"`
	DiskSize     int64  `json:"diskSize"`
	InstanceType string `json:"instanceType"`
}

func (c *VendorV3Client) CreateClusterAddonPostgres(opts CreateClusterAddonPostgresOpts) (*types.ClusterAddon, error) {
	req := CreateClusterAddonPostgresRequest{
		Version:      opts.Version,
		DiskSize:     opts.DiskSize,
		InstanceType: opts.InstanceType,
	}
	return c.doCreateClusterAddonPostgresRequest(opts.ClusterID, req, opts.DryRun)
}

func (c *VendorV3Client) doCreateClusterAddonPostgresRequest(clusterID string, req CreateClusterAddonPostgresRequest, dryRun bool) (*types.ClusterAddon, error) {
	resp := CreateClusterAddonObjectStoreResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/addons/postgres", clusterID)
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
				return nil, errors.New(errResp.Error)
			}
		}

		return nil, err
	}

	return resp.Addon, nil
}

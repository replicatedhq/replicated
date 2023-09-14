package kotsclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type UpgradeClusterRequest struct {
	KubernetesVersion string `json:"kubernetes_version"`
}

type UpgradeClusterResponse struct {
	Cluster                *types.Cluster    `json:"cluster"`
	Errors                 []string          `json:"errors"`
	SupportedDistributions map[string]string `json:"supported_distributions"`
}

type UpgradeClusterOpts struct {
	KubernetesVersion string
	DryRun            bool
}

func (c *VendorV3Client) UpgradeCluster(clusterID string, opts UpgradeClusterOpts) (*types.Cluster, *ClusterValidationError, error) {
	req := UpgradeClusterRequest{
		KubernetesVersion: opts.KubernetesVersion,
	}

	if opts.DryRun {
		ve, err := c.doUpgradeClusterDryRunRequest(clusterID, req)
		return nil, ve, err
	}
	return c.doUpgradeClusterRequest(clusterID, req)
}

func (c *VendorV3Client) doUpgradeClusterRequest(clusterID string, req UpgradeClusterRequest) (*types.Cluster, *ClusterValidationError, error) {
	resp := UpgradeClusterResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/upgrade", clusterID)
	err := c.DoJSON("POST", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := err.(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				veResp := &ClusterValidationError{}
				if jsonErr := json.Unmarshal(apiErr.Body, veResp); jsonErr != nil {
					return nil, nil, fmt.Errorf("unmarshal validation error response: %w", err)
				}
				return nil, veResp, nil
			}
		}
		return nil, nil, err
	}

	return resp.Cluster, nil, nil
}

func (c *VendorV3Client) doUpgradeClusterDryRunRequest(clusterID string, req UpgradeClusterRequest) (*ClusterValidationError, error) {
	resp := ClusterValidationError{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/upgrade?dry-run=true", clusterID)
	err := c.DoJSON("POST", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

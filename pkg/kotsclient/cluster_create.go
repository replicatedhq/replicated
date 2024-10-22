package kotsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterRequest struct {
	Name                   string      `json:"name"`
	KubernetesDistribution string      `json:"kubernetes_distribution"`
	KubernetesVersion      string      `json:"kubernetes_version"`
	IPFamily               string      `json:"ip_family"`
	LicenseID              string      `json:"license_id"`
	NodeCount              int         `json:"node_count"`
	MinNodeCount           *int        `json:"min_node_count"`
	MaxNodeCount           *int        `json:"max_node_count"`
	DiskGiB                int64       `json:"disk_gib"`
	TTL                    string      `json:"ttl"`
	NodeGroups             []NodeGroup `json:"node_groups"`
	InstanceType           string      `json:"instance_type"`
	Tags                   []types.Tag `json:"tags"`
}

type CreateClusterResponse struct {
	Cluster                *types.Cluster    `json:"cluster"`
	Errors                 []string          `json:"errors"`
	SupportedDistributions map[string]string `json:"supported_distributions"`
}

type CreateClusterDryRunResponse struct {
	TotalCost *int64                  `json:"total_cost"`
	TTL       *string                 `json:"ttl"`
	Error     CreateClusterErrorError `json:"error"`
}

type CreateClusterOpts struct {
	Name                   string
	KubernetesDistribution string
	KubernetesVersion      string
	IPFamily               string
	LicenseID              string
	NodeCount              int
	MinNodeCount           *int
	MaxNodeCount           *int
	DiskGiB                int64
	TTL                    string
	InstanceType           string
	NodeGroups             []NodeGroup
	Tags                   []types.Tag
	DryRun                 bool
}

type NodeGroup struct {
	Name         string `json:"name"`
	InstanceType string `json:"instance_type"`
	Nodes        int    `json:"node_count"`
	MinNodes     *int   `json:"min_node_count"`
	MaxNodes     *int   `json:"max_node_count"`
	Disk         int    `json:"disk_gib"`
}

type CreateClusterErrorResponse struct {
	Error CreateClusterErrorError `json:"error"`
}

type CreateClusterErrorError struct {
	Message         string                  `json:"message"`
	MaxDiskGiB      int64                   `json:"maxDiskGiB,omitempty"`
	MaxEKS          int64                   `json:"maxEKS,omitempty"`
	MaxGKE          int64                   `json:"maxGKE,omitempty"`
	MaxAKS          int64                   `json:"maxAKS,omitempty"`
	ValidationError *ClusterValidationError `json:"validationError,omitempty"`
}

type ClusterValidationError struct {
	Errors                 []string                `json:"errors"`
	SupportedDistributions []*types.ClusterVersion `json:"supported_distributions"`
}

func (c *VendorV3Client) CreateCluster(opts CreateClusterOpts) (*types.Cluster, *CreateClusterErrorError, error) {
	req := CreateClusterRequest{
		Name:                   opts.Name,
		KubernetesDistribution: opts.KubernetesDistribution,
		KubernetesVersion:      opts.KubernetesVersion,
		IPFamily:               opts.IPFamily,
		LicenseID:              opts.LicenseID,
		NodeCount:              opts.NodeCount,
		MinNodeCount:           opts.MinNodeCount,
		MaxNodeCount:           opts.MaxNodeCount,
		DiskGiB:                opts.DiskGiB,
		TTL:                    opts.TTL,
		InstanceType:           opts.InstanceType,
		NodeGroups:             opts.NodeGroups,
		Tags:                   opts.Tags,
	}

	if opts.DryRun {
		return c.doCreateClusterDryRunRequest(req)
	}
	return c.doCreateClusterRequest(req)
}

func (c *VendorV3Client) doCreateClusterRequest(req CreateClusterRequest) (*types.Cluster, *CreateClusterErrorError, error) {
	resp := CreateClusterResponse{}
	endpoint := "/v3/cluster"
	err := c.DoJSON(context.TODO(), "POST", endpoint, http.StatusCreated, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				veResp := &CreateClusterErrorResponse{}
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

func (c *VendorV3Client) doCreateClusterDryRunRequest(req CreateClusterRequest) (*types.Cluster, *CreateClusterErrorError, error) {
	resp := CreateClusterDryRunResponse{}
	endpoint := "/v3/cluster?dry-run=true"
	err := c.DoJSON(context.TODO(), "POST", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return nil, nil, err
	}

	if resp.Error.Message != "" {
		return nil, &resp.Error, nil
	}
	cl := &types.Cluster{
		EstimatedCost: *resp.TotalCost,
		TTL:           *resp.TTL,
	}

	return cl, nil, nil
}

package kotsclient

import (
	"net/http"
	"strings"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterRequest struct {
	Name                   string `json:"name"`
	KubernetesDistribution string `json:"kubernetes_distribution"`
	KubernetesVersion      string `json:"kubernetes_version"`
	NodeCount              int    `json:"node_count"`
	DiskGiB                int64  `json:"disk_gib"`
	TTL                    string `json:"ttl"`
	InstanceType           string `json:"instance_type"`
}

type CreateClusterResponse struct {
	Cluster                *types.Cluster    `json:"cluster"`
	Errors                 []string          `json:"errors"`
	SupportedDistributions map[string]string `json:"supported_distributions"`
}

type CreateClusterOpts struct {
	Name                   string
	KubernetesDistribution string
	KubernetesVersion      string
	NodeCount              int
	DiskGiB                int64
	TTL                    string
	InstanceType           string
	DryRun                 bool
}

type ClusterValidationError struct {
	Errors                 []string                `json:"errors"`
	SupportedDistributions []*types.ClusterVersion `json:"supported_distributions"`
}

func (c *VendorV3Client) CreateCluster(opts CreateClusterOpts) (*types.Cluster, *ClusterValidationError, error) {
	req := CreateClusterRequest{
		Name:                   opts.Name,
		KubernetesDistribution: opts.KubernetesDistribution,
		KubernetesVersion:      opts.KubernetesVersion,
		NodeCount:              opts.NodeCount,
		DiskGiB:                opts.DiskGiB,
		TTL:                    opts.TTL,
		InstanceType:           opts.InstanceType,
	}

	if opts.DryRun {
		ve, err := c.doCreateClusterDryRunRequest(req)
		return nil, ve, err
	}
	return c.doCreateClusterRequest(req)
}

func (c *VendorV3Client) doCreateClusterRequest(req CreateClusterRequest) (*types.Cluster, *ClusterValidationError, error) {
	resp := CreateClusterResponse{}
	endpoint := "/v3/cluster"
	err := c.DoJSON("POST", endpoint, http.StatusCreated, req, &resp)
	if err != nil {
		if strings.Contains(err.Error(), " 400: ") {
			// to avoid a lot of brittle string parsing, we make the request again with
			// a dry-run flag and expect to get the same result back
			ve, _ := c.doCreateClusterDryRunRequest(req)
			if ve != nil && len(ve.Errors) > 0 {
				return nil, ve, nil
			}
		}
		return nil, nil, err
	}

	return resp.Cluster, nil, nil
}

func (c *VendorV3Client) doCreateClusterDryRunRequest(req CreateClusterRequest) (*ClusterValidationError, error) {
	resp := ClusterValidationError{}
	endpoint := "/v3/cluster?dry-run=true"
	err := c.DoJSON("POST", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

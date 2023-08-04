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
	DryRun                 bool
	InstanceType           string
}

type ValidationError struct {
	Errors                 []string            `json:"errors"`
	SupportedDistributions map[string][]string `json:"supported_distributions"`
}

func (c *VendorV3Client) CreateCluster(opts CreateClusterOpts) (*types.Cluster, *ValidationError, error) {
	reqBody := &CreateClusterRequest{
		Name:                   opts.Name,
		KubernetesDistribution: opts.KubernetesDistribution,
		KubernetesVersion:      opts.KubernetesVersion,
		NodeCount:              opts.NodeCount,
		DiskGiB:                opts.DiskGiB,
		TTL:                    opts.TTL,
		InstanceType:           opts.InstanceType,
	}

	cluster := CreateClusterResponse{}
	endpoint := "/v3/cluster"
	if opts.DryRun {
		endpoint = "/v3/cluster?dry-run=true"
		ve := ValidationError{}
		err := c.DoJSON("POST", endpoint, http.StatusOK, reqBody, &ve)
		if err != nil {
			return nil, nil, err
		}

		return nil, &ve, nil
	}
	err := c.DoJSON("POST", endpoint, http.StatusCreated, reqBody, &cluster)
	if err != nil {
		if strings.Contains(err.Error(), " 400: ") {
			// to avoid a lot of brittle string parsing, we make the request again with
			// a dry-run flag and expect to get the same result back
			ve := ValidationError{}
			err := c.DoJSON("POST", "/v3/cluster?dry-run=true", http.StatusOK, reqBody, &ve)
			if err != nil {
				return nil, nil, err
			}

			return nil, &ve, nil
		}
		return nil, nil, err
	}

	return cluster.Cluster, nil, nil
}

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
	VCpus                  int64  `json:"vcpus"`
	VCpuType               string `json:"vcpu_type"`
	MemoryMiB              int64  `json:"memory_mib"`
	DiskMiB                int64  `json:"disk_mib"`
	DiskType               string `json:"disk_type"`
	TTL                    string `json:"ttl"`
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
	VCpus                  int64
	VCpuType               string
	MemoryMiB              int64
	DiskMiB                int64
	DiskType               string
	TTL                    string
	DryRun                 bool
}

type ValidationError struct {
	Errors                 []string            `json:"errors"`
	SupportedDistributions map[string][]string `json:"supported_distributions"`
}

var defaultCreateClusterOpts = CreateClusterOpts{
	Name:                   "", // server will generate
	KubernetesDistribution: "kind",
	KubernetesVersion:      "v1.25.3",
	NodeCount:              int(1),
	VCpus:                  int64(4),
	MemoryMiB:              int64(4096),
	DiskMiB:                int64(51200),
	TTL:                    "2h",
}

func (c *VendorV3Client) CreateCluster(opts CreateClusterOpts) (*types.Cluster, *ValidationError, error) {
	// merge opts with defaults
	if opts.KubernetesDistribution == "" {
		opts.KubernetesDistribution = defaultCreateClusterOpts.KubernetesDistribution
	}
	if opts.KubernetesVersion == "" {
		opts.KubernetesVersion = defaultCreateClusterOpts.KubernetesVersion
	}
	if opts.NodeCount == int(0) {
		opts.NodeCount = defaultCreateClusterOpts.NodeCount
	}
	if opts.VCpus == int64(0) {
		opts.VCpus = defaultCreateClusterOpts.VCpus
	}
	if opts.MemoryMiB == int64(0) {
		opts.MemoryMiB = defaultCreateClusterOpts.MemoryMiB
	}
	if opts.DiskMiB == int64(0) {
		opts.DiskMiB = defaultCreateClusterOpts.DiskMiB
	}
	if opts.TTL == "" {
		opts.TTL = defaultCreateClusterOpts.TTL
	}

	reqBody := &CreateClusterRequest{
		Name:                   opts.Name,
		KubernetesDistribution: opts.KubernetesDistribution,
		KubernetesVersion:      opts.KubernetesVersion,
		NodeCount:              opts.NodeCount,
		VCpus:                  opts.VCpus,
		VCpuType:               opts.VCpuType,
		MemoryMiB:              opts.MemoryMiB,
		DiskMiB:                opts.DiskMiB,
		DiskType:               opts.DiskType,
		TTL:                    opts.TTL,
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

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
	MemoryMiB              int64  `json:"memory_gib"`
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
	MemoryGiB              int64
	TTL                    string
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
	MemoryGiB:              int64(4),
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
	if opts.MemoryGiB == int64(0) {
		opts.MemoryGiB = defaultCreateClusterOpts.MemoryGiB
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
		MemoryMiB:              opts.MemoryGiB,
		TTL:                    opts.TTL,
	}
	cluster := CreateClusterResponse{}
	err := c.DoJSON("POST", "/v3/cluster", http.StatusCreated, reqBody, &cluster)
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

package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterRequest struct {
	Name                   string `json:"name"`
	KubernetesDistribution string `json:"kubernetes_distribution"`
	KubernetesVersion      string `json:"kubernetes_version"`
	NodeCount              int    `json:"node_count"`
	VCpus                  int64  `json:"vcpus"`
	MemoryMiB              int64  `json:"memory_mib"`
	TTL                    string `json:"ttl"`
}

type CreateClusterResponse struct {
	Cluster *types.Cluster `json:"cluster"`
}

type CreateClusterOpts struct {
	Name                   string
	KubernetesDistribution string
	KubernetesVersion      string
	NodeCount              int
	VCpus                  int64
	MemoryMiB              int64
	TTL                    string
}

var defaultCreateClusterOpts = CreateClusterOpts{
	Name:                   "", // server will generate
	KubernetesDistribution: "kind",
	KubernetesVersion:      "1.25.3",
	NodeCount:              int(1),
	VCpus:                  int64(4),
	MemoryMiB:              int64(4096),
	TTL:                    "2h",
}

func (c *VendorV3Client) CreateCluster(opts CreateClusterOpts) (*types.Cluster, error) {
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
	if opts.TTL == "" {
		opts.TTL = defaultCreateClusterOpts.TTL
	}

	reqBody := &CreateClusterRequest{
		Name:                   opts.Name,
		KubernetesDistribution: opts.KubernetesDistribution,
		KubernetesVersion:      opts.KubernetesVersion,
		NodeCount:              opts.NodeCount,
		VCpus:                  opts.VCpus,
		MemoryMiB:              opts.MemoryMiB,
		TTL:                    opts.TTL,
	}
	cluster := CreateClusterResponse{}
	err := c.DoJSON("POST", "/v3/cluster", http.StatusCreated, reqBody, &cluster)
	if err != nil {
		return nil, err
	}
	return cluster.Cluster, nil
}

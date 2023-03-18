package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterRequest struct {
	Name         string `json:"name"`
	Distribution string `json:"distribution"`
	Version      string `json:"version"`
	TTL          string `json:"ttl"`
}

type CreateClusterResponse struct {
	Cluster *types.Cluster `json:"cluster"`
}

type CreateClusterOpts struct {
	Name         string
	Distribution string
	Version      string
	TTL          string
}

var defaultCreateClusterOpts = CreateClusterOpts{
	Name:         "", // server will generate
	Distribution: "kind",
	Version:      "1.23.0",
	TTL:          "2h",
}

func (c *VendorV3Client) CreateCluster(opts CreateClusterOpts) (*types.Cluster, error) {
	// merge opts with defaults
	if opts.Distribution == "" {
		opts.Distribution = defaultCreateClusterOpts.Distribution
	}
	if opts.Version == "" {
		opts.Version = defaultCreateClusterOpts.Version
	}
	if opts.TTL == "" {
		opts.TTL = defaultCreateClusterOpts.TTL
	}

	reqBody := &CreateClusterRequest{
		Name:         opts.Name,
		Distribution: opts.Distribution,
		Version:      opts.Version,
		TTL:          opts.TTL,
	}
	cluster := CreateClusterResponse{}
	err := c.DoJSON("POST", "/v3/cluster", http.StatusCreated, reqBody, &cluster)
	if err != nil {
		return nil, err
	}
	return cluster.Cluster, nil
}

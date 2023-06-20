package types

import "time"

type Cluster struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	KubernetesDistribution string `json:"kubernetes_distribution"`
	KubernetesVersion      string `json:"kubernetes_version"`
	NodeCount              int    `json:"node_count"`
	VCpus                  int64  `json:"vcpus"`
	VCpuType               string `json:"vcpu_type"`
	MemoryMiB              int64  `json:"memory_mib"`
	DiskGiB                int64  `json:"disk_gib"`
	DiskType               string `json:"disk_type"`

	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreditRate float64   `json:"credit_rate"`
}

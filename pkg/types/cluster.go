package types

import "time"

type Cluster struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	OSDistribution         string    `json:"os_distribution"`
	OSVersion              string    `json:"os_version"`
	KubernetesDistribution string    `json:"kubernetes_distribution"`
	KubernetesVersion      string    `json:"kubernetes_version"`
	NodeCount              int       `json:"node_count"`
	VCpus                  int64     `json:"vcpus"`
	MemoryMiB              int64     `json:"memory_mib"`
	Status                 string    `json:"status"`
	CreatedAt              time.Time `json:"created_at"`
	ExpiresAt              time.Time `json:"expires_at"`
}

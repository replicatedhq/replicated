package types

import "time"

type Cluster struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	KubernetesDistribution string `json:"kubernetes_distribution"`
	KubernetesVersion      string `json:"kubernetes_version"`
	NodeCount              int    `json:"node_count"`
	DiskGiB                int64  `json:"disk_gib"`

	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ClusterVersion struct {
	Name     string   `json:"short_name"`
	Versions []string `json:"versions"`
}

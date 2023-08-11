package types

import "time"

type ClusterStatus string

const (
	ClusterStatusQueued       ClusterStatus = "queued"       // Not assigned to a runner yet
	ClusterStatusAssigned     ClusterStatus = "assigned"     // Assigned to a runner, but have not heard back from the runner
	ClusterStatusPreparing    ClusterStatus = "preparing"    // The runner sets this when is receives the request
	ClusterStatusProvisioning ClusterStatus = "provisioning" // The runner sets this when it starts provisioning
	ClusterStatusRunning      ClusterStatus = "running"      // The runner sets this when it is done provisioning and available
	ClusterStatusTerminated   ClusterStatus = "terminated"   // This is set when the cluster expires or is deleted
	ClusterStatusError        ClusterStatus = "error"        // Something unexpected
	ClusterStatusDeleted      ClusterStatus = "deleted"
)

type Cluster struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	KubernetesDistribution string `json:"kubernetes_distribution"`
	KubernetesVersion      string `json:"kubernetes_version"`
	NodeCount              int    `json:"node_count"`
	DiskGiB                int64  `json:"disk_gib"`

	Status    ClusterStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
}

type ClusterVersion struct {
	Name     string   `json:"short_name"`
	Versions []string `json:"versions"`
}

package types

import "time"

type ClusterStatus string

const (
	ClusterStatusQueued       ClusterStatus = "queued"        // Not assigned to a runner yet
	ClusterStatusAssigned     ClusterStatus = "assigned"      // Assigned to a runner, but have not heard back from the runner
	ClusterStatusPreparing    ClusterStatus = "preparing"     // The runner sets this when is receives the request
	ClusterStatusProvisioning ClusterStatus = "provisioning"  // The runner sets this when it starts provisioning
	ClusterStatusRunning      ClusterStatus = "running"       // The runner sets this when it is done provisioning or upgrading and available
	ClusterStatusTerminated   ClusterStatus = "terminated"    // This is set when the cluster expires or is deleted
	ClusterStatusError        ClusterStatus = "error"         // Something unexpected
	ClusterStatusUpgrading    ClusterStatus = "upgrading"     // The runner sets this when it starts upgrading
	ClusterStatusUpgradeError ClusterStatus = "upgrade_error" // Something unexpected during an upgrade
	ClusterStatusDeleted      ClusterStatus = "deleted"
)

type Cluster struct {
	ID                     string       `json:"id"`
	Name                   string       `json:"name"`
	KubernetesDistribution string       `json:"kubernetes_distribution"`
	KubernetesVersion      string       `json:"kubernetes_version"`
	NodeGroups             []*NodeGroup `json:"node_groups"`

	Status    ClusterStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`

	Tags []Tag `json:"tags"`
}

type NodeGroup struct {
	ID           string `json:"id"`
	ClusterID    string `json:"cluster_id"`
	IsDefault    bool   `json:"is_default"`
	InstanceType string `json:"instance_type"`

	Name      string `json:"name"`
	NodeCount int    `json:"node_count"`
	DiskGiB   int64  `json:"disk_gib"`

	CreatedAt      time.Time  `json:"created_at"`
	ProvisioningAt *time.Time `json:"-"`
	RunningAt      *time.Time `json:"running_at"`
	CreditsPerHour int64      `json:"credits_per_hour"`

	TotalCredits  int64 `json:"total_credits,omitempty"` // this is only present after the cluster is stopped
	MinutesBilled int64 `json:"minutes_billed"`
}

type ClusterDistributionStatus struct {
	Enabled       bool   `json:"enabled"`
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
}

type ClusterVersion struct {
	Name          string                     `json:"short_name"`
	Versions      []string                   `json:"versions"`
	InstanceTypes []string                   `json:"instance_types"`
	NodesMax      int                        `json:"nodes_max"`
	Status        *ClusterDistributionStatus `json:"status,omitempty"`
}

type AddOnState string

const (
	AddOnStatePending  AddOnState = "pending"  // No attempts to install this ingress yet
	AddOnStateApplied  AddOnState = "applied"  // The ingress has been applied to the cluster
	AddOnStateRunning  AddOnState = "ready"    // The ingress is ready to be used
	AddOnStateError    AddOnState = "error"    // The ingress has an error
	AddOnStateRemoving AddOnState = "removing" // The ingress is being removed
	AddOnStateRemoved  AddOnState = "removed"  // The ingress has been removed
)

type ClusterAddOn struct {
	ID    string     `json:"id"`
	State AddOnState `json:"state"`

	Ingress *ClusterIngressAddOn `json:"ingress,omitempty"`
}

type ClusterIngressAddOn struct {
	Hostname  string `json:"hostname"`
	Target    string `json:"target"`
	Namespace string `json:"namespace"`
}

type ClusterExposedPort struct {
	Protocol    string `json:"protocol"`
	ExposedPort int    `json:"exposed_port"`
}

type ClusterPort struct {
	UpstreamPort int                  `json:"upstream_port"`
	ExposedPorts []ClusterExposedPort `json:"exposed_ports"`
	CreatedAt    time.Time            `json:"created_at"`
	Hostname     string               `json:"hostname"`
	State        AddOnState           `json:"state"`
}

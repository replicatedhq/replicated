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

	TTL string `json:"ttl"`

	CreditsPerHourPerCluster int64 `json:"credits_per_hour_per_cluster"`
	FlatFee                  int64 `json:"flat_fee"`
	TotalCredits             int64 `json:"total_credits"`

	EstimatedCost int64 `json:"estimated_cost"` // Represents estimated credits for this cluster based on the TTL

	OverlayEndpoint string `json:"overlay_endpoint,omitempty"`
	OverlayToken    string `json:"overlay_token,omitempty"`

	Tags []Tag `json:"tags"`
}

type NodeGroup struct {
	ID           string `json:"id"`
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

type ClusterAddonStatus string

const (
	ClusterAddonStatusPending  ClusterAddonStatus = "pending"  // No attempts to install this addon
	ClusterAddonStatusApplied  ClusterAddonStatus = "applied"  // The addon has been applied to the cluster
	ClusterAddonStatusRunning  ClusterAddonStatus = "ready"    // The addon is ready to be used
	ClusterAddonStatusError    ClusterAddonStatus = "error"    // The addon has an error
	ClusterAddonStatusRemoving ClusterAddonStatus = "removing" // The addon is being removed
	ClusterAddonStatusRemoved  ClusterAddonStatus = "removed"  // The addon has been removed
)

type ClusterAddon struct {
	ID        string             `json:"id"`
	ClusterID string             `json:"cluster_id"`
	Status    ClusterAddonStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`

	ObjectStore *ClusterAddonObjectStore `json:"object_store,omitempty"`
	Postgres    *ClusterAddonPostgres    `json:"postgres,omitempty"`
}

type ClusterAddonObjectStore struct {
	BucketPrefix               string `json:"bucket_prefix"`
	BucketName                 string `json:"bucket_name,omitempty"`
	ServiceAccountNamespace    string `json:"service_account_namespace,omitempty"`
	ServiceAccountName         string `json:"service_account_name,omitempty"`
	ServiceAccountNameReadOnly string `json:"service_account_name_read_only,omitempty"`
}

type ClusterAddonPostgres struct {
	Version      string `json:"version"`
	DiskGiB      int64  `json:"disk_gib"`
	InstanceType string `json:"instance_type"`

	URI string `json:"uri,omitempty"`
}

func (addon *ClusterAddon) TypeName() string {
	switch {
	case addon.ObjectStore != nil:
		return "Object Store"
	case addon.Postgres != nil:
		return "Postgres"
	default:
		return "Unknown"
	}
}

type ClusterExposedPort struct {
	Protocol    string `json:"protocol"`
	ExposedPort int    `json:"exposed_port"`
}

type ClusterPort struct {
	ClusterID    string               `json:"cluster_id"`
	AddonID      string               `json:"addon_id"`
	UpstreamPort int                  `json:"upstream_port"`
	ExposedPorts []ClusterExposedPort `json:"exposed_ports"`
	IsWildcard   bool                 `json:"is_wildcard"`
	CreatedAt    time.Time            `json:"created_at"`
	Hostname     string               `json:"hostname"`
	PortName     string               `json:"port_name"`
	State        ClusterAddonStatus   `json:"state"`
}

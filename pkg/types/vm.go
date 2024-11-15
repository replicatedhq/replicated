package types

import "time"

type VM struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Distribution string `json:"distribution"`
	Version      string `json:"version"`

	Status    VMStatus  `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`

	TTL string `json:"ttl"`

	CreditsPerHour int64 `json:"credits_per_hour"`
	FlatFee        int64 `json:"flat_fee"`
	TotalCredits   int64 `json:"total_credits"`

	EstimatedCost int64 `json:"estimated_cost"` // Represents estimated credits for this vm based on the TTL

	DirectSSHPort     int64  `json:"direct_ssh_port,omitempty"`
	DirectSSHEndpoint string `json:"direct_ssh_endpoint,omitempty"`

	Tags []Tag `json:"tags"`
}

type VMStatus string

const (
	VMStatusQueued       VMStatus = "queued"       // Not assigned to a runner yet
	VMStatusAssigned     VMStatus = "assigned"     // Assigned to a runner, but have not heard back from the runner
	VMStatusPreparing    VMStatus = "preparing"    // The runner sets this when is receives the request
	VMStatusProvisioning VMStatus = "provisioning" // The runner sets this when it starts provisioning
	VMStatusRunning      VMStatus = "running"      // The runner sets this when it is done provisioning or upgrading and available
	VMStatusTerminated   VMStatus = "terminated"   // This is set when the cluster expires or is deleted
	VMStatusError        VMStatus = "error"        // Something unexpected
	VMStatusDeleted      VMStatus = "deleted"
)

type VMVersion struct {
	Name          string                     `json:"short_name"`
	Versions      []string                   `json:"versions"`
	InstanceTypes []string                   `json:"instance_types"`
	Status        *ClusterDistributionStatus `json:"status,omitempty"`
}

type VMAddonStatus string

const (
	VMAddonStatusPending  VMAddonStatus = "pending"  // No attempts to install this addon
	VMAddonStatusApplied  VMAddonStatus = "applied"  // The addon has been applied to the vm
	VMAddonStatusRunning  VMAddonStatus = "ready"    // The addon is ready to be used
	VMAddonStatusError    VMAddonStatus = "error"    // The addon has an error
	VMAddonStatusRemoving VMAddonStatus = "removing" // The addon is being removed
	VMAddonStatusRemoved  VMAddonStatus = "removed"  // The addon has been removed
)

type VMExposedPort struct {
	Protocol    string `json:"protocol"`
	ExposedPort int    `json:"exposed_port"`
}

type VMPort struct {
	VMID         string          `json:"vm_id"`
	AddonID      string          `json:"addon_id"`
	UpstreamPort int             `json:"upstream_port"`
	ExposedPorts []VMExposedPort `json:"exposed_ports"`
	IsWildcard   bool            `json:"is_wildcard"`
	CreatedAt    time.Time       `json:"created_at"`
	Hostname     string          `json:"hostname"`
	PortName     string          `json:"port_name"`
	State        VMAddonStatus   `json:"state"`
}

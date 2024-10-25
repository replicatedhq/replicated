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

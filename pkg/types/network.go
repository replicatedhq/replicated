package types

import "time"

type Network struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Status    NetworkStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`

	TTL string `json:"ttl"`

	OverlayEndpoint string `json:"overlay_endpoint,omitempty"`
	OverlayToken    string `json:"overlay_token,omitempty"`
}

type NetworkStatus string

const (
	StatusQueued       NetworkStatus = "queued"       // Not assigned to a runner yet
	StatusAssigned     NetworkStatus = "assigned"     // Assigned to a runner, but have not heard back from the runner
	StatusPreparing    NetworkStatus = "preparing"    // The runner sets this when is receives the request
	StatusProvisioning NetworkStatus = "provisioning" // The runner sets this when it starts provisioning
	StatusVerifying    NetworkStatus = "verifying"    // The runner sets this when it is done provisioning and available
	StatusRunning      NetworkStatus = "running"      // The runner sets this when it is done verifying and available
	StatusDeleting     NetworkStatus = "deleting"     // The runner sets this when it is deleting the network
	StatusDeleted      NetworkStatus = "deleted"      // The runner sets this when it has deleted the network
	StatusTerminated   NetworkStatus = "terminated"   // This is set when the vm is moved to the history table
	StatusError        NetworkStatus = "error"        // Something unexpected
)

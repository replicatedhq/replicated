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
	NetworkStatusQueued       NetworkStatus = "queued"       // Not assigned to a runner yet
	NetworkStatusAssigned     NetworkStatus = "assigned"     // Assigned to a runner, but have not heard back from the runner
	NetworkStatusPreparing    NetworkStatus = "preparing"    // The runner sets this when is receives the request
	NetworkStatusProvisioning NetworkStatus = "provisioning" // The runner sets this when it starts provisioning
	NetworkStatusVerifying    NetworkStatus = "verifying"    // The runner sets this when it is done provisioning and available
	NetworkStatusRunning      NetworkStatus = "running"      // The runner sets this when it is done verifying and available
	NetworkStatusDeleting     NetworkStatus = "deleting"     // The runner sets this when it is deleting the network
	NetworkStatusDeleted      NetworkStatus = "deleted"      // The runner sets this when it has deleted the network
	NetworkStatusTerminated   NetworkStatus = "terminated"   // This is set when the vm is moved to the history table
	NetworkStatusError        NetworkStatus = "error"        // Something unexpected
)

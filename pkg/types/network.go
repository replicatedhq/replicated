package types

import "time"

type Network struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Status               NetworkStatus `json:"status"`
	LastSchedulingStatus string        `json:"last_scheduling_status"`

	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`

	TTL string `json:"ttl"`

	OverlayEndpoint string `json:"overlay_endpoint,omitempty"`
	OverlayToken    string `json:"overlay_token,omitempty"`

	Outbound string `json:"outbound,omitempty"`
	Policy   string `json:"policy,omitempty"`
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

type NetworkReport struct {
	Events []*NetworkEvent `json:"events"`
}

type NetworkEvent struct {
	CreatedAt time.Time `json:"created_at"`
	EventData string    `json:"event_data"`
}

type NetworkEventData struct {
	Timestamp     string `json:"timestamp"`
	SrcIP         string `json:"srcIp"`
	DstIP         string `json:"dstIp"`
	SrcPort       int    `json:"srcPort"`
	DstPort       int    `json:"dstPort"`
	Protocol      string `json:"proto"`
	Command       string `json:"comm"`
	PID           int    `json:"pid"`
	LikelyService string `json:"likelyService"`
}

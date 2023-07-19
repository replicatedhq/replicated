package types

import "time"

// CustomHostname represents a custom hostname in cloudflare for a team
type CustomHostname struct {
	TeamID                     string    `json:"team_id"`
	OriginServer               string    `json:"origin_server"`
	Hostname                   string    `json:"hostname"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
	DomainVerificationType     string    `json:"domain_verification_type"`
	DomainVerificationStatus   string    `json:"domain_verification_status"`
	DomainTxtRecordName        string    `json:"domain_txt_record_name"`
	DomainTxtRecordValue       string    `json:"domain_txt_record_value"`
	TLSVerificationType        string    `json:"tls_verification_type"`
	TLSVerificationStatus      string    `json:"tls_verification_status"`
	TLSTxtRecordName           string    `json:"tls_txt_record_name"`
	TLSTxtRecordValue          string    `json:"tls_txt_record_value"`
	CloudflareCustomHostnameID string    `json:"cloudflare_custom_hostname_id"`
	CloudflareWorkerRouteID    string    `json:"cloudflare_worker_route_id,omitempty"`
	VerificationErrors         []string  `json:"verification_errors"`
	FailureCount               int       `json:"failure_count"`
	FailureReason              string    `json:"failure_reason"`
}

// KotsAppCustomHostname represents a custom hostname configured for a kots app
type KotsAppCustomHostname struct {
	AppID     string `json:"app_id"`
	IsDefault bool   `json:"is_default"`
	CustomHostname
}

// KotsAppCustomHostnames all custom hostnames configured for a kots app
type KotsAppCustomHostnames struct {
	Registry       []KotsAppCustomHostname `json:"registry"`
	Proxy          []KotsAppCustomHostname `json:"proxy"`
	DownloadPortal []KotsAppCustomHostname `json:"downloadPortal"`
	ReplicatedApp  []KotsAppCustomHostname `json:"replicatedApp"`
}

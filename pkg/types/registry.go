package types

import "time"

type Registry struct {
	Provider   string     `json:"provider"`
	Endpoint   string     `json:"endpoint"`
	Slug       string     `json:"slug"`
	AuthType   string     `json:"authType"`
	Username   string     `json:"username"`
	AppIds     []string   `json:"appIds,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

type RegistryLog struct {
	CreatedAt string  `json:"createdAt"`
	Action    string  `json:"action"`
	Status    *int    `json:"statusCode"`
	Success   bool    `json:"isSuccess"`
	Image     *string `json:"image"`
}

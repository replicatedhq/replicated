package types

import "time"

type Registry struct {
	Provider   string     `json:"provider"`
	Endpoint   string     `json:"endpoint"`
	AuthType   string     `json:"authType"`
	Username   string     `json:"username"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

type RegistryLog struct {
	CreatedAt string  `json:"createdAt"`
	Action    string  `json:"action"`
	Status    *int    `json:"statusCode"`
	Success   bool    `json:"isSuccess"`
	Image     *string `json:"image"`
}

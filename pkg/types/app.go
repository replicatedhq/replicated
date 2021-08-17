package types

import "time"

type App struct {
	ID        string
	Name      string
	Scheduler string
	Slug      string
}

type AppAndChannels struct {
	App      *App
	Channels []Channel
}

type KotsAppWithChannels struct {
	Channels    []Channel `json:"channels,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Description string    `json:"description,omitempty"`
	Id          string    `json:"id,omitempty"`
	IsArchived  bool      `json:"isArchived,omitempty"`
	IsKotsApp   bool      `json:"isKotsApp,omitempty"`
	Name        string    `json:"name,omitempty"`
	RenamedAt   time.Time `json:"renamedAt,omitempty"`
	Slug        string    `json:"slug,omitempty"`
	TeamId      string    `json:"teamId,omitempty"`
}

type KotsAppChannel struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

package types

import "time"

type App struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Scheduler    string `json:"scheduler"`
	Slug         string `json:"slug"`
	IsFoundation bool   `json:"isFoundation"`
}

type AppAndChannels struct {
	App      *App      `json:"app"`
	Channels []Channel `json:"channels"`
}

type KotsAppWithChannels struct {
	Channels     []Channel `json:"channels,omitempty"`
	Created      time.Time `json:"created,omitempty"`
	Description  string    `json:"description,omitempty"`
	Id           string    `json:"id,omitempty"`
	IsArchived   bool      `json:"isArchived,omitempty"`
	IsKotsApp    bool      `json:"isKotsApp,omitempty"`
	IsFoundation bool      `json:"isFoundation,omitempty"`
	Name         string    `json:"name,omitempty"`
	RenamedAt    time.Time `json:"renamedAt,omitempty"`
	Slug         string    `json:"slug,omitempty"`
	TeamId       string    `json:"teamId,omitempty"`
}

type KotsAppChannel struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

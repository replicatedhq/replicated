package types

import "time"

type CollectorSpec struct {
	ID         string    `json:"id"`
	Spec       string    `json:"spec"`
	Name       string    `json:"name"`
	AppID      string    `json:"appId"`
	IsArchived bool      `json:"isArchived"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Channels   []Channel `json:"channels"`
}

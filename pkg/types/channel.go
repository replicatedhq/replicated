package types

type Channel struct {
	ID          string
	Name        string
	Description string
	Slug        string

	ReleaseSequence int64
	ReleaseLabel    string

	IsArchived bool `json:"isArchived"`

	InstallCommands *InstallCommands
}

func (c *Channel) Copy() *Channel {
	return &Channel{
		ID:              c.ID,
		Name:            c.Name,
		Description:     c.Description,
		Slug:            c.Slug,
		ReleaseSequence: c.ReleaseSequence,
		ReleaseLabel:    c.ReleaseLabel,
		InstallCommands: c.InstallCommands,
	}
}

type InstallCommands struct {
	Existing string
	Embedded string
	Airgap   string
}

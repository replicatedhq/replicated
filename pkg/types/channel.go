package types

type Channel struct {
	ID          string
	Name        string
	Description string
	Slug        string

	ReleaseSequence int64
	ReleaseLabel    string

	InstallCommands *InstallCommands
}

type InstallCommands struct {
	Existing string
	Embedded string
	Airgap   string
}

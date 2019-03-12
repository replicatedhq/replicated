package types

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

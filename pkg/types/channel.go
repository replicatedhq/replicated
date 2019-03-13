package types

type Channel struct {
	ID          string
	Name        string
	Description string

	ReleaseSequence int64
	ReleaseLabel    string
}

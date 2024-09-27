package types

type VM struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	Status ClusterStatus `json:"status"`
}

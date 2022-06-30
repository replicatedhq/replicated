package types

type KotsSingleSpec struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Content  string           `json:"content"`
	Children []KotsSingleSpec `json:"children"`
}

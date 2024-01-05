package types

import "strings"

type Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Used for template evaluation
func (t Tag) String() string {
	return t.Key + "=" + t.Value
}

type Tags []Tag

// Used for template evaluation
func (t Tags) String() string {
	tags := []string{}
	for _, tag := range t {
		tags = append(tags, tag.String())
	}
	return strings.Join(tags, ",")
}

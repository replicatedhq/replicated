package print

import (
	"text/template"
	"time"
)

var funcs = template.FuncMap{
	// Use RFC 3339 for standard time printing in all output
	"time": func(t time.Time) string {
		return t.Format(time.RFC3339)
	},
}

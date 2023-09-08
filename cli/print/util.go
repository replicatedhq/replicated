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
	"padding": func(s string, width int) string {
		for len(s) < width {
			s += " "
		}
		return s
	},
}

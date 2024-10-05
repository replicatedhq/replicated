package print

import (
	"fmt"
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
	"add": func(a, b int) int {
		return a + b
	},
	"formatURL": func(protocol, hostname string) string {
		return fmt.Sprintf("%s://%s", protocol, hostname)
	},
	"localeTime": func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		return t.Local().Format("2006-01-02 15:04 MST")
	},
}

package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var customersTmplSrc = `ID	NAME	CHANNELS	EXPIRES	TYPE
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{range .Channels}}{{.Name}}{{end}}	{{if not .Expires}}Never{{else}}{{.Expires}}{{end}}	{{.Type}}
{{ end }}`

var customersTmpl = template.Must(template.New("channels").Parse(customersTmplSrc))

func Customers(w *tabwriter.Writer, customers []types.Customer) error {
	if err := customersTmpl.Execute(w, customers); err != nil {
		return err
	}
	return w.Flush()
}

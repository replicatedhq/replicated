package print

import (
	"encoding/json"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var customersTmplSrc = `ID	NAME	CHANNELS	EXPIRES	TYPE
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{range .Channels}} {{.Name}}{{end}}	{{if not .Expires}}Never{{else}}{{.Expires}}{{end}}	{{.Type}}
{{ end }}`

var customersTmpl = template.Must(template.New("channels").Parse(customersTmplSrc))

func Customers(outputFormat string, w *tabwriter.Writer, customers []types.Customer) error {
	if outputFormat == "table" {
		if err := customersTmpl.Execute(w, customers); err != nil {
			return err
		}
		return w.Flush()
	}
	if outputFormat == "json" {
		defer w.Flush()
		var cAsByte []byte
		if len(customers) == 1 {
			cAsByte, _ = json.MarshalIndent(customers[0], "", "  ")
		} else {
			cAsByte, _ = json.MarshalIndent(customers, "", "  ")
		}
		_, err := w.Write(cAsByte)
		return err
	}
	return nil
}

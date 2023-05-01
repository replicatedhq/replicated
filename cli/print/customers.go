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
	} else if outputFormat == "json" {
		defer w.Flush()
		var cAsByte []byte
		cAsByte, _ = json.MarshalIndent(customers, "", "  ")
		_, err := w.Write(cAsByte)
		return err
	}
	return nil
}

func Customer(outputFormat string, w *tabwriter.Writer, customer *types.Customer) error {
	if outputFormat == "table" {
		if err := customersTmpl.Execute(w, []types.Customer{*customer}); err != nil {
			return err
		}
		return w.Flush()
	} else if outputFormat == "json" {
		defer w.Flush()
		var cAsByte []byte
		cAsByte, _ = json.MarshalIndent(customer, "", "  ")
		_, err := w.Write(cAsByte)
		return err
	}
	return nil
}

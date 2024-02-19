package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

var customersTmplSrc = `ID	NAME	CHANNELS	EXPIRES	TYPE	CUSTOM_ID
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{range .Channels}} {{.Name}}{{end}}	{{if not .Expires}}Never{{else}}{{.Expires}}{{end}}	{{.Type}}	{{if not .CustomID}}Not Set{{else}}{{.CustomID}}{{end}}
{{ end }}`

var customersTmpl = template.Must(template.New("channels").Parse(customersTmplSrc))

func Customers(outputFormat string, w *tabwriter.Writer, customers []types.Customer) error {
	if outputFormat == "table" {
		if err := customersTmpl.Execute(w, customers); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(customers, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func CustomersWithInstances(outputFormat string, w *tabwriter.Writer, customers []types.Customer) error {
	// Define a new template for use in this function
	var customersTmplSrc = `CUSTOMER NAME	INSTANCE ID	LAST ACTIVE	VERSION	
{{ range . -}}
{{ .Name }}	{{range .Instances}}{{.InstanceId}}	{{.LastActive}}	{{range .VersionHistory}}{{.VersionLabel}}{{end}}{{end}}
{{ end }}`

	var customersTmpl = template.Must(template.New("customers").Parse(customersTmplSrc))

	if outputFormat == "table" {
		if err := customersTmpl.Execute(w, customers); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(customers, "", "  ")
		if _, err := w.Write(cAsByte); err != nil {
			return err
		}
	}
	return w.Flush()
}

func Customer(outputFormat string, w *tabwriter.Writer, customer *types.Customer) error {
	if outputFormat == "table" {
		if err := customersTmpl.Execute(w, []types.Customer{*customer}); err != nil {
			return errors.Wrap(err, "execute template")
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(customer, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return errors.Wrap(err, "write json")
		}
	}
	return errors.Wrap(w.Flush(), "flush")
}

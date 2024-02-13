package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var addOnsTmplHeaderSrc = `ID	TYPE 	STATUS	NOTES`
var addOnsTmplRowSrc = `{{ range . -}}
{{ .ID }}	{{ if .Ingress }}Ingress{{ else }}Other{{ end }}	{{ printf "%-12s" .State }}	{{ if .Ingress }}http[s]://{{ .Ingress.Hostname }}{{ end }}
{{ end }}`
var addOnsTmplSrc = fmt.Sprintln(addOnsTmplHeaderSrc) + addOnsTmplRowSrc
var addOnsTmpl = template.Must(template.New("ingresses").Funcs(funcs).Parse(addOnsTmplSrc))
var addOnsTmplNoHeader = template.Must(template.New("ingresses").Funcs(funcs).Parse(addOnsTmplRowSrc))

func AddOns(outputFormat string, w *tabwriter.Writer, addOns []*types.ClusterAddOn, header bool) error {
	if outputFormat == "table" {
		if header {
			if err := addOnsTmpl.Execute(w, addOns); err != nil {
				return err
			}
		} else {
			if err := addOnsTmplNoHeader.Execute(w, addOns); err != nil {
				return err
			}
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(addOns, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func AddOn(outputFormat string, w *tabwriter.Writer, addOn *types.ClusterAddOn) error {
	if outputFormat == "table" {
		if err := addOnsTmpl.Execute(w, []*types.ClusterAddOn{addOn}); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(addOn, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

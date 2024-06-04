package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var modelCollectionsTmplSrc = `ID	NAME
{{ range . -}}
{{ .ID }}	{{ .Name }}
{{ end }}`

var modelCollectionsTmpl = template.Must(template.New("registries").Funcs(funcs).Parse(modelCollectionsTmplSrc))

func ModelCollections(outputFormat string, w *tabwriter.Writer, collections []types.ModelCollection) error {
	if outputFormat == "table" {
		if err := modelCollectionsTmpl.Execute(w, collections); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(collections, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}

	return w.Flush()
}

var modelsTmplSrc = `NAME
{{ range . -}}
{{ .Name }}
{{ end }}`

var modelsTmpl = template.Must(template.New("registries").Funcs(funcs).Parse(modelsTmplSrc))

func Models(outputFormat string, w *tabwriter.Writer, models []types.Model) error {
	if outputFormat == "table" {
		if err := modelsTmpl.Execute(w, models); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		mAsByte, _ := json.MarshalIndent(models, "", "  ")
		if _, err := fmt.Fprintln(w, string(mAsByte)); err != nil {
			return err
		}
	}

	return w.Flush()
}

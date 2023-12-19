package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var appsTmplSrc = `ID	NAME	SLUG	SCHEDULER
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .Slug}}	{{ .Scheduler }}
{{ end }}`

var appsTmpl = template.Must(template.New("apps").Funcs(funcs).Parse(appsTmplSrc))

func Apps(outputFormat string, w *tabwriter.Writer, apps []types.AppAndChannels) error {
	if outputFormat == "table" {
		var as []*types.App

		for _, a := range apps {
			as = append(as, a.App)
		}

		if err := appsTmpl.Execute(w, as); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(apps, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}

	return w.Flush()
}

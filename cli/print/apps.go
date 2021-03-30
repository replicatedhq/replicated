package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var appsTmplSrc = `ID	NAME	SLUG	SCHEDULER
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .Slug}}	{{ .Scheduler }}
{{ end }}`

var appsTmpl = template.Must(template.New("apps").Funcs(funcs).Parse(appsTmplSrc))

func Apps(w *tabwriter.Writer, apps []types.AppAndChannels) error {
	var as []*types.App

	for _, a := range apps {
		as = append(as, a.App)
	}

	if err := appsTmpl.Execute(w, as); err != nil {
		return err
	}

	return w.Flush()
}

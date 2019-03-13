package print

import (
	"text/tabwriter"
	"text/template"

	apps "github.com/replicatedhq/replicated/gen/go/v1"
)

var appsTmplSrc = `ID	NAME	SCHEDULER
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .Scheduler }}
{{ end }}`

var appsTmpl = template.Must(template.New("apps").Funcs(funcs).Parse(appsTmplSrc))

func Apps(w *tabwriter.Writer, apps []apps.AppAndChannels) error {
	as := make([]map[string]interface{}, len(apps))

	for i, a := range apps {
		as[i] = map[string]interface{}{
			"ID":        a.App.Id,
			"Name":      a.App.Name,
			"Scheduler": a.App.Scheduler,
		}
	}

	if err := appsTmpl.Execute(w, as); err != nil {
		return err
	}

	return w.Flush()
}

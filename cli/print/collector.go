package print

import (
	"text/tabwriter"
	"text/template"

	collectors "github.com/replicatedhq/replicated/gen/go/v1"
)

var collectorTmplSrc = `NAME:	{{ .Name }}
CREATED:	{{ time .CreatedAt }}
ACTIVE_CHANNELS: {{ .ActiveChannels}}
CONFIG:
{{ .Config }}
`

var collectorTmpl = template.Must(template.New("Collector").Funcs(funcs).Parse(collectorTmplSrc))

func Collector(w *tabwriter.Writer, collector *collectors.AppCollector) error {
	if err := collectorTmpl.Execute(w, collector); err != nil {
		return err
	}
	return w.Flush()
}

package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var collectorTmplSrc = `NAME:	{{ .Name }}
CREATED:	{{ time .CreatedAt }}
CONFIG:
{{ .Spec }}
`

var collectorTmpl = template.Must(template.New("Collector").Funcs(funcs).Parse(collectorTmplSrc))

func Collector(w *tabwriter.Writer, collector *types.CollectorSpec) error {
	if err := collectorTmpl.Execute(w, collector); err != nil {
		return err
	}
	return w.Flush()
}

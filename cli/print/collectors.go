package print

import (
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var collectorsTmplSrc = `NAME	CREATED	ACTIVE_CHANNELS
{{ range . -}}
{{ .Name }}	{{ time .CreatedAt }}	{{ .ActiveChannels }}
{{ end }}`

var collectorsTmpl = template.Must(template.New("Collectors").Funcs(funcs).Parse(collectorsTmplSrc))

func Collectors(w *tabwriter.Writer, appCollectors []types.CollectorInfo) error {
	rs := make([]map[string]interface{}, len(appCollectors))

	for i, r := range appCollectors {
		// join active channel names like "Stable,Unstable"
		activeChans := make([]string, len(r.ActiveChannels))
		for j, activeChan := range r.ActiveChannels {
			activeChans[j] = activeChan.Name
		}
		activeChansField := strings.Join(activeChans, ",")

		rs[i] = map[string]interface{}{
			"Name":           r.Name,
			"CreatedAt":      r.CreatedAt,
			"ActiveChannels": activeChansField,
		}
	}

	if err := collectorsTmpl.Execute(w, rs); err != nil {
		return err
	}

	return w.Flush()
}

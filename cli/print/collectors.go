package print

import (
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var collectorsTmplSrc = `NAME	CREATED	 SPEC_ID 	ACTIVE_CHANNELS
{{ range . -}}
{{ .Name }}	{{ time .CreatedAt }}	{{ .SpecID }}	{{ .ActiveChannels }}
{{ end }}`

var collectorsTmpl = template.Must(template.New("Collectors").Funcs(funcs).Parse(collectorsTmplSrc))

func Collectors(w *tabwriter.Writer, appCollectors []types.CollectorSpec) error {
	rs := make([]map[string]interface{}, len(appCollectors))

	for i, r := range appCollectors {
		// join active channel names like "Stable,Unstable"
		activeChans := make([]string, len(r.Channels))
		for j, activeChan := range r.Channels {
			activeChans[j] = activeChan.Name
		}
		activeChansField := strings.Join(activeChans, ",")

		// don't print edited if it's the same as created
		// edited := r.EditedAt.Format(time.RFC3339)
		// if r.CreatedAt.Equal(r.EditedAt) {
		// 	edited = ""
		// }

		rs[i] = map[string]interface{}{
			"Name":           r.Name,
			"CreatedAt":      r.CreatedAt,
			"SpecID":         r.ID,
			"ActiveChannels": activeChansField,
		}
	}

	if err := collectorsTmpl.Execute(w, rs); err != nil {
		return err
	}

	return w.Flush()
}

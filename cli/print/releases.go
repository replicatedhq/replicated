package print

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

var releasesTmplSrc = `SEQUENCE	CREATED	EDITED	ACTIVE_CHANNELS
{{ range . -}}
{{ .Sequence }}	{{ time .CreatedAt }}	{{ .EditedAt }}	{{ .ActiveChannels }}
{{ end }}`

var releasesTmpl = template.Must(template.New("Releases").Funcs(funcs).Parse(releasesTmplSrc))

func Releases(outputFormat string, w *tabwriter.Writer, appReleases []types.ReleaseInfo) error {
	if outputFormat == "table" {
		rs := make([]map[string]interface{}, len(appReleases))

		for i, r := range appReleases {
			// join active channel names like "Stable,Unstable"
			activeChans := make([]string, len(r.ActiveChannels))
			for j, activeChan := range r.ActiveChannels {
				activeChans[j] = activeChan.Name
			}
			activeChansField := strings.Join(activeChans, ",")

			// don't print edited if it's the same as created
			edited := r.EditedAt.Format(time.RFC3339)
			if r.CreatedAt.Equal(r.EditedAt) {
				edited = ""
			}
			rs[i] = map[string]interface{}{
				"Sequence":       r.Sequence,
				"CreatedAt":      r.CreatedAt,
				"EditedAt":       edited,
				"ActiveChannels": activeChansField,
			}
		}

		if err := releasesTmpl.Execute(w, rs); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(appReleases, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}

	return w.Flush()
}

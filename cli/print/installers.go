package print

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var installersTmplSrc = `SEQUENCE	CREATED	ACTIVE_CHANNELS
{{ range . -}}
{{ .Sequence }}	{{ time .CreatedAt.Time }}	{{ .ActiveChannels }}
{{ end }}`

var installersTmpl = template.Must(template.New("Installers").Funcs(funcs).Parse(installersTmplSrc))

func Installers(outputFormat string, w *tabwriter.Writer, appReleases []types.InstallerSpec) error {
	if outputFormat == "table" {
		rs := make([]map[string]interface{}, len(appReleases))

		for i, r := range appReleases {
			// join active channel names like "Stable,Unstable"
			activeChans := make([]string, len(r.ActiveChannels))
			for j, activeChan := range r.ActiveChannels {
				activeChans[j] = activeChan.Name
			}
			activeChansField := strings.Join(activeChans, ",")

			rs[i] = map[string]interface{}{
				"Sequence":       r.Sequence,
				"CreatedAt":      r.CreatedAt,
				"ActiveChannels": activeChansField,
			}
		}

		if err := installersTmpl.Execute(w, rs); err != nil {
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

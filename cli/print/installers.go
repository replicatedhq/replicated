package print

import (
	"github.com/replicatedhq/replicated/pkg/types"
	"strings"
	"text/tabwriter"
	"text/template"
)

var installersTmplSrc = `SEQUENCE	CREATED	ACTIVE_CHANNELS
{{ range . -}}
{{ .Sequence }}	{{ time .CreatedAt.Time }}	{{ .ActiveChannels }}
{{ end }}`

var installersTmpl = template.Must(template.New("Installers").Funcs(funcs).Parse(installersTmplSrc))

func Installers(w *tabwriter.Writer, appReleases []types.InstallerSpec) error {
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

	return w.Flush()
}

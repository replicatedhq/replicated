package print

import (
	"text/tabwriter"
	"text/template"

	releases "github.com/replicatedhq/replicated/gen/go/releases"
)

var releaseTmplSrc = `SEQUENCE: {{ .Sequence }}
CREATED: {{ .CreatedAt }}
EDITED: {{ .EditedAt }}
CONFIG:
{{ .Config }}
`

var releaseTmpl = template.Must(template.New("Release").Parse(releaseTmplSrc))

func Release(w *tabwriter.Writer, release *releases.AppRelease) error {
	if err := releaseTmpl.Execute(w, release); err != nil {
		return err
	}
	return w.Flush()
}

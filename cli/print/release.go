package print

import (
	"text/tabwriter"
	"text/template"

	releases "github.com/replicatedhq/replicated/gen/go/v1"
)

var releaseTmplSrc = `SEQUENCE:	{{ .Sequence }}
CREATED:	{{ time .CreatedAt }}
EDITED:	{{ time .EditedAt }}
CONFIG:
{{ .Config }}
`

var releaseTmpl = template.Must(template.New("Release").Funcs(funcs).Parse(releaseTmplSrc))

func Release(w *tabwriter.Writer, release *releases.AppRelease) error {
	if err := releaseTmpl.Execute(w, release); err != nil {
		return err
	}
	return w.Flush()
}

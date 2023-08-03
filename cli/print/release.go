package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var releaseTmplSrc = `SEQUENCE:	{{ .Sequence }}
CREATED:	{{ time .CreatedAt }}
EDITED:	{{ time .EditedAt }}
CONFIG:
{{ .Config }}
`

var releaseTmpl = template.Must(template.New("Release").Funcs(funcs).Parse(releaseTmplSrc))

func Release(w *tabwriter.Writer, release *types.AppRelease) error {
	if err := releaseTmpl.Execute(w, release); err != nil {
		return err
	}
	return w.Flush()
}

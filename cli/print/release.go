package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var releaseTmplSrc = `SEQUENCE:	{{ .Sequence }}
CREATED:	{{ time .CreatedAt }}
EDITED:	{{ time .EditedAt }}
{{if .CompatibilityResults}}COMPATIBILITY RESULTS:
	DISTRIBUTION	VERSION	SUCCESS_AT	SUCCESS_NOTES	FAILURE_AT	FAILURE_NOTES
	{{ range .CompatibilityResults -}}
	{{ .Distribution }}	{{ .Version }}	{{if .SuccessAt}}{{ time .SuccessAt }}{{else}}-{{end}}	{{ .SuccessNotes }}	{{if .FailureAt}}{{ time .FailureAt }}{{else}}-{{end}}	{{ .FailureNotes }}
	{{ end }}{{end}}
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

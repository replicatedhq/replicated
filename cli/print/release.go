package print

import (
	"encoding/json"
	"fmt"
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

func Release(outputFormat string, w *tabwriter.Writer, release *types.AppRelease) error {
	if outputFormat == "table" {
		if err := releaseTmpl.Execute(w, release); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(release, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

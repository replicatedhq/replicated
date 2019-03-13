package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var lintTmplSrc = `RULE	TYPE	START LINE	END LINE
{{ range . -}}
{{ .Rule }}	{{ .Type }}	{{ (index .Positions 0).Start.Line }} 	{{ (index .Positions 0).End.Line }}
{{ end }}`

var lintTmpl = template.Must(template.New("lint").Parse(lintTmplSrc))

func LintErrors(w *tabwriter.Writer, lintErrors []types.LintMessage) error {
	if err := lintTmpl.Execute(w, lintErrors); err != nil {
		return err
	}
	return w.Flush()
}

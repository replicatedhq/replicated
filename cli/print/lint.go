package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var lintTmplSrc = `RULE	TYPE	LINE	MESSAGE	
{{ range . -}}
{{ .Rule }}	{{ .Type }}	{{with .Positions}}{{ (index . 0).Start.Line }}{{else}}	{{end}}	{{ .Message}}
{{ end }}`

var lintTmpl = template.Must(template.New("lint").Parse(lintTmplSrc))

func LintErrors(w *tabwriter.Writer, lintErrors []types.LintMessage) error {
	for _, lintError := range lintErrors {
		if len(lintError.Positions) > 0 {
			lintError.Positions[0].Start.Line += 1
		}
	}
	if err := lintTmpl.Execute(w, lintErrors); err != nil {
		return err
	}
	return w.Flush()
}

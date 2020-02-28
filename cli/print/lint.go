package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var lintTmplSrc = `RULE	TYPE	FILENAME	LINE	MESSAGE
{{ range . -}}
{{ .Rule }}	{{ .Type }}	{{ .Path }}	{{with .Positions}}{{ (index . 0).Start.Line }}{{else}}	{{end}}	{{ .Message}}	
{{ end }}`

var lintTmpl = template.Must(template.New("lint").Parse(lintTmplSrc))

func LintErrors(w *tabwriter.Writer, lintErrors []types.LintMessage) error {
	lintErrors = incrementLineNumbersToOneIndexed(lintErrors)
	if err := lintTmpl.Execute(w, lintErrors); err != nil {
		return err
	}
	return w.Flush()
}

// line numbers come back zero-indexed from the API
// Since we now attach messages from the yaml parser,
// lets make this 1-indexed so they match if the message includes a line number
// this is here because its primarily client logic, its about how we're displaying the data
func incrementLineNumbersToOneIndexed(lintErrors []types.LintMessage) []types.LintMessage {
	for _, lintError := range lintErrors {
		if len(lintError.Positions) > 0 {
			lintError.Positions[0].Start.Line += 1
		}
	}
	return lintErrors
}

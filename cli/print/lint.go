package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var lintTmplSrc = `RULE	TYPE	FILENAME	LINE	MESSAGE
{{ range . -}}
{{ .Rule }}	{{ .Type }}	{{ .Path }}	{{with .Positions}}{{ (index . 0).Start.Line }}{{else}}	{{end}}	{{ .Message}}	
{{ end }}`

var lintTmpl = template.Must(template.New("lint").Parse(lintTmplSrc))

func LintErrors(outputFormat string, w *tabwriter.Writer, lintErrors []types.LintMessage) error {
	if outputFormat == "table" {
		if err := lintTmpl.Execute(w, lintErrors); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(lintErrors, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

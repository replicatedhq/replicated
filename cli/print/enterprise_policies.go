package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterprisePoliciesTmplSrc = `ID	NAME
{{ range . -}}
{{ .ID }}	{{ .Name }}
{{ end }}`

var enterprisePoliciesTmpl = template.Must(template.New("enterprisepolicies").Parse(enterprisePoliciesTmplSrc))

func EnterprisePolicies(outputFormat string, w *tabwriter.Writer, policies []*enterprisetypes.Policy) error {
	if outputFormat == "table" {
		if err := enterprisePoliciesTmpl.Execute(w, policies); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(policies, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

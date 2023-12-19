package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterprisePolicyTmplSrc = `ID	NAME
{{ .ID }}	{{ .Name }}
`

var enterprisePolicyTmpl = template.Must(template.New("entrerprisepolicy").Parse(enterprisePolicyTmplSrc))

func EnterprisePolicy(outputFormat string, w *tabwriter.Writer, policy *enterprisetypes.Policy) error {
	if outputFormat == "table" {
		if err := enterprisePolicyTmpl.Execute(w, policy); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(policy, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

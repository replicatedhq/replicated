package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterprisePolicyTmplSrc = `ID	NAME	# CHANNELS
{{ range . -}}
{{ .ID }}	{{ .Name }}
{{ end }}`

var enterprisePolicyTmpl = template.Must(template.New("entrerprisepolicy").Parse(enterprisePolicyTmplSrc))

func EnterprisePolicy(w *tabwriter.Writer, policy *enterprisetypes.Policy) error {
	if err := enterprisePolicyTmpl.Execute(w, policy); err != nil {
		return err
	}
	return w.Flush()
}

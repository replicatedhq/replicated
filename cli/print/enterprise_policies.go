package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterprisePoliciesTmplSrc = `ID	NAME
{{ range . -}}
{{ .ID }}	{{ .Name }}
{{ end }}`

var enterprisePoliciesTmpl = template.Must(template.New("enterprisepolicies").Parse(enterprisePoliciesTmplSrc))

func EnterprisePolicies(w *tabwriter.Writer, policies []*enterprisetypes.Policy) error {
	if err := enterprisePoliciesTmpl.Execute(w, policies); err != nil {
		return err
	}
	return w.Flush()
}

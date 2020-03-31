package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseChannelsTmplSrc = `ID	NAME	# VENDORS
{{ range . -}}
{{ .ID }}	{{ .Name }}
{{ end }}`

var enterpriseChannelsTmpl = template.Must(template.New("enterprisechannels").Parse(enterpriseChannelsTmplSrc))

func EnterpriseChannels(w *tabwriter.Writer, channels []*enterprisetypes.Channel) error {
	if err := enterpriseChannelsTmpl.Execute(w, channels); err != nil {
		return err
	}
	return w.Flush()
}

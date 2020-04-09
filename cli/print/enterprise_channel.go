package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseChannelTmplSrc = `ID	NAME	# VENDORS	# POLICIES
{{ .ID }}	{{ .Name }}
`

var enterpriseChannelTmpl = template.Must(template.New("enterprisechannel").Parse(enterpriseChannelTmplSrc))

func EnterpriseChannel(w *tabwriter.Writer, channel *enterprisetypes.Channel) error {
	if err := enterpriseChannelTmpl.Execute(w, channel); err != nil {
		return err
	}
	return w.Flush()
}

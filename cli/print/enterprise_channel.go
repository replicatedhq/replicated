package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseChannelTmplSrc = `ID	NAME
{{ .ID }}	{{ .Name }}
`

var enterpriseChannelTmpl = template.Must(template.New("enterprisechannel").Parse(enterpriseChannelTmplSrc))

func EnterpriseChannel(outputFormat string, w *tabwriter.Writer, channel *enterprisetypes.Channel) error {
	if outputFormat == "table" {
		if err := enterpriseChannelTmpl.Execute(w, channel); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(channel, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

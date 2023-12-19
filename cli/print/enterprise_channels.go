package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseChannelsTmplSrc = `ID	NAME
{{ range . -}}
{{ .ID }}	{{ .Name }}
{{ end }}`

var enterpriseChannelsTmpl = template.Must(template.New("enterprisechannels").Parse(enterpriseChannelsTmplSrc))

func EnterpriseChannels(outputFormat string, w *tabwriter.Writer, channels []*enterprisetypes.Channel) error {
	if outputFormat == "table" {
		if err := enterpriseChannelsTmpl.Execute(w, channels); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(channels, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

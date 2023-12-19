package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	platformChannels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

var channelsTmplSrc = `ID	NAME	RELEASE	VERSION
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ if ge .ReleaseSequence 1 }}{{ .ReleaseSequence }}{{else}}	{{end}}	{{ .ReleaseLabel }}
{{ end }}`

var channelsTmpl = template.Must(template.New("channels").Parse(channelsTmplSrc))

func Channels(outputFormat string, w *tabwriter.Writer, channels []*types.Channel) error {
	if outputFormat == "table" {
		if err := channelsTmpl.Execute(w, channels); err != nil {
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

func PlatformChannels(w *tabwriter.Writer, channels []platformChannels.AppChannel) error {
	if err := channelsTmpl.Execute(w, channels); err != nil {
		return err
	}
	return w.Flush()
}

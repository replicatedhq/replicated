package print

import (
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

func Channels(w *tabwriter.Writer, channels []types.Channel) error {
	if err := channelsTmpl.Execute(w, channels); err != nil {
		return err
	}
	return w.Flush()
}

func PlatformChannels(w *tabwriter.Writer, channels []platformChannels.AppChannel) error {
	if err := channelsTmpl.Execute(w, channels); err != nil {
		return err
	}
	return w.Flush()
}

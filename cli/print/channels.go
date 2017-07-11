package print

import (
	"text/tabwriter"
	"text/template"

	channels "github.com/replicatedhq/replicated/gen/go/channels"
)

var channelsTmplSrc = `ID	NAME	RELEASE	VERSION
{{ range . -}}
{{ .Id }}	{{ .Name }}	{{ if ge .ReleaseSequence 1 }}{{ .ReleaseSequence }}{{else}}	{{end}}	{{ .ReleaseLabel }}
{{ end }}`

var channelsTmpl = template.Must(template.New("channels").Parse(channelsTmplSrc))

func Channels(w *tabwriter.Writer, channels []channels.AppChannel) error {
	if err := channelsTmpl.Execute(w, channels); err != nil {
		return err
	}
	return w.Flush()
}

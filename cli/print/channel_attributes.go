package print

import (
	"text/tabwriter"
	"text/template"

	channels "github.com/replicatedhq/replicated/gen/go/v1"
)

var channelAttrsTmplSrc = `ID:	{{ .Id }}
NAME:	{{ .Name }}
DESCRIPTION:	{{ .Description }}
RELEASE:	{{ if ge .ReleaseSequence 1 }}{{ .ReleaseSequence }}{{else}}	{{end}}
VERSION:	{{ .ReleaseLabel }}
`

var channelAttrsTmpl = template.Must(template.New("ChannelAttributes").Parse(channelAttrsTmplSrc))

func ChannelAttrs(w *tabwriter.Writer, appChan *channels.AppChannel) error {
	if err := channelAttrsTmpl.Execute(w, appChan); err != nil {
		return err
	}
	return w.Flush()
}

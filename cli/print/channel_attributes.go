package print

import (
	"github.com/replicatedhq/replicated/pkg/types"
	"text/tabwriter"
	"text/template"
)

var channelAttrsTmplSrc = `ID:	{{ .ID }}
NAME:	{{ .Name }}
DESCRIPTION:	{{ .Description }}
RELEASE:	{{ if ge .ReleaseSequence 1 }}{{ .ReleaseSequence }}{{else}}	{{end}}
VERSION:	{{ .ReleaseLabel }}{{ with .InstallCommands }}
EXISTING:

{{ .Existing }}

EMBEDDED:

{{ .Embedded }}

AIRGAP:

{{ .Airgap }}
{{end}}
`

var channelAttrsTmpl = template.Must(template.New("ChannelAttributes").Parse(channelAttrsTmplSrc))

func ChannelAttrs(w *tabwriter.Writer, appChan *types.Channel) error {
	if err := channelAttrsTmpl.Execute(w, appChan); err != nil {
		return err
	}
	return w.Flush()
}

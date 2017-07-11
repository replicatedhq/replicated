package print

import (
	"fmt"
	"text/tabwriter"
	"text/template"

	channels "github.com/replicatedhq/replicated/gen/go/channels"
)

var channelReleasesTmplSrc = `CHANNEL_SEQUENCE	RELEASE_SEQUENCE	RELEASED	VERSION	REQUIRED	AIRGAP_STATUS	RELEASE_NOTES
{{ range . -}}
{{ .ChannelSequence }}	{{ .ReleaseSequence }}	{{ time .Created }}	{{ .Version }}	{{ .Required }}	{{ .AirgapBuildStatus}}	{{ .ReleaseNotes }}
{{ end }}`

var channelReleasesTmpl = template.Must(template.New("ChannelReleases").Funcs(funcs).Parse(channelReleasesTmplSrc))

func ChannelReleases(w *tabwriter.Writer, releases []channels.ChannelRelease) error {
	if len(releases) == 0 {
		if _, err := fmt.Fprintln(w, "No releases in channel"); err != nil {
			return err
		}
		return w.Flush()
	}

	if err := channelReleasesTmpl.Execute(w, releases); err != nil {
		return err
	}

	return w.Flush()
}

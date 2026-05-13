package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

var channelReleasesTmplSrc = `CHANNEL_SEQUENCE	RELEASE_SEQUENCE	RELEASED	VERSION	REQUIRED	AIRGAP_STATUS	RELEASE_NOTES
{{ range . -}}
{{ .ChannelSequence }}	{{ .ReleaseSequence }}	{{ time .Created }}	{{ .Version }}	{{ .Required }}	{{ .AirgapBuildStatus}}	{{ .ReleaseNotes }}
{{ end }}`

var channelReleasesTmpl = template.Must(template.New("ChannelReleases").Funcs(funcs).Parse(channelReleasesTmplSrc))

func ChannelReleases(outputFormat string, w *tabwriter.Writer, releases []channels.ChannelRelease) error {
	if outputFormat == "json" {
		out, _ := json.MarshalIndent(releases, "", "  ")
		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
		return w.Flush()
	}

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

var kotsChannelReleasesTmplSrc = `CHANNEL_SEQUENCE	RELEASE_SEQUENCE	VERSION	CREATED	RELEASED	STATE	AIRGAP_STATUS	AIRGAP_ERROR
{{ range . -}}
{{ .ChannelSequence }}	{{ .Sequence }}	{{ .Semver }}	{{ time .Created }}	{{ time .ReleasedAt }}	{{ .State }}	{{ .AirgapBuildStatus }}	{{ .AirgapBuildError }}
{{ end }}`

var kotsChannelReleasesTmpl = template.Must(template.New("KotsChannelReleases").Funcs(funcs).Parse(kotsChannelReleasesTmplSrc))

func KotsChannelReleases(outputFormat string, w *tabwriter.Writer, releases []*types.ChannelRelease) error {
	if outputFormat == "json" {
		out, _ := json.MarshalIndent(releases, "", "  ")
		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
		return w.Flush()
	}

	if len(releases) == 0 {
		if _, err := fmt.Fprintln(w, "No releases in channel"); err != nil {
			return err
		}
		return w.Flush()
	}

	rows := make([]map[string]interface{}, len(releases))
	for i, r := range releases {
		state := "active"
		if r.IsDemoted {
			state = "demoted"
		}
		rows[i] = map[string]interface{}{
			"ChannelSequence":   r.ChannelSequence,
			"Sequence":          r.Sequence,
			"Semver":            r.Semver,
			"Created":           r.Created,
			"ReleasedAt":        r.ReleasedAt,
			"State":             state,
			"AirgapBuildStatus": r.AirgapBuildStatus,
			"AirgapBuildError":  r.AirgapBuildError,
		}
	}

	if err := kotsChannelReleasesTmpl.Execute(w, rows); err != nil {
		return err
	}

	return w.Flush()
}

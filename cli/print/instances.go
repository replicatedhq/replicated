package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

var instancesTmplSrc = `ID	NAME	LAST ACTIVE	VERSION	TAGS
{{ range . -}}
{{ .InstanceID }}	{{ .Name }}	{{ .LastActive }}	{{ .LatestVersion }}	{{ .Tags.String }}
{{ end }}`

var instancesTmpl = template.Must(template.New("instances").Parse(instancesTmplSrc))

func Instances(outputFormat string, w *tabwriter.Writer, instances []types.Instance) error {
	if outputFormat == "table" {
		if err := instancesTmpl.Execute(w, instances); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(instances, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

var instanceTmplSrc = `ID	NAME	LAST ACTIVE	VERSION	TAGS
{{ .InstanceID }}	{{ .Name }}	{{ .LastActive }}	{{ .LatestVersion }}	{{ .Tags.String }}

VERSION LABEL	CHANNEL ID	RELEASE SEQUENCE	FIRST SEEN	LAST SEEN
{{ range .VersionHistory -}}
{{ .VersionLabel }}	{{ .DownStreamChannelID }}	{{ .DownStreamReleaseSequence }}	{{ .IntervalStart }}	{{ .IntervalLast }}
{{ end }}`

var instanceTmpl = template.Must(template.New("instances").Parse(instanceTmplSrc))

func Instance(outputFormat string, w *tabwriter.Writer, instance types.Instance) error {
	if outputFormat == "table" {
		if err := instanceTmpl.Execute(w, instance); err != nil {
			return errors.Wrap(err, "execute template")
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(instance, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return errors.Wrap(err, "write json")
		}
	}
	return errors.Wrap(w.Flush(), "flush")
}

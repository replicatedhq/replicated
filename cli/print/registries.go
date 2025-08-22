package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var registriesTmplSrc = `NAME	PROVIDER	ENDPOINT	APP_IDS	AUTHTYPE
{{ range . -}}
{{ .Slug }}	{{ .Provider }}	{{ .Endpoint }}	{{ if .AppIds }}{{ join .AppIds "," }}{{ else }}-{{ end }}	{{ .AuthType }}
{{ end }}`

var registriesTmpl = template.Must(template.New("registries").Funcs(funcs).Parse(registriesTmplSrc))

func Registries(outputFormat string, w *tabwriter.Writer, registries []types.Registry) error {
	if outputFormat == "table" {
		if err := registriesTmpl.Execute(w, registries); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(registries, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}

	return w.Flush()
}

var registryLogsTmplSrc = `DATE	IMAGE	ACTION	STATUS	SUCCESS
{{ range . -}}
{{ .CreatedAt }}	{{ if not .Image }}{{ else }}{{ .Image }}{{ end }} 	{{ .Action }}	{{ if not .Status }}{{ else }}{{ .Status }}{{ end }}	{{ .Success }}
{{ end }}`

var registryLogsTmpl = template.Must(template.New("registryLogs").Funcs(funcs).Parse(registryLogsTmplSrc))

func RegistryLogs(w *tabwriter.Writer, logs []types.RegistryLog) error {
	if err := registryLogsTmpl.Execute(w, logs); err != nil {
		return err
	}

	return w.Flush()
}

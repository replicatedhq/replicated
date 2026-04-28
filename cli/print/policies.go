package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var policiesTmplSrc = `ID	NAME	DESCRIPTION	READ ONLY
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .Description }}	{{ .ReadOnly }}
{{ end }}`

var policiesTmpl = template.Must(template.New("policies").Parse(policiesTmplSrc))

func Policies(outputFormat string, w *tabwriter.Writer, policies []*types.Policy) error {
	if outputFormat == "table" {
		if err := policiesTmpl.Execute(w, policies); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		b, _ := json.MarshalIndent(policies, "", "  ")
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func Policy(outputFormat string, w *tabwriter.Writer, policy *types.Policy) error {
	if outputFormat == "table" {
		if err := policiesTmpl.Execute(w, []*types.Policy{policy}); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		b, _ := json.MarshalIndent(policy, "", "  ")
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	}
	return w.Flush()
}

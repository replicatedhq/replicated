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
	switch outputFormat {
	case "table":
		if err := policiesTmpl.Execute(w, policies); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(policies, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal policies: %w", err)
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func Policy(outputFormat string, w *tabwriter.Writer, policy *types.Policy) error {
	switch outputFormat {
	case "table":
		if err := policiesTmpl.Execute(w, []*types.Policy{policy}); err != nil {
			return err
		}
	case "json":
		b, err := json.MarshalIndent(policy, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal policy: %w", err)
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

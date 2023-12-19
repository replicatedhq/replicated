package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseInstallersTmplSrc = `ID
{{ range . -}}
{{ .ID }}
{{ end }}`

var enterpriseInstallersTmpl = template.Must(template.New("enterpriseinstallers").Parse(enterpriseInstallersTmplSrc))

func EnterpriseInstallers(outputFormat string, w *tabwriter.Writer, installers []*enterprisetypes.Installer) error {
	if outputFormat == "table" {
		if err := enterpriseInstallersTmpl.Execute(w, installers); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(installers, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}

	return w.Flush()
}

package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseInstallerTmplSrc = `ID
{{ .ID }}
`

var enterpriseInstallerTmpl = template.Must(template.New("enterpriseinstaller").Parse(enterpriseInstallerTmplSrc))

func EnterpriseInstaller(outputFormat string, w *tabwriter.Writer, installer *enterprisetypes.Installer) error {
	if outputFormat == "table" {
		if err := enterpriseInstallerTmpl.Execute(w, installer); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(installer, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}

	return w.Flush()
}

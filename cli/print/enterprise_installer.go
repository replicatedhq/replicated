package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseInstallerTmplSrc = `ID
{{ .ID }}
`

var enterpriseInstallerTmpl = template.Must(template.New("enterpriseinstaller").Parse(enterpriseInstallerTmplSrc))

func EnterpriseInstaller(w *tabwriter.Writer, installer *enterprisetypes.Installer) error {
	if err := enterpriseInstallerTmpl.Execute(w, installer); err != nil {
		return err
	}
	return w.Flush()
}

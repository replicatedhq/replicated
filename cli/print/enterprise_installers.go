package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

var enterpriseInstallersTmplSrc = `ID
{{ range . -}}
{{ .ID }}
{{ end }}`

var enterpriseInstallersTmpl = template.Must(template.New("enterpriseinstallers").Parse(enterpriseInstallersTmplSrc))

func EnterpriseInstallers(w *tabwriter.Writer, installers []*enterprisetypes.Installer) error {
	if err := enterpriseInstallersTmpl.Execute(w, installers); err != nil {
		return err
	}
	return w.Flush()
}

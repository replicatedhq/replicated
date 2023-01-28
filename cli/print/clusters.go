package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var clustersTmplSrc = `ID	NAME	DISTRO	VERSION	STATUS	CREATED	EXPIRES
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .Distribution }}	{{ .Version }}	{{ .Status }}	{{ .CreatedAt}}	{{ .ExpiresAt }}
{{ end }}`

var clustersTmpl = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersTmplSrc))

func Clusters(w *tabwriter.Writer, clusters []*types.Cluster) error {
	if err := clustersTmpl.Execute(w, clusters); err != nil {
		return err
	}

	return w.Flush()
}

package print

import (
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

// TODO: implement a -o wide, and expose nodecount, vcpus and memory also?
var clustersTmplSrc = `ID	NAME	K8S DISTRO	K8S VERSION	STATUS	CREATED	EXPIRES
{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .KubernetesDistribution}}	{{ .KubernetesVersion	}}	{{ .Status }}	{{ .CreatedAt}}	{{ .ExpiresAt }}
{{ end }}`

var clustersTmpl = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersTmplSrc))

func Clusters(w *tabwriter.Writer, clusters []*types.Cluster) error {
	if err := clustersTmpl.Execute(w, clusters); err != nil {
		return err
	}

	return w.Flush()
}

func NoClusters(w *tabwriter.Writer) error {
	_, err := w.Write([]byte(`No clusters found. Use the "replicated cluster create" command to create a new cluster.`))
	if err != nil {
		return err
	}

	return w.Flush()
}

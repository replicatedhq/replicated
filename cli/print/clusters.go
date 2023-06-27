package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

// TODO: implement a -o wide, and expose nodecount, vcpus and memory also?
var clustersSourceTmpHeader = `ID	NAME	K8S DISTRO	K8S VERSION	STATUS	CREATED	EXPIRES
`

var clustersTmplSrc = `{{ range . -}}
{{ .ID }}	{{ .Name }}	{{ .KubernetesDistribution}}	{{ .KubernetesVersion	}}	{{ .Status }}	{{ .CreatedAt}}	{{ .ExpiresAt }}
{{ end }}`

var clusterTmpNoHeader = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersTmplSrc))
var clustersTmpl = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersSourceTmpHeader + clustersTmplSrc))

func Clusters(outputFormat string, w *tabwriter.Writer, clusters []*types.Cluster, includeHeader bool) error {
	tmpl := clustersTmpl

	if outputFormat == "table" {
		if err := tmpl.Execute(w, clusters); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(clusters, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func NoClusters(outputFormat string, w *tabwriter.Writer) error {
	if outputFormat == "table" {
		_, err := fmt.Fprintln(w, "No clusters found. Use the `replicated cluster create` command to create a new cluster.")
		if err != nil {
			return err
		}
	} else if outputFormat == "json" {
		if _, err := fmt.Fprintln(w, "[]"); err != nil {
			return err
		}
	}
	return w.Flush()
}

func Cluster(outputFormat string, w *tabwriter.Writer, cluster *types.Cluster, includerHeader bool) error {
	return Clusters(outputFormat, w, []*types.Cluster{cluster}, false)
}

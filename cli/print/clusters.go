package print

import (
	"encoding/json"
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

func Clusters(outputFormat string, w *tabwriter.Writer, clusters []*types.Cluster) error {
	if outputFormat == "table" {
		if err := clustersTmpl.Execute(w, clusters); err != nil {
			return err
		}
		return w.Flush()

	}
	if outputFormat == "json" {
		defer w.Flush()
		var cAsByte []byte
		cAsByte, _ = json.MarshalIndent(clusters, "", "  ")
		_, err := w.Write(cAsByte)
		return err
	}
	return nil
}

func NoClusters(outputFormat string, w *tabwriter.Writer) error {
	if outputFormat == "table" {
		_, err := w.Write([]byte(`No clusters found. Use the "replicated cluster create" command to create a new cluster.`))
		if err != nil {
			return err
		}

		return w.Flush()
	} else if outputFormat == "json" {
		defer w.Flush()
		_, err := w.Write([]byte("[]"))
		return err
	}
	return nil
}

func Cluster(outputFormat string, w *tabwriter.Writer, cluster *types.Cluster) error {
	if outputFormat == "table" {
		if err := clustersTmpl.Execute(w, []types.Cluster{*cluster}); err != nil {
			return err
		}
		return w.Flush()

	}
	if outputFormat == "json" {
		defer w.Flush()
		var cAsByte []byte
		cAsByte, _ = json.MarshalIndent(cluster, "", "  ")
		_, err := w.Write(cAsByte)
		return err
	}
	return nil
}

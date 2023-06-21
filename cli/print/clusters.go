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

var clusterVersionsTmplSrc = `K8S DISTRO	K8S VERSION
{{ range $d := . -}}{{ range $v := $d.Versions -}}
{{ $d.Name }}	{{ $v }} 
{{ end }}{{ end }}`

var clustersTmpl = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersTmplSrc))
var clusterVersionsTmpl = template.Must(template.New("clusterVersions").Funcs(funcs).Parse(clusterVersionsTmplSrc))

func Clusters(outputFormat string, w *tabwriter.Writer, clusters []*types.Cluster) error {
	if outputFormat == "table" {
		if err := clustersTmpl.Execute(w, clusters); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(clusters, "", "  ")
		if _, err := w.Write(cAsByte); err != nil {
			return err
		}
	}
	return w.Flush()
}

func NoClusters(outputFormat string, w *tabwriter.Writer) error {
	if outputFormat == "table" {
		_, err := w.Write([]byte(`No clusters found. Use the "replicated cluster create" command to create a new cluster.`))
		if err != nil {
			return err
		}
	} else if outputFormat == "json" {
		if _, err := w.Write([]byte("[]")); err != nil {
			return err
		}
	}
	return w.Flush()
}

func Cluster(outputFormat string, w *tabwriter.Writer, cluster *types.Cluster) error {
	if outputFormat == "table" {
		if err := clustersTmpl.Execute(w, []types.Cluster{*cluster}); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(cluster, "", "  ")
		if _, err := w.Write(cAsByte); err != nil {
			return err
		}
	}
	return w.Flush()
}

func NoClusterVersions(outputFormat string, w *tabwriter.Writer) error {
	if outputFormat == "table" {
		_, err := w.Write([]byte("No cluster versions found.\n"))
		if err != nil {
			return err
		}
	} else if outputFormat == "json" {
		if _, err := w.Write([]byte("[]\n")); err != nil {
			return err
		}
	}
	return w.Flush()
}

func ClusterVersions(outputFormat string, w *tabwriter.Writer, clusters []*types.ClusterVersion) error {
	if outputFormat == "table" {
		if err := clusterVersionsTmpl.Execute(w, clusters); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(clusters, "", "  ")
		if _, err := w.Write(append(cAsByte, "\n"...)); err != nil {
			return err
		}
	}
	return w.Flush()
}

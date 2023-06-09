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

var supportedClustersTmplSrc = `K8S DISTRO	K8S VERSION
{{ range $d := . -}}{{ range $v := $d.Versions -}}
{{ $d.Name }}	{{ $v }} 
{{ end }}{{ end }}`

var clustersTmpl = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersTmplSrc))
var supportedClustersTmpl = template.Must(template.New("supportedClusters").Funcs(funcs).Parse(supportedClustersTmplSrc))

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

func NoSupportedClusters(outputFormat string, w *tabwriter.Writer) error {
	if outputFormat == "table" {
		_, err := w.Write([]byte("No supported clusters found.\n"))
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

func SupportedClusters(outputFormat string, w *tabwriter.Writer, clusters []*types.SupportedCluster) error {
	if outputFormat == "table" {
		if err := supportedClustersTmpl.Execute(w, clusters); err != nil {
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

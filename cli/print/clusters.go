package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

// TODO: implement a -o wide, and expose nodecount also?
var clustersTmplHeaderSrc = `ID	NAME	DISTRIBUTION	VERSION	NODE GROUPS	STATUS	CREATED	EXPIRES	TAGS`
var clustersTmplRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding .KubernetesDistribution 12 }}	{{ padding .KubernetesVersion 10 }}	{{ add (len .AdditionalNodeGroups) 1 }}	{{ padding (printf "%s" .Status) 12 }}	{{ .CreatedAt}}	{{if .ExpiresAt.IsZero}}-{{else}}{{ .ExpiresAt }}{{end}}	{{ range $index, $tag := .Tags }}{{if $index}}, {{end}}{{ $tag.Key }}={{ $tag.Value }}{{ end }}
{{ end }}`
var clustersTmplSrc = fmt.Sprintln(clustersTmplHeaderSrc) + clustersTmplRowSrc

var clusterVersionsTmplSrc = `Supported Kubernetes distributions and versions are:
{{ range $d := . -}}
DISTRIBUTION: {{ $d.Name }}
• VERSIONS: {{ range $i, $v := $d.Versions -}}{{if $i}}, {{end}}{{ $v }}{{ end }}
• INSTANCE TYPES: {{ range $i, $it := $d.InstanceTypes -}}{{if $i}}, {{end}}{{ $it }}{{ end }}
• MAX NODES: {{ $d.NodesMax }}{{if $d.Status}}
• ENABLED: {{ $d.Status.Enabled }}
• STATUS: {{ $d.Status.Status }}
• DETAILS: {{ $d.Status.StatusMessage }}{{end}}

{{ end }}`

var clustersTmpl = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersTmplSrc))
var clustersTmplNoHeader = template.Must(template.New("clusters").Funcs(funcs).Parse(clustersTmplRowSrc))
var clusterVersionsTmpl = template.Must(template.New("clusterVersions").Funcs(funcs).Parse(clusterVersionsTmplSrc))

var clusterNodeGroupsTmplHeaderSrc = `ID	NAME	NODE COUNT	DISK	INSTANCE TYPE`
var clusterNodeGroupTmplRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27 }}	{{ .NodeCount }}	{{ .DiskGiB }} GiB	{{ .InstanceType }}
{{ end }}`
var clusterNodeGroupsTmplSrc = fmt.Sprintln(clusterNodeGroupsTmplHeaderSrc) + clusterNodeGroupTmplRowSrc

var clusterNodeGroupsTmpl = template.Must(template.New("clusterNodeGroups").Funcs(funcs).Parse(clusterNodeGroupsTmplSrc))
var clusterNodeGroupsTmplNoHeader = template.Must(template.New("clusterNodeGroups").Funcs(funcs).Parse(clusterNodeGroupTmplRowSrc))

func NoNodeGroups(outputFormat string, w *tabwriter.Writer) error {
	if outputFormat == "table" {
		_, err := fmt.Fprintln(w, "No additional node groups found. Use the `replicated cluster nodegroup create` command to create an additional node group.")
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

func ClusterNodeGroups(outputFormat string, w *tabwriter.Writer, nodeGroups []*types.ClusterNodeGroup, header bool) error {
	if outputFormat == "table" {
		if header {
			if err := clusterNodeGroupsTmpl.Execute(w, nodeGroups); err != nil {
				return err
			}
		} else {
			if err := clusterNodeGroupsTmplNoHeader.Execute(w, nodeGroups); err != nil {
				return err
			}
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(nodeGroups, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func ClusterNodeGroup(outputFormat string, w *tabwriter.Writer, ng *types.ClusterNodeGroup) error {
	return nil
}

func Clusters(outputFormat string, w *tabwriter.Writer, clusters []*types.Cluster, header bool) error {
	if outputFormat == "table" {
		if header {
			if err := clustersTmpl.Execute(w, clusters); err != nil {
				return err
			}
		} else {
			if err := clustersTmplNoHeader.Execute(w, clusters); err != nil {
				return err
			}
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

func Cluster(outputFormat string, w *tabwriter.Writer, cluster *types.Cluster) error {
	if outputFormat == "table" {
		if err := clustersTmpl.Execute(w, []*types.Cluster{cluster}); err != nil {
			return err
		}
	} else if outputFormat == "json" {
		cAsByte, _ := json.MarshalIndent(cluster, "", "  ")
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func NoClusterVersions(outputFormat string, w *tabwriter.Writer) error {
	if outputFormat == "table" {
		_, err := fmt.Fprintln(w, "No cluster versions found.")
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

func ClusterVersions(outputFormat string, w *tabwriter.Writer, clusters []*types.ClusterVersion) error {
	if outputFormat == "table" {
		if err := clusterVersionsTmpl.Execute(w, clusters); err != nil {
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

func ClusterInstanceTypes(outpuFormat string, w *tabwriter.Writer, instanceTypes []string) error {
	return nil
}

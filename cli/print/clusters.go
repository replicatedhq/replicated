package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

var clusterFuncs = template.FuncMap{
	"CreditsToDollarsDisplay": CreditsToDollarsDisplay,
}

// Table formatting
var clustersTmplTableHeaderSrc = `ID	NAME	DISTRIBUTION	VERSION	STATUS	CREATED	EXPIRES	COST`
var clustersTmplTableRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding .KubernetesDistribution 12 }}	{{ padding .KubernetesVersion 10 }}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" .CreatedAt) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" .ExpiresAt) 30 }}{{end}}	{{ padding (CreditsToDollarsDisplay .EstimatedCost) 11 }}
{{ end }}`
var clustersTmplTableSrc = fmt.Sprintln(clustersTmplTableHeaderSrc) + clustersTmplTableRowSrc
var clustersTmplTable = template.Must(template.New("clusters").Funcs(clusterFuncs).Funcs(funcs).Parse(clustersTmplTableSrc))
var clustersTmplTableNoHeader = template.Must(template.New("clusters").Funcs(clusterFuncs).Funcs(funcs).Parse(clustersTmplTableRowSrc))

// Wide table formatting
var clustersTmplWideHeaderSrc = `ID	NAME	DISTRIBUTION	VERSION	STATUS	CREATED	EXPIRES	COST	TOTAL NODES	NODEGROUPS	TAGS`
var clustersTmplWideRowSrc = `{{ range . -}}
{{ .ID }}	{{ padding .Name 27	}}	{{ padding .KubernetesDistribution 12 }}	{{ padding .KubernetesVersion 10 }}	{{ padding (printf "%s" .Status) 12 }}	{{ padding (printf "%s" .CreatedAt) 30 }}	{{if .ExpiresAt.IsZero}}{{ padding "-" 30 }}{{else}}{{ padding (printf "%s" .ExpiresAt) 30 }}{{end}}	{{ padding (CreditsToDollarsDisplay .EstimatedCost) 11 }}	{{$nodecount:=0}}{{ range $index, $ng := .NodeGroups}}{{$nodecount = add $nodecount $ng.NodeCount}}{{ end }}{{ padding (printf "%d" $nodecount) 11 }}	{{ len .NodeGroups}}	{{ range $index, $tag := .Tags }}{{if $index}}, {{end}}{{ $tag.Key }}={{ $tag.Value }}{{ end }}
{{ end }}`
var clustersTmplWideSrc = fmt.Sprintln(clustersTmplWideHeaderSrc) + clustersTmplWideRowSrc
var clustersTmplWide = template.Must(template.New("clusters").Funcs(clusterFuncs).Funcs(funcs).Parse(clustersTmplWideSrc))
var clustersTmplWideNoHeader = template.Must(template.New("clusters").Funcs(clusterFuncs).Funcs(funcs).Parse(clustersTmplWideRowSrc))

// Cluster versions
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
var clusterVersionsTmpl = template.Must(template.New("clusterVersions").Funcs(funcs).Parse(clusterVersionsTmplSrc))

func CreditsToDollarsDisplay(credits int64) string {
	dollars := float64(credits) / 100000.0
	return fmt.Sprintf("$%.2f", dollars)
}

func Clusters(outputFormat string, w *tabwriter.Writer, clusters []*types.Cluster, header bool) error {
	for _, cluster := range clusters {
		updateEstimatedCost(cluster)
	}
	switch outputFormat {
	case "table":
		if header {
			if err := clustersTmplTable.Execute(w, clusters); err != nil {
				return err
			}
		} else {
			if err := clustersTmplTableNoHeader.Execute(w, clusters); err != nil {
				return err
			}
		}
	case "wide":
		if header {
			if err := clustersTmplWide.Execute(w, clusters); err != nil {
				return err
			}
		} else {
			if err := clustersTmplWideNoHeader.Execute(w, clusters); err != nil {
				return err
			}
		}
	case "json":
		cAsByte, err := json.MarshalIndent(clusters, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func NoClusters(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table", "wide":
		_, err := fmt.Fprintln(w, "No clusters found. Use the `replicated cluster create` command to create a new cluster.")
		if err != nil {
			return err
		}
	case "json":
		if _, err := fmt.Fprintln(w, "[]"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func Cluster(outputFormat string, w *tabwriter.Writer, cluster *types.Cluster) error {
	updateEstimatedCost(cluster)
	switch outputFormat {
	case "table":
		if err := clustersTmplTable.Execute(w, []*types.Cluster{cluster}); err != nil {
			return err
		}
	case "wide":
		if err := clustersTmplWide.Execute(w, []*types.Cluster{cluster}); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(cluster, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func updateEstimatedCost(cluster *types.Cluster) {
	if cluster.EstimatedCost != 0 {
		return
	}
	if cluster.TotalCredits > 0 {
		cluster.EstimatedCost = cluster.TotalCredits
	} else {
		expireDuration, _ := time.ParseDuration(cluster.TTL)
		minutesRunning := int64(expireDuration.Minutes())
		totalCredits := int64(minutesRunning) * cluster.CreditsPerHourPerCluster / 60.0
		cluster.EstimatedCost = cluster.FlatFee + totalCredits
		for _, ng := range cluster.NodeGroups {
			cluster.EstimatedCost += int64(minutesRunning) * ng.CreditsPerHour / 60.0 * int64(ng.NodeCount)
		}
	}
}

func NoClusterVersions(outputFormat string, w *tabwriter.Writer) error {
	switch outputFormat {
	case "table":
		_, err := fmt.Fprintln(w, "No cluster versions found.")
		if err != nil {
			return err
		}
	case "json":
		if _, err := fmt.Fprintln(w, "[]"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

func ClusterVersions(outputFormat string, w *tabwriter.Writer, clusters []*types.ClusterVersion) error {
	switch outputFormat {
	case "table":
		if err := clusterVersionsTmpl.Execute(w, clusters); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(clusters, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return w.Flush()
}

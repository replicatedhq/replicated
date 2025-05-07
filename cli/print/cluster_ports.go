package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var portsTmplHeaderSrc = `ID	CLUSTER PORT	PROTOCOL	EXPOSED PORT	WILDCARD	STATUS`
var portsTmplRowSrc = `{{- range . }}
{{- $id := .AddonID }}
{{- $upstreamPort := .UpstreamPort }}
{{- $hostname := .Hostname }}
{{- $isWildcard := .IsWildcard }}
{{- $state := .State }}
{{- range .ExposedPorts }}
{{ $id }}	{{ $upstreamPort }}	{{ .Protocol }}	{{ formatURL .Protocol $hostname }}	{{ $isWildcard }}	{{ printf "%-12s" $state }}
{{ end }}
{{ end }}`
var portsTmplSrc = fmt.Sprintln(portsTmplHeaderSrc) + portsTmplRowSrc
var portsTmpl = template.Must(template.New("ports").Funcs(funcs).Parse(portsTmplSrc))
var portsTmplNoHeader = template.Must(template.New("ports").Funcs(funcs).Parse(portsTmplRowSrc))

const (
	clusterPortsMinWidth = 16
	clusterPortsTabWidth = 8
	clusterPortsPadding  = 4
	clusterPortsPadChar  = ' '
)

func ClusterPorts(outputFormat string, w *tabwriter.Writer, ports []*types.ClusterPort, header bool) error {
	// we need to configure our writer with custom settings because our column widths are large
	w.Init(w, clusterPortsMinWidth, clusterPortsTabWidth, clusterPortsPadding, clusterPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table", "wide":
		if header {
			if err := portsTmpl.Execute(w, ports); err != nil {
				return err
			}
		} else {
			if err := portsTmplNoHeader.Execute(w, ports); err != nil {
				return err
			}
		}
	case "json":
		cAsByte, err := json.MarshalIndent(ports, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return w.Flush()
}

func ClusterPort(outputFormat string, w *tabwriter.Writer, port *types.ClusterPort) error {
	// we need to configure our writer with custom settings because our column widths are large
	w.Init(w, clusterPortsMinWidth, clusterPortsTabWidth, clusterPortsPadding, clusterPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table":
		if err := portsTmpl.Execute(w, []*types.ClusterPort{port}); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(port, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return w.Flush()
}

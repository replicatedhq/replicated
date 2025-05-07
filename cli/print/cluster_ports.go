package print

import (
	"encoding/json"
	"fmt"
	"os"
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
{{ $id }}	{{ $upstreamPort }}	{{ .Protocol }}	{{ formatURL .Protocol $hostname }}	{{ $isWildcard }}	{{ $state }}
{{- end }}
{{- end }}`
var portsTmplSrc = fmt.Sprint(portsTmplHeaderSrc) + portsTmplRowSrc
var portsTmpl = template.Must(template.New("ports").Funcs(funcs).Parse(portsTmplSrc))
var portsTmplNoHeader = template.Must(template.New("ports").Funcs(funcs).Parse(portsTmplRowSrc))

const (
	clusterPortsMinWidth = 16
	clusterPortsTabWidth = 8
	clusterPortsPadding  = 4
	clusterPortsPadChar  = ' '
)

func ClusterPorts(outputFormat string, _ *tabwriter.Writer, ports []*types.ClusterPort, header bool) error {
	// we need a custom tab writer here because our column widths are large
	portsWriter := tabwriter.NewWriter(os.Stdout, clusterPortsMinWidth, clusterPortsTabWidth, clusterPortsPadding, clusterPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table", "wide":
		if header {
			if err := portsTmpl.Execute(portsWriter, ports); err != nil {
				return err
			}
		} else {
			if err := portsTmplNoHeader.Execute(portsWriter, ports); err != nil {
				return err
			}
		}
	case "json":
		cAsByte, err := json.MarshalIndent(ports, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(portsWriter, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return portsWriter.Flush()
}

func ClusterPort(outputFormat string, _ *tabwriter.Writer, port *types.ClusterPort) error {
	// we need a custom tab writer here because our column widths are large
	portsWriter := tabwriter.NewWriter(os.Stdout, clusterPortsMinWidth, clusterPortsTabWidth, clusterPortsPadding, clusterPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table":
		if err := portsTmpl.Execute(portsWriter, []*types.ClusterPort{port}); err != nil {
			return err
		}
	case "json":
		cAsByte, err := json.MarshalIndent(port, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(portsWriter, string(cAsByte)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return portsWriter.Flush()
}

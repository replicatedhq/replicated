package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var vmPortsTmplHeaderSrc = `ID	VM PORT	PROTOCOL	EXPOSED PORT	STATUS`
var vmPortsTmplRowSrc = `{{- range . }}
{{- $id := .AddonID }}
{{- $upstreamPort := .UpstreamPort }}
{{- $hostname := .Hostname }}
{{- $state := .State }}
{{- range .ExposedPorts }}
{{ $id }}	{{ $upstreamPort }}	{{ .Protocol }}	{{ formatURL .Protocol $hostname }}	{{ printf "%-12s" $state }}
{{ end }}
{{ end }}`
var vmPortsTmplSrc = fmt.Sprintln(vmPortsTmplHeaderSrc) + vmPortsTmplRowSrc
var vmPortsTmpl = template.Must(template.New("ports").Funcs(funcs).Parse(vmPortsTmplSrc))
var vmPortsTmplNoHeader = template.Must(template.New("ports").Funcs(funcs).Parse(vmPortsTmplRowSrc))

const (
	vmPortsMinWidth = 16
	vmPortsTabWidth = 8
	vmPortsPadding  = 4
	vmPortsPadChar  = ' '
)

func VMPorts(outputFormat string, w *tabwriter.Writer, ports []*types.VMPort, header bool) error {
	// we need to configure our writer with custom settings because our column widths are large
	w.Init(w, vmPortsMinWidth, vmPortsTabWidth, vmPortsPadding, vmPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table", "wide":
		if header {
			if err := vmPortsTmpl.Execute(w, ports); err != nil {
				return err
			}
		} else {
			if err := vmPortsTmplNoHeader.Execute(w, ports); err != nil {
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

func VMPort(outputFormat string, w *tabwriter.Writer, port *types.VMPort) error {
	// we need to configure our writer with custom settings because our column widths are large
	w.Init(w, vmPortsMinWidth, vmPortsTabWidth, vmPortsPadding, vmPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table":
		if err := vmPortsTmpl.Execute(w, []*types.VMPort{port}); err != nil {
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

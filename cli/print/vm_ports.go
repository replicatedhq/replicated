package print

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/replicatedhq/replicated/pkg/types"
)

var vmPortsTmplHeaderSrc = `ID	VM PORT	PROTOCOL	EXPOSED PORT	STATUS`
var vmPortsTmplRowSrc = `{{- range . -}}
{{- $id := .AddonID -}}
{{- $upstreamPort := .UpstreamPort -}}
{{- $hostname := .Hostname -}}
{{- $state := .State -}}
{{- range .ExposedPorts -}}
{{ $id }}	{{ $upstreamPort }}	{{ .Protocol }}	{{ formatURL .Protocol $hostname }}	{{ printf "%-12s" $state }}
{{- end -}}
{{- end -}}`
var vmPortsTmplSrc = fmt.Sprint(vmPortsTmplHeaderSrc) + vmPortsTmplRowSrc
var vmPortsTmpl = template.Must(template.New("ports").Funcs(funcs).Parse(vmPortsTmplSrc))
var vmPortsTmplNoHeader = template.Must(template.New("ports").Funcs(funcs).Parse(vmPortsTmplRowSrc))

const (
	vmPortsMinWidth = 16
	vmPortsTabWidth = 8
	vmPortsPadding  = 4
	vmPortsPadChar  = ' '
)

func VMPorts(outputFormat string, _ *tabwriter.Writer, ports []*types.VMPort, header bool) error {
	// we need a custom tab writer here because our column widths are large
	portsWriter := tabwriter.NewWriter(os.Stdout, vmPortsMinWidth, vmPortsTabWidth, vmPortsPadding, vmPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table", "wide":
		if header {
			if err := vmPortsTmpl.Execute(portsWriter, ports); err != nil {
				return err
			}
		} else {
			if err := vmPortsTmplNoHeader.Execute(portsWriter, ports); err != nil {
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

func VMPort(outputFormat string, _ *tabwriter.Writer, port *types.VMPort) error {
	// we need a custom tab writer here because our column widths are large
	portsWriter := tabwriter.NewWriter(os.Stdout, vmPortsMinWidth, vmPortsTabWidth, vmPortsPadding, vmPortsPadChar, tabwriter.TabIndent)

	switch outputFormat {
	case "table":
		if err := vmPortsTmpl.Execute(portsWriter, []*types.VMPort{port}); err != nil {
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
